package validator_test_test

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/statediff"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vulcanize/ipld-eth-server/pkg/eth"
	"github.com/vulcanize/ipld-eth-server/pkg/eth/test_helpers"

	"github.com/vulcanize/ipld-eth-db-validator/pkg/validator"
	"github.com/vulcanize/ipld-eth-db-validator/validator_test"
)

const (
	chainLength = 20
	blockHeight = 1
	trail       = 2
)

var _ = Describe("eth state reading tests", func() {
	var (
		blocks      []*types.Block
		receipts    []types.Receipts
		chain       *core.BlockChain
		db          *sqlx.DB
		chainConfig = validator.TestChainConfig
		mockTD      = big.NewInt(1337)
	)

	It("test init", func() {
		db = eth.SetupTestDB()
		transformer := eth.SetupTestStateDiffIndexer(context.Background(), chainConfig, validator_test.Genesis.Hash())

		// make the test blockchain (and state)
		blocks, receipts, chain = validator_test.MakeChain(chainLength, validator_test.Genesis, validator_test.TestChainGen)
		params := statediff.Params{
			IntermediateStateNodes:   true,
			IntermediateStorageNodes: true,
		}

		// iterate over the blocks, generating statediff payloads, and transforming the data into Postgres
		builder := statediff.NewBuilder(chain.StateCache())
		for i, block := range blocks {
			var args statediff.Args
			var rcts types.Receipts
			if i == 0 {
				args = statediff.Args{
					OldStateRoot: common.Hash{},
					NewStateRoot: block.Root(),
					BlockNumber:  block.Number(),
					BlockHash:    block.Hash(),
				}
			} else {
				args = statediff.Args{
					OldStateRoot: blocks[i-1].Root(),
					NewStateRoot: block.Root(),
					BlockNumber:  block.Number(),
					BlockHash:    block.Hash(),
				}
				rcts = receipts[i-1]
			}

			diff, err := builder.BuildStateDiffObject(args, params)
			Expect(err).ToNot(HaveOccurred())
			tx, err := transformer.PushBlock(block, rcts, mockTD)
			Expect(err).ToNot(HaveOccurred())

			for _, node := range diff.Nodes {
				err := transformer.PushStateNode(tx, node, block.Hash().String())
				Expect(err).ToNot(HaveOccurred())
			}

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		}

		// Insert some non-canonical data into the database so that we test our ability to discern canonicity
		indexAndPublisher := eth.SetupTestStateDiffIndexer(context.Background(), chainConfig, validator_test.Genesis.Hash())

		tx, err := indexAndPublisher.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
		Expect(err).ToNot(HaveOccurred())

		err = tx.Submit(err)
		Expect(err).ToNot(HaveOccurred())

		// The non-canonical header has a child
		tx, err = indexAndPublisher.PushBlock(test_helpers.MockChild, test_helpers.MockReceipts, test_helpers.MockChild.Difficulty())
		Expect(err).ToNot(HaveOccurred())

		hash := sdtypes.CodeAndCodeHash{
			Hash: test_helpers.CodeHash,
			Code: test_helpers.ContractCode,
		}

		err = indexAndPublisher.PushCodeAndCodeHash(tx, hash)
		Expect(err).ToNot(HaveOccurred())

		err = tx.Submit(err)
		Expect(err).ToNot(HaveOccurred())
	})

	defer It("test teardown", func() {
		eth.TearDownTestDB(db)
		chain.Stop()
	})

	Describe("state_validation", func() {
		It("Validator", func() {
			srvc := validator.NewService(db, blockHeight, trail, validator.TestChainConfig)

			_, err := srvc.Start(context.Background())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
