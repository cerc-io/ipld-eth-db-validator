package validator_test

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/statediff/indexer/interfaces"
	"github.com/ethereum/go-ethereum/statediff/indexer/mocks"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cerc-io/ipld-eth-db-validator/pkg/validator"
	"github.com/cerc-io/ipld-eth-server/v4/pkg/eth/test_helpers"
	"github.com/cerc-io/ipld-eth-server/v4/pkg/shared"
)

var _ = Describe("RefIntegrity", func() {
	var (
		ctx = context.Background()

		db          *sqlx.DB
		diffIndexer interfaces.StateDiffIndexer
	)

	BeforeEach(func() {
		db = shared.SetupDB()
		diffIndexer = shared.SetupTestStateDiffIndexer(ctx, params.TestChainConfig, test_helpers.Genesis.Hash())
	})

	AfterEach(func() {
		shared.TearDownDB(db)
	})

	Describe("ValidateHeaderCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of header_cids table", func() {
			err := validator.ValidateHeaderCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateHeaderCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

	Describe("ValidateUncleCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of uncle_cids table", func() {
			err := validator.ValidateUncleCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header_cid entry not found", func() {
			err := deleteEntriesFrom(db, "eth.header_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateUncleCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.header_cids"))
		})

		It("Throws an error if corresponding uncle IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateUncleCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

	Describe("ValidateTransactionCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of transaction_cids table", func() {
			err := validator.ValidateTransactionCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header_cid entry not found", func() {
			err := deleteEntriesFrom(db, "eth.header_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateTransactionCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.header_cids"))
		})

		It("Throws an error if corresponding transaction IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateTransactionCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

	Describe("ValidateReceiptCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of receipt_cids table", func() {
			err := validator.ValidateReceiptCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding transaction_cids entry not found", func() {
			err := deleteEntriesFrom(db, "eth.transaction_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateReceiptCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.transaction_cids"))
		})

		It("Throws an error if corresponding receipt IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateReceiptCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

	Describe("ValidateStateCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			for _, node := range test_helpers.MockStateNodes {
				err = diffIndexer.PushStateNode(tx, node, test_helpers.MockBlock.Hash().String())
				Expect(err).ToNot(HaveOccurred())
			}

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of state_cids table", func() {
			err := validator.ValidateStateCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding header_cids entry not found", func() {
			err := deleteEntriesFrom(db, "eth.header_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStateCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.header_cids"))
		})

		It("Throws an error if corresponding state IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStateCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

	Describe("ValidateStorageCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			for _, node := range test_helpers.MockStateNodes {
				err = diffIndexer.PushStateNode(tx, node, test_helpers.MockBlock.Hash().String())
				Expect(err).ToNot(HaveOccurred())
			}

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of storage_cids table", func() {
			err := validator.ValidateStorageCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding state_cids entry not found", func() {
			err := deleteEntriesFrom(db, "eth.state_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStorageCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.state_cids"))
		})

		It("Throws an error if corresponding storage IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStorageCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

	Describe("ValidateStateAccountsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			for _, node := range test_helpers.MockStateNodes {
				err = diffIndexer.PushStateNode(tx, node, test_helpers.MockBlock.Hash().String())
				Expect(err).ToNot(HaveOccurred())
			}

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of state_accounts table", func() {
			err := validator.ValidateStateAccountsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding state_cids entry not found", func() {
			err := deleteEntriesFrom(db, "eth.state_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateStateAccountsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.state_cids"))
		})
	})

	Describe("ValidateAccessListElementsRef", func() {
		BeforeEach(func() {
			indexAndPublisher := shared.SetupTestStateDiffIndexer(ctx, mocks.TestConfig, test_helpers.Genesis.Hash())

			tx, err := indexAndPublisher.PushBlock(mocks.MockBlock, mocks.MockReceipts, mocks.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of access_list_elements table", func() {
			err := validator.ValidateAccessListElementsRef(db, mocks.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding transaction_cids entry not found", func() {
			err := deleteEntriesFrom(db, "eth.transaction_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateAccessListElementsRef(db, mocks.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.transaction_cids"))
		})
	})

	Describe("ValidateLogCIDsRef", func() {
		BeforeEach(func() {
			tx, err := diffIndexer.PushBlock(test_helpers.MockBlock, test_helpers.MockReceipts, test_helpers.MockBlock.Difficulty())
			Expect(err).ToNot(HaveOccurred())

			for _, node := range test_helpers.MockStateNodes {
				err = diffIndexer.PushStateNode(tx, node, test_helpers.MockBlock.Hash().String())
				Expect(err).ToNot(HaveOccurred())
			}

			err = tx.Submit(err)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Validates referential integrity of log_cids table", func() {
			err := validator.ValidateLogCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).ToNot(HaveOccurred())
		})

		It("Throws an error if corresponding receipt_cids entry not found", func() {
			err := deleteEntriesFrom(db, "eth.receipt_cids")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateLogCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "eth.receipt_cids"))
		})

		It("Throws an error if corresponding log IPFS block entry not found", func() {
			err := deleteEntriesFrom(db, "public.blocks")
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateLogCIDsRef(db, test_helpers.MockBlock.NumberU64())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(validator.EntryNotFoundErr, "public.blocks"))
		})
	})

})

func deleteEntriesFrom(db *sqlx.DB, tableName string) error {
	pgStr := "DELETE FROM %s"
	_, err := db.Exec(fmt.Sprintf(pgStr, tableName))
	return err
}
