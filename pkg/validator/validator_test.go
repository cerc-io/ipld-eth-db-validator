package validator_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff/indexer/ipld"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"

	// import server helpers for non-canonical chain data
	server_mocks "github.com/cerc-io/ipld-eth-server/v5/pkg/eth/test_helpers"

	"github.com/cerc-io/ipld-eth-db-validator/v5/internal/chaingen"
	"github.com/cerc-io/ipld-eth-db-validator/v5/internal/helpers"
	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/validator"
)

const (
	chainLength = 10
	startBlock  = 1
)

var (
	chainConfig = TestChainConfig
	mockTD      = big.NewInt(1337)
	testDB      = rawdb.NewMemoryDatabase()
)

func init() {
	// The geth sync logs are noisy, silence them
	log.Root().SetHandler(log.DiscardHandler())
}

func setup(t *testing.T) *sqlx.DB {
	// Make the test blockchain and state
	gen := chaingen.DefaultGenContext(chainConfig, testDB)
	blocks, receipts, chain := gen.MakeChain(chainLength)

	t.Cleanup(func() {
		chain.Stop()
	})

	indexer, err := helpers.TestStateDiffIndexer(context.Background(), chainConfig, gen.Genesis.Hash())
	if err != nil {
		t.Fatal(err)
	}
	helpers.IndexChain(indexer, helpers.IndexChainParams{
		StateCache:      chain.StateCache(),
		Blocks:          blocks,
		Receipts:        receipts,
		TotalDifficulty: mockTD,
	})

	// Insert some non-canonical data into the database so that we test our ability to discern canonicity
	tx, err := indexer.PushBlock(server_mocks.MockBlock, server_mocks.MockReceipts,
		server_mocks.MockBlock.Difficulty())
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Submit(err)
	if err != nil {
		t.Fatal(err)
	}

	// The non-canonical header has a child
	tx, err = indexer.PushBlock(server_mocks.MockChild, server_mocks.MockReceipts, server_mocks.MockChild.Difficulty())
	if err != nil {
		t.Fatal(err)
	}

	ipld := sdtypes.IPLD{
		CID:     ipld.Keccak256ToCid(ipld.RawBinary, server_mocks.CodeHash.Bytes()).String(),
		Content: server_mocks.ContractCode,
	}
	err = indexer.PushIPLD(tx, ipld)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Submit(err)
	if err != nil {
		t.Fatal(err)
	}

	db := helpers.SetupDB()

	t.Cleanup(func() {
		helpers.TearDownDB(db)
	})
	return db
}

func TestStateValidation(t *testing.T) {
	db := setup(t)

	t.Run("Validator", func(t *testing.T) {
		api, err := validator.EthAPI(context.Background(), db, chainConfig)
		if err != nil {
			t.Fatal(err)
		}

		for i := uint64(startBlock); i <= chainLength; i++ {
			blockToBeValidated, err := api.B.BlockByNumber(context.Background(), rpc.BlockNumber(i))
			if err != nil {
				t.Fatal(err)
			}
			if blockToBeValidated == nil {
				t.Fatal("blockToBeValidated is nil")
			}

			err = validator.ValidateBlock(blockToBeValidated, api.B, i)
			if err != nil {
				t.Fatal(err)
			}

			err = validator.ValidateReferentialIntegrity(db, i)
			if err != nil {
				t.Fatal(err)
			}
		}
	})
}
