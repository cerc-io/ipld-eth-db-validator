package validator_test

import (
	"context"
	"math/big"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	// import server helpers for non-canonical chain data
	server_mocks "github.com/cerc-io/ipld-eth-server/v5/pkg/eth/test_helpers"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff/indexer/ipld"
	sdtypes "github.com/ethereum/go-ethereum/statediff/types"

	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/validator"
	"github.com/cerc-io/ipld-eth-db-validator/v5/test/chaingen"
	"github.com/cerc-io/ipld-eth-db-validator/v5/test/helpers"
)

const (
	chainLength = 10
	startBlock  = 1
)

func init() {
	// The geth sync logs are noisy, silence them
	log.Root().SetHandler(log.DiscardHandler())
}

var _ = Describe("State validation", func() {
	var (
		chain       *core.BlockChain
		db          *sqlx.DB
		chainConfig = validator.TestChainConfig
		mockTD      = big.NewInt(1337)
	)

	BeforeEach(func() {
		var (
			blocks   []*types.Block
			receipts []types.Receipts
		)
		// make the test blockchain (and state)
		gen := chaingen.DefaultGenContext(chainConfig, helpers.TestDB)
		blocks, receipts, chain = gen.MakeChain(chainLength)

		indexer, err := helpers.TestStateDiffIndexer(context.Background(), chainConfig, gen.Genesis.Hash())
		Expect(err).ToNot(HaveOccurred())
		helpers.IndexChain(indexer, helpers.IndexChainParams{
			StateCache:      chain.StateCache(),
			Blocks:          blocks,
			Receipts:        receipts,
			TotalDifficulty: mockTD,
		})

		// Insert some non-canonical data into the database so that we test our ability to discern canonicity
		tx, err := indexer.PushBlock(server_mocks.MockBlock, server_mocks.MockReceipts,
			server_mocks.MockBlock.Difficulty())
		Expect(err).ToNot(HaveOccurred())

		err = tx.Submit(err)
		Expect(err).ToNot(HaveOccurred())

		// The non-canonical header has a child
		tx, err = indexer.PushBlock(server_mocks.MockChild, server_mocks.MockReceipts, server_mocks.MockChild.Difficulty())
		Expect(err).ToNot(HaveOccurred())

		ipld := sdtypes.IPLD{
			CID:     ipld.Keccak256ToCid(ipld.RawBinary, server_mocks.CodeHash.Bytes()).String(),
			Content: server_mocks.ContractCode,
		}
		err = indexer.PushIPLD(tx, ipld)
		Expect(err).ToNot(HaveOccurred())

		err = tx.Submit(err)
		Expect(err).ToNot(HaveOccurred())

		db = helpers.SetupDB()
	})

	AfterEach(func() {
		helpers.TearDownDB(db)
		chain.Stop()
	})

	It("Validator", func() {
		api, err := validator.EthAPI(context.Background(), db, chainConfig)
		Expect(err).ToNot(HaveOccurred())

		for i := uint64(startBlock); i <= chainLength; i++ {
			blockToBeValidated, err := api.B.BlockByNumber(context.Background(), rpc.BlockNumber(i))
			Expect(err).ToNot(HaveOccurred())
			Expect(blockToBeValidated).ToNot(BeNil())

			err = validator.ValidateBlock(blockToBeValidated, api.B, i)
			Expect(err).ToNot(HaveOccurred())

			err = validator.ValidateReferentialIntegrity(db, i)
			Expect(err).ToNot(HaveOccurred())
		}
	})
})
