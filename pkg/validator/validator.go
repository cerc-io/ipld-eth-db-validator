// VulcanizeDB
// Copyright © 2022 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	ipfsethdb "github.com/cerc-io/ipfs-ethdb/v5/postgres/v0"
	ipldeth "github.com/cerc-io/ipld-eth-server/v5/pkg/eth"
	"github.com/cerc-io/ipld-eth-server/v5/pkg/shared"
	ipldstate "github.com/cerc-io/ipld-eth-statedb/trie_by_cid/state"

	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/prom"
)

var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

type Service struct {
	db *sqlx.DB

	chainConfig           *params.ChainConfig
	ethClient             *rpc.Client
	blockNum, trail       uint64
	retryInterval         time.Duration
	stateDiffMissingBlock bool
	stateDiffTimeout      time.Duration

	quitChan     chan bool
	progressChan chan<- uint64
}

func NewService(cfg *Config, progressChan chan<- uint64) (*Service, error) {
	db, err := postgres.ConnectSQLX(context.Background(), cfg.DBConfig)
	if err != nil {
		return nil, err
	}
	// Enable DB stats
	if cfg.DBStats {
		prom.RegisterDBCollector(cfg.DBConfig.DatabaseName, db)
	}

	return &Service{
		db:                    db,
		chainConfig:           cfg.ChainConfig,
		ethClient:             cfg.Client,
		blockNum:              cfg.FromBlock,
		trail:                 cfg.Trail,
		retryInterval:         cfg.RetryInterval,
		stateDiffMissingBlock: cfg.StateDiffMissingBlock,
		stateDiffTimeout:      cfg.StateDiffTimeout,
		quitChan:              make(chan bool),
		progressChan:          progressChan,
	}, nil
}

// Start is used to begin the service
func (s *Service) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	api, err := EthAPI(ctx, s.db, s.chainConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	nextBlockNum := s.blockNum
	var delay time.Duration
	for {
		log.Debug("delaying %s", delay)
		select {
		case <-s.quitChan:
			log.Info("stopping ipld-eth-db-validator process")
			if s.progressChan != nil {
				close(s.progressChan)
			}
			if err := api.B.Close(); err != nil {
				log.Errorf("error closing backend: %s", err)
			}
			return
		case <-time.After(delay):
			err := s.Validate(ctx, api, nextBlockNum)
			// If chain is not synced, wait for trail to catch up before trying again
			if err, ok := err.(*ChainNotSyncedError); ok {
				log.Debugf("waiting for chain to advance, head is at block %d", err.Head)
				delay = s.retryInterval
				continue
			}
			if err != nil {
				log.Fatal(err)
				return
			}
			prom.SetLastValidatedBlock(float64(nextBlockNum))
			nextBlockNum++
		}
	}
}

// Stop is used to gracefully stop the service
func (s *Service) Stop() {
	close(s.quitChan)
}

func (s *Service) Validate(ctx context.Context, api *ipldeth.PublicEthAPI, idxBlockNum uint64) error {
	log.Debugf("validating block %d", idxBlockNum)
	headBlockNum, err := fetchHeadBlockNumber(ctx, api)
	if err != nil {
		return err
	}

	// Check if block at requested height can be validated
	if idxBlockNum > headBlockNum-s.trail {
		return &ChainNotSyncedError{headBlockNum}
	}

	blockToBeValidated, err := api.B.BlockByNumber(ctx, rpc.BlockNumber(idxBlockNum))
	if err != nil {
		log.Errorf("failed to fetch block at height %d", idxBlockNum)
		return err
	}

	// Make a writeStateDiffAt call if block not found in the db
	if blockToBeValidated == nil {
		return s.writeStateDiffAt(idxBlockNum)
	}

	err = ValidateBlock(blockToBeValidated, api.B, idxBlockNum)
	if err != nil {
		log.Errorf("failed to verify state root at block %d", idxBlockNum)
		return err
	}
	log.Infof("state root verified for block %d", idxBlockNum)

	tx := s.db.MustBegin()
	defer tx.Rollback()
	err = ValidateReferentialIntegrity(tx, idxBlockNum)
	if err != nil {
		log.Errorf("failed to verify referential integrity at block %d", idxBlockNum)
		return err
	}
	log.Infof("referential integrity verified for block %d", idxBlockNum)

	if s.progressChan != nil {
		s.progressChan <- idxBlockNum
	}
	return nil
}

