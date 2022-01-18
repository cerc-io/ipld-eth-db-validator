package validator

import (
	"context"
	"errors"
	"fmt"
	"math/big"
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
	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	ipfsethdb "github.com/vulcanize/ipfs-ethdb/postgres"
	ipldEth "github.com/vulcanize/ipld-eth-server/pkg/eth"
	ethServerShared "github.com/vulcanize/ipld-eth-server/pkg/shared"
)

var (
	big8     = big.NewInt(8)
	big32    = big.NewInt(32)
	testHash = common.HexToHash("0x1283a0bca5cce009bcf3e5a860eccdc202d1345f464024f2ee8ea1e1254349e7")
)

type service struct {
	db              *postgres.DB
	blockNum, trail uint64
	logger          *log.Logger
}

func NewService(db *postgres.DB, blockNum, trailNum uint64) *service {
	return &service{
		db:       db,
		blockNum: blockNum,
		trail:    trailNum,
		logger:   log.New(),
	}
}

func NewEthBackend(db *postgres.DB, c *ipldEth.Config) (*ipldEth.Backend, error) {
	gcc := c.GroupCacheConfig

	groupName := gcc.StateDB.Name
	if groupName == "" {
		groupName = ipldEth.StateDBGroupCacheName
	}

	r := ipldEth.NewCIDRetriever(db)
	ethDB := ipfsethdb.NewDatabase(db.DB, ipfsethdb.CacheConfig{
		Name:           groupName,
		Size:           gcc.StateDB.CacheSizeInMB * 1024 * 1024,
		ExpiryDuration: time.Minute * time.Duration(gcc.StateDB.CacheExpiryInMins),
	})

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

func NewDB(connectString string, config postgres.ConnectionConfig, node node.Info) (*postgres.DB, error) {
	db, connectErr := sqlx.Connect("postgres", connectString)
	if connectErr != nil {
		return &postgres.DB{}, postgres.ErrDBConnectionFailed(connectErr)
	}
	if config.MaxOpen > 0 {
		db.SetMaxOpenConns(config.MaxOpen)
	}
	if config.MaxIdle > 0 {
		db.SetMaxIdleConns(config.MaxIdle)
	}
	if config.MaxLifetime > 0 {
		lifetime := time.Duration(config.MaxLifetime) * time.Second
		db.SetConnMaxLifetime(lifetime)
	}
	pg := postgres.DB{DB: db, Node: node}
	return &pg, nil
}

// Start is used to begin the service
func (s *service) Start(ctx context.Context) (uint64, error) {
	api, err := ethAPI(ctx, s.db)
	if err != nil {
		return 0, err
	}

	idxBlockNum := s.blockNum
	headBlock, _ := api.B.BlockByNumber(ctx, rpc.LatestBlockNumber)
	headBlockNum := headBlock.NumberU64()

	for headBlockNum-s.trail >= idxBlockNum {
		validateBlock, err := api.B.BlockByNumber(ctx, rpc.BlockNumber(idxBlockNum))
		if err != nil {
			return idxBlockNum, err
		}

		stateDB, err := applyTransaction(validateBlock, api.B)
		if err != nil {
			return idxBlockNum, err
		}

		blockStateRoot := validateBlock.Header().Root.String()

		dbStateRoot := stateDB.IntermediateRoot(true).String()
		if blockStateRoot != dbStateRoot {
			s.logger.Errorf("failed to verify state root at block %d", idxBlockNum)
			return idxBlockNum, fmt.Errorf("failed to verify state root at block")
		}

		s.logger.Infof("state root verified for block= %d", idxBlockNum)

		// again fetch head block
		headBlock, err = api.B.BlockByNumber(ctx, rpc.LatestBlockNumber)
		if err != nil {
			return idxBlockNum, err
		}

		headBlockNum = headBlock.NumberU64()
		idxBlockNum++
	}

	s.logger.Infof("last validated block %v", idxBlockNum)

	return idxBlockNum, nil
}

func ethAPI(ctx context.Context, db *postgres.DB) (*ipldEth.PublicEthAPI, error) {
	// TODO: decide network for custom chainConfig.
	backend, err := NewEthBackend(db, &ipldEth.Config{
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
	case ghash == testHash:
		return TestChainConfig
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
