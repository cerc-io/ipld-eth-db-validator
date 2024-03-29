package validator_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cerc-io/ipld-eth-db-validator/v5/internal/chaingen"
	"github.com/cerc-io/ipld-eth-db-validator/v5/internal/helpers"
	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/validator"
)

func TestRefIntegrity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ETH IPLD validator ref integrity suite test")
}

var _ = Describe("referential integrity", Ordered, func() {
	var (
		db           *sqlx.DB
		tx           *sqlx.Tx
		checkedBlock *types.Block // Generated block of interest
	)
	BeforeAll(func() {
		var (
			blocks      []*types.Block
			receipts    []types.Receipts
			chain       *core.BlockChain
			chainConfig = TestChainConfig
			mockTD      = big.NewInt(1337)
			testdb      = rawdb.NewMemoryDatabase()
		)

		gen := chaingen.DefaultGenContext(chainConfig, testdb)
		gen.AddFunction(func(i int, block *core.BlockGen) {
			if i >= 2 {
				uncle := &types.Header{
					Number:      big.NewInt(int64(i - 1)),
					Root:        common.HexToHash("0x1"),
					TxHash:      common.HexToHash("0x1"),
					ReceiptHash: common.HexToHash("0x1"),
					ParentHash:  block.PrevBlock(i - 1).Hash(),
				}
				block.AddUncle(uncle)
			}
		})
		blocks, receipts, chain = gen.MakeChain(5)

		indexer, err := helpers.TestStateDiffIndexer(context.Background(), chainConfig, gen.Genesis.Hash())
		Expect(err).ToNot(HaveOccurred())
		helpers.IndexChain(indexer, helpers.IndexChainParams{
			StateCache:      chain.StateCache(),
			Blocks:          blocks,
			Receipts:        receipts,
			TotalDifficulty: mockTD,
		})
		checkedBlock = blocks[5]

		db = helpers.SetupDB()
	})
	AfterAll(func() { helpers.TearDownDB(db) })

	BeforeEach(func() { tx = db.MustBegin() })
	AfterEach(func() {
		tx.Rollback()
	})

	Describe("ValidateHeaderCIDsRef", func() {
		It("Validates referential integrity of header_cids table", func() {
			err := validator.ValidateHeaderCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateHeaderCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateUncleCIDsRef", func() {
		It("Validates referential integrity of uncle_cids table", func() {
			err := validator.ValidateUncleCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header_cid entry not found", func() {
			err := deleteEntriesFrom(tx, "eth.header_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateUncleCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.header_cids"))
		})

		It("Throws an error if corresponding uncle IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateUncleCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateTransactionCIDsRef", func() {
		It("Validates referential integrity of transaction_cids table", func() {
			err := validator.ValidateTransactionCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header_cid entry not found", func() {
			err := deleteEntriesFrom(tx, "eth.header_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateTransactionCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.header_cids"))
		})

		It("Throws an error if corresponding transaction IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateTransactionCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateReceiptCIDsRef", func() {
		It("Validates referential integrity of receipt_cids table", func() {
			err := validator.ValidateReceiptCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding transaction_cids entry not found", func() {
			err := deleteEntriesFrom(tx, "eth.transaction_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateReceiptCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.transaction_cids"))
		})

		It("Throws an error if corresponding receipt IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateReceiptCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateStateCIDsRef", func() {
		It("Validates referential integrity of state_cids table", func() {
			err := validator.ValidateStateCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header_cids entry not found", func() {
			err := deleteEntriesFrom(tx, "eth.header_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStateCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.header_cids"))
		})

		It("Throws an error if corresponding state IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStateCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateStorageCIDsRef", func() {
		It("Validates referential integrity of storage_cids table", func() {
			err := validator.ValidateStorageCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding state_cids entry not found", func() {
			err := deleteEntriesFrom(tx, "eth.state_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStorageCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.state_cids"))
		})

		It("Throws an error if corresponding storage IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStorageCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateLogCIDsRef", func() {
		It("Validates referential integrity of log_cids table", func() {
			err := validator.ValidateLogCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding receipt_cids entry not found", func() {
			err := deleteEntriesFrom(tx, "eth.receipt_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateLogCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.receipt_cids"))
		})

		It("Throws an error if corresponding log IPFS block entry not found", func() {
			err := deleteEntriesFrom(tx, "ipld.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateLogCIDsRef(tx, checkedBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "ipld.blocks"))
		})
	})

	Describe("ValidateReferentialIntegrity", func() {
		It("Validates referential integrity of full chain", func() {
			for i := uint64(startBlock); i <= chainLength; i++ {
				err := validator.ValidateReferentialIntegrity(tx, i)
				Expect(err).ToNot(HaveOccurred())
			}
		})
	})
})

func deleteEntriesFrom(tx *sqlx.Tx, tableName string) error {
	pgStr := "TRUNCATE %s"
	_, err := tx.Exec(fmt.Sprintf(pgStr, tableName))
	return err
}