// ValidateBlock validates block at the given height
func ValidateBlock(blockToBeValidated *types.Block, b *ipldeth.Backend, blockNumber uint64) error {
	state, err := applyTransactions(blockToBeValidated, b)
	if err != nil {
		return err
	}

	blockStateRoot := blockToBeValidated.Header().Root
	dbStateRoot := state.IntermediateRoot(true)
	if blockStateRoot != dbStateRoot {
		return fmt.Errorf("state roots do not match at block %d", blockNumber)
	}
	return nil
}

func EthAPI(ctx context.Context, db *sqlx.DB, chainCfg *params.ChainConfig) (*ipldeth.PublicEthAPI, error) {
	// TODO: decide network for custom chainConfig.
	backend, err := ethBackend(db, &ipldeth.Config{
		ChainConfig: chainCfg,
		GroupCacheConfig: &shared.GroupCacheConfig{
			StateDB: shared.GroupConfig{
				Name: "cerc_validator",
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var genesisBlock *types.Block
	if backend.Config.ChainConfig == nil {
		genesisBlock, err = backend.BlockByNumber(ctx, rpc.BlockNumber(0))
		if err != nil {
			return nil, err
		}
		backend.Config.ChainConfig = setChainConfig(genesisBlock.Hash())
	}
	config := ipldeth.APIConfig{
		SupportsStateDiff:   false,
		ForwardEthCalls:     false,
		ForwardGetStorageAt: false,
		ProxyOnError:        false,
		StateDiffTimeout:    0,
	}
	return ipldeth.NewPublicEthAPI(backend, nil, config)
}

func ethBackend(db *sqlx.DB, c *ipldeth.Config) (*ipldeth.Backend, error) {
	gcc := c.GroupCacheConfig

	groupName := gcc.StateDB.Name
	if groupName == "" {
		groupName = ipldeth.StateDBGroupCacheName
	}

	r := ipldeth.NewRetriever(db)
	ethDB := ipfsethdb.NewDatabase(db, ipfsethdb.CacheConfig{
		Name:           groupName,
		Size:           gcc.StateDB.CacheSizeInMB * 1024 * 1024,
		ExpiryDuration: time.Minute * time.Duration(gcc.StateDB.CacheExpiryInMins),
	})
	// Read only wrapper around ipfs-ethdb eth.Database implementation
	customEthDB := newDatabase(ethDB)

	return &ipldeth.Backend{
		DB:                    db,
		Retriever:             r,
		EthDB:                 customEthDB,
		IpldTrieStateDatabase: ipldstate.NewDatabase(customEthDB),
		Config:                c,
	}, nil
}

// fetchHeadBlockNumber gets the latest block number from the db
func fetchHeadBlockNumber(ctx context.Context, api *ipldeth.PublicEthAPI) (uint64, error) {
	headBlock, err := api.B.BlockByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		return 0, err
	}

	return headBlock.NumberU64(), nil
}

// writeStateDiffAt calls out to a statediffing geth client to fill in a gap in the index
func (s *Service) writeStateDiffAt(height uint64) error {
	if !s.stateDiffMissingBlock {
		return nil
	}

	var data json.RawMessage
	params := statediff.Params{
		IncludeBlock:    true,
		IncludeReceipts: true,
		IncludeTD:       true,
		IncludeCode:     true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.stateDiffTimeout)
	defer cancel()

	log.Warnf("calling writeStateDiffAt at block %d", height)
	if err := s.ethClient.CallContext(ctx, &data, "statediff_writeStateDiffAt", height, params); err != nil {
		log.Errorf("writeStateDiffAt %d failed with err %s", height, err)
		return err
	}
	return nil
}

// applyTransaction attempts to apply block transactions to the given state database
// and uses the input parameters for its environment. It returns the stateDB of parent with applied txs.
func applyTransactions(block *types.Block, backend *ipldeth.Backend) (*ipldstate.StateDB, error) {
	if block.NumberU64() == 0 {
		return nil, errors.New("no transaction in genesis")
	}

	// Create the parent state database
	parentHash := block.ParentHash()
	nrOrHash := rpc.BlockNumberOrHash{BlockHash: &parentHash}
	statedb, _, err := backend.IPLDTrieStateDBAndHeaderByNumberOrHash(context.Background(), nrOrHash)
	if err != nil {
		return nil, err
	}

	var gp core.GasPool
	gp.AddGas(block.GasLimit())

	signer := types.MakeSigner(backend.Config.ChainConfig, block.Number())
	blockContext := core.NewEVMBlockContext(block.Header(), backend, getAuthor(backend, block.Header()))
	evm := vm.NewEVM(blockContext, vm.TxContext{}, statedb, backend.Config.ChainConfig, vm.Config{})
	rules := backend.Config.ChainConfig.Rules(block.Number(), true, block.Time())

	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		msg, err := core.TransactionToMessage(tx, signer, block.BaseFee())
		if err != nil {
			return nil, err
		}
		statedb.SetTxContext(tx.Hash(), i)
		statedb.Prepare(rules, msg.From, block.Coinbase(), msg.To, nil, nil)

		// Create a new context to be used in the EVM environment.
		evm.Reset(core.NewEVMTxContext(msg), statedb)
		// Apply the transaction to the current state (included in the env).
		if _, err := core.ApplyMessage(evm, msg, &gp); err != nil {
			return nil, fmt.Errorf("transaction %#x failed: %w", tx.Hash(), err)
		}
	}

	if backend.Config.ChainConfig.Ethash != nil {
		accumulateRewards(backend.Config.ChainConfig, statedb, block.Header(), block.Uncles())
	}

	return statedb, nil
}

// accumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(config *params.ChainConfig, state *ipldstate.StateDB, header *types.Header, uncles []*types.Header) {
	// Select the correct block reward based on chain progression
	blockReward := ethash.FrontierBlockReward
	if config.IsByzantium(header.Number) {
		blockReward = ethash.ByzantiumBlockReward
	}

	if config.IsConstantinople(header.Number) {
		blockReward = ethash.ConstantinopleBlockReward
	}

	// Accumulate the rewards for the miner and any included uncles
	reward := new(big.Int).Set(blockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, blockReward)
		r.Div(r, big8)
		state.AddBalance(uncle.Coinbase, r)

		r.Div(blockReward, big32)
		reward.Add(reward, r)
	}

	state.AddBalance(header.Coinbase, reward)
}

func setChainConfig(ghash common.Hash) *params.ChainConfig {
	switch {
	case ghash == params.MainnetGenesisHash:
		return params.MainnetChainConfig
	case ghash == params.SepoliaGenesisHash:
		return params.SepoliaChainConfig
	case ghash == params.RinkebyGenesisHash:
		return params.RinkebyChainConfig
	case ghash == params.GoerliGenesisHash:
		return params.GoerliChainConfig
	default:
		return params.AllEthashProtocolChanges
	}
}

func getAuthor(b *ipldeth.Backend, header *types.Header) *common.Address {
	author, err := getEngine(b).Author(header)
	if err != nil {
		return nil
	}

	return &author
}

func getEngine(b *ipldeth.Backend) consensus.Engine {
	// TODO: add logic for other engines
	if b.Config.ChainConfig.Clique != nil {
		engine := clique.New(b.Config.ChainConfig.Clique, nil)
		return engine
	}

	return ethash.NewFaker()
}
