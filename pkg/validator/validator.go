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
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	ipfsethdb "github.com/vulcanize/ipfs-ethdb/v4/postgres"
	ipldEth "github.com/vulcanize/ipld-eth-server/v4/pkg/eth"
	ethServerShared "github.com/vulcanize/ipld-eth-server/v4/pkg/shared"

	"github.com/vulcanize/ipld-eth-db-validator/pkg/prom"
)

var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)

	ReferentialIntegrityErr = "referential integrity check failed at block %d, entry for %s not found"
	EntryNotFoundErr        = "entry for %s not found"
)

type service struct {
	db              *sqlx.DB
	blockNum, trail uint64
	sleepInterval   uint
	chainCfg        *params.ChainConfig

	stateDiffMissingBlock bool
	stateDiffTimeout      uint
	ethClient             *rpc.Client

	quitChan     chan bool
	progressChan chan uint64
}

func NewService(cfg *Config, progressChan chan uint64) *service {
	return &service{
		db:                    cfg.DB,
		blockNum:              cfg.BlockNum,
		trail:                 cfg.Trail,
		sleepInterval:         cfg.SleepInterval,
		chainCfg:              cfg.ChainCfg,
		stateDiffMissingBlock: cfg.StateDiffMissingBlock,
		stateDiffTimeout:      cfg.StateDiffTimeout,
		ethClient:             cfg.Client,
		quitChan:              make(chan bool),
		progressChan:          progressChan,
	}
}

func NewEthBackend(db *sqlx.DB, c *ipldEth.Config) (*ipldEth.Backend, error) {
	gcc := c.GroupCacheConfig

	groupName := gcc.StateDB.Name
	if groupName == "" {
		groupName = ipldEth.StateDBGroupCacheName
	}

	r := ipldEth.NewCIDRetriever(db)
	ethDB := ipfsethdb.NewDatabase(db, ipfsethdb.CacheConfig{
		Name:           groupName,
		Size:           gcc.StateDB.CacheSizeInMB * 1024 * 1024,
		ExpiryDuration: time.Minute * time.Duration(gcc.StateDB.CacheExpiryInMins),
	})

	// Read only wrapper around ipfs-ethdb eth.Database implementation
	customEthDB := newDatabase(ethDB)

	return &ipldEth.Backend{
		DB:            db,
		Retriever:     r,
		Fetcher:       ipldEth.NewIPLDFetcher(db),
		IPLDRetriever: ipldEth.NewIPLDRetriever(db),
		EthDB:         customEthDB,
		StateDatabase: state.NewDatabase(customEthDB),
		Config:        c,
	}, nil
}

// Start is used to begin the service
func (s *service) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	api, err := EthAPI(ctx, s.db, s.chainCfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	idxBlockNum := s.blockNum

	for {
		select {
		case <-s.quitChan:
			log.Infof("last validated block %v", idxBlockNum-1)
			if s.progressChan != nil {
				close(s.progressChan)
			}
			return
		default:
			idxBlockNum, err = s.Validate(ctx, api, idxBlockNum)
			if err != nil {
				log.Infof("last validated block %v", idxBlockNum-1)
				log.Fatal(err)
				return
			}

			prom.SetLastValidatedBlock(float64(idxBlockNum))
		}
	}
}

// Stop is used to gracefully stop the service
func (s *service) Stop() {
	log.Info("stopping ipld-eth-db-validator process")
	close(s.quitChan)
}

func (s *service) Validate(ctx context.Context, api *ipldEth.PublicEthAPI, idxBlockNum uint64) (uint64, error) {
	headBlockNum, err := fetchHeadBlockNumber(ctx, api)
	if err != nil {
		return idxBlockNum, err
	}

	// Check if it block at height idxBlockNum can be validated
	if idxBlockNum <= headBlockNum-s.trail {
		blockToBeValidated, err := api.B.BlockByNumber(ctx, rpc.BlockNumber(idxBlockNum))
		if err != nil {
			log.Errorf("failed to fetch block at height %d", idxBlockNum)
			return idxBlockNum, err
		}

		// Make a writeStateDiffAt call if block not found in the db
		if blockToBeValidated == nil {
			err = s.writeStateDiffAt(idxBlockNum)
			return idxBlockNum, err
		}

		err = ValidateBlock(blockToBeValidated, api.B, idxBlockNum)
		if err != nil {
			log.Errorf("failed to verify state root at block %d", idxBlockNum)
			return idxBlockNum, err
		}

		log.Infof("state root verified for block %d", idxBlockNum)

		err = ValidateReferentialIntegrity(s.db, idxBlockNum)
		if err != nil {
			log.Errorf("failed to verify referential integrity at block %d", idxBlockNum)
			return idxBlockNum, err
		}
		log.Infof("referential integrity verified for block %d", idxBlockNum)

		if s.progressChan != nil {
			s.progressChan <- idxBlockNum
		}

		idxBlockNum++
	} else {
		// Sleep / wait for head to move ahead
		time.Sleep(time.Second * time.Duration(s.sleepInterval))
	}

	return idxBlockNum, nil
}

