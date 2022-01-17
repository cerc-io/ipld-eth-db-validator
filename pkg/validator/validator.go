package validator

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	customEthDB := NewDatabase(ethDB)

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
	api, err := ethAPI(s.db)
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

func ethAPI(db *postgres.DB) (*ipldEth.PublicEthAPI, error) {
	// TODO: decide network for chainConfig.
	backend, err := NewEthBackend(db, &ipldEth.Config{
		ChainConfig: params.RinkebyChainConfig,
		GroupCacheConfig: &ethServerShared.GroupCacheConfig{
			StateDB: ethServerShared.GroupConfig{
				Name: "vulcanize_validator",
			},
		},
	})

	if err != nil {
		return nil, err
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
		ctx := core.NewEVMBlockContext(block.Header(), backend, nil)

		// Not yet the searched for transaction, execute on top of the current state
		newEVM := vm.NewEVM(ctx, txContext, stateDB, backend.Config.ChainConfig, vm.Config{})

		stateDB.Prepare(tx.Hash(), idx)
		if _, err := core.ApplyMessage(newEVM, msg, new(core.GasPool).AddGas(block.GasLimit())); err != nil {
			return nil, fmt.Errorf("transaction %#x failed: %w", tx.Hash(), err)
		}
	}
	return stateDB, nil
}