// writeStateDiffAt calls out to a statediffing geth client to fill in a gap in the index
func (s *service) writeStateDiffAt(height uint64) error {
	if !s.stateDiffMissingBlock {
		return nil
	}

	var data json.RawMessage
	params := statediff.Params{
		IntermediateStateNodes:   true,
		IntermediateStorageNodes: true,
		IncludeBlock:             true,
		IncludeReceipts:          true,
		IncludeTD:                true,
		IncludeCode:              true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.stateDiffTimeout*uint(time.Second)))
	defer cancel()

	log.Warnf("making writeStateDiffAt call at height %d", height)
	if err := s.ethClient.CallContext(ctx, &data, "statediff_writeStateDiffAt", height, params); err != nil {
		log.Errorf("writeStateDiffAt %d faild with err %s", height, err.Error())
		return err
	}

	return nil
}

// ValidateBlock validates block at the given height
func ValidateBlock(blockToBeValidated *types.Block, b *ipldEth.Backend, blockNumber uint64) error {
	stateDB, err := applyTransaction(blockToBeValidated, b)
	if err != nil {
		return err
	}

	blockStateRoot := blockToBeValidated.Header().Root.String()

	dbStateRoot := stateDB.IntermediateRoot(true).String()
	if blockStateRoot != dbStateRoot {
		return fmt.Errorf("state roots do not match at block %d", blockNumber)
	}

	return nil
}

func EthAPI(ctx context.Context, db *sqlx.DB, chainCfg *params.ChainConfig) (*ipldEth.PublicEthAPI, error) {
	// TODO: decide network for custom chainConfig.
	backend, err := NewEthBackend(db, &ipldEth.Config{
		ChainConfig: chainCfg,
		GroupCacheConfig: &ethServerShared.GroupCacheConfig{
			StateDB: ethServerShared.GroupConfig{
				Name: "vulcanize_validator",
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

	return ipldEth.NewPublicEthAPI(backend, nil, false, false, false)
}

// fetchHeadBlockNumber gets the latest block number from the db
func fetchHeadBlockNumber(ctx context.Context, api *ipldEth.PublicEthAPI) (uint64, error) {
	headBlock, err := api.B.BlockByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		return 0, err
	}

	return headBlock.NumberU64(), nil
}

// applyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the stateDB of parent with applied transa
func applyTransaction(block *types.Block, backend *ipldEth.Backend) (*state.StateDB, error) {
	if block.NumberU64() == 0 {
		return nil, errors.New("no transaction in genesis")
	}

	// Create the parent state database
	parentHash := block.ParentHash()
	parentRPCBlockHash := rpc.BlockNumberOrHash{
		BlockHash:        &parentHash,
		RequireCanonical: false,
	}

	parent, _ := backend.BlockByNumberOrHash(context.Background(), parentRPCBlockHash)
	if parent == nil {
		return nil, fmt.Errorf("parent %#x not found", block.ParentHash())
	}

	stateDB, _, err := backend.StateAndHeaderByNumberOrHash(context.Background(), parentRPCBlockHash)
	if err != nil {
		return nil, err
	}

	signer := types.MakeSigner(backend.Config.ChainConfig, block.Number())

	for idx, tx := range block.Transactions() {
		// Assemble the transaction call message and return if the requested offset
		msg, _ := tx.AsMessage(signer, block.BaseFee())
		txContext := core.NewEVMTxContext(msg)
		ctx := core.NewEVMBlockContext(block.Header(), backend, getAuthor(backend, block.Header()))

		// Not yet the searched for transaction, execute on top of the current state
		newEVM := vm.NewEVM(ctx, txContext, stateDB, backend.Config.ChainConfig, vm.Config{})

		stateDB.Prepare(tx.Hash(), idx)
		if _, err := core.ApplyMessage(newEVM, msg, new(core.GasPool).AddGas(block.GasLimit())); err != nil {
			return nil, fmt.Errorf("transaction %#x failed: %w", tx.Hash(), err)
		}
	}

	if backend.Config.ChainConfig.Ethash != nil {
		accumulateRewards(backend.Config.ChainConfig, stateDB, block.Header(), block.Uncles())
	}

	return stateDB, nil
}

// accumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
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
	case ghash == params.RopstenGenesisHash:
		return params.RopstenChainConfig
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

func getAuthor(b *ipldEth.Backend, header *types.Header) *common.Address {
	author, err := getEngine(b).Author(header)
	if err != nil {
		return nil
	}

	return &author
}

func getEngine(b *ipldEth.Backend) consensus.Engine {
	// TODO: add logic for other engines
	if b.Config.ChainConfig.Clique != nil {
		engine := clique.New(b.Config.ChainConfig.Clique, nil)
		return engine
	}

	return ethash.NewFaker()
}
