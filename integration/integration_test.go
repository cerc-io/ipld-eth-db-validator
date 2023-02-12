package integration

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	ipldeth "github.com/cerc-io/ipld-eth-server/v5/pkg/eth"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/validator"
	"github.com/cerc-io/ipld-eth-db-validator/v5/test/helpers"
)

var (
	chainConfig = &params.ChainConfig{
		ChainID:             big.NewInt(1212),
		HomesteadBlock:      big.NewInt(1),
		EIP150Block:         big.NewInt(1),
		EIP155Block:         big.NewInt(1),
		EIP158Block:         big.NewInt(1),
		ByzantiumBlock:      big.NewInt(1),
		ConstantinopleBlock: big.NewInt(1),
		PetersburgBlock:     big.NewInt(1),
		IstanbulBlock:       big.NewInt(1),
		BerlinBlock:         big.NewInt(1),
		LondonBlock:         big.NewInt(1),

		Clique: &params.CliqueConfig{
			Period: 0,
			Epoch:  30000,
		},
	}

	testAddresses = []string{
		"0x1111111111111111111111111111111111111112",
		"0x1ca7c995f8eF0A2989BbcE08D5B7Efe50A584aa1",
		"0x9a4b666af23a2cdb4e5538e1d222a445aeb82134",
		"0xF7C7AEaECD2349b129d5d15790241c32eeE4607B",
		"0x992b6E9BFCA1F7b0797Cee10b0170E536EAd3532",
		"0x29ed93a7454Bc17a8D4A24D0627009eE0849B990",
		"0x66E3dCA826b04B5d4988F7a37c91c9b1041e579D",
		"0x96288939Ac7048c27E0E087b02bDaad3cd61b37b",
		"0xD354280BCd771541c935b15bc04342c26086FE9B",
		"0x7f887e25688c274E77b8DeB3286A55129B55AF14",
	}
)

const (
	timeout       = 2 * time.Minute
	retryInterval = 4 * time.Second
)

var (
	ctx      = context.Background()
	dbConfig = helpers.DBConfig
	wg       sync.WaitGroup
)

func setup(t *testing.T, progressChan chan uint64) *atomicBlockSet {
	// Start validator at current head, but not before PoS transition
	// (test chain Merge is at block 1)
	startFrom := latestBlock(t)
	if startFrom < 1 {
		startFrom = 1
	}
	cfg := validator.Config{
		DBConfig:      dbConfig,
		FromBlock:     startFrom,
		Trail:         0,
		RetryInterval: retryInterval,
		ChainConfig:   chainConfig,
	}
	service, err := validator.NewService(&cfg, progressChan)
	require.NoError(t, err)

	// Start tracking validated blocks, so we don't miss any
	validated := newBlockSet()
	go func() {
		for block := range progressChan {
			validated.add(block)
		}
	}()

	wg.Add(1)
	go service.Start(ctx, &wg)

	t.Cleanup(func() {
		service.Stop()
		wg.Wait()

		g := gomega.NewWithT(t)
		g.Expect(progressChan).To(BeClosed())
	})
	return validated
}

func TestValidateContracts(t *testing.T) {
	progressChan := make(chan uint64, 10)
	validated := setup(t, progressChan)

	contract, err := DeployTestContract()
	require.NoError(t, err)

	t.Run("contract deployment", func(t *testing.T) {
		g := gomega.NewWithT(t)
		g.Expect(progressChan).ToNot(BeClosed())
		g.Eventually(validated.contains, timeout).WithArguments(contract.BlockNumber).Should(BeTrue())
	})

	t.Run("contract method calls", func(t *testing.T) {
		g := gomega.NewWithT(t)

		var blocks []uint64
		for i := 0; i < 10; i++ {
			res, err := PutTestValue(contract.Address, i)
			require.NoError(t, err)
			t.Logf("Put() called at block %d", res.BlockNumber)
			blocks = append(blocks, res.BlockNumber)
		}

		g.Expect(progressChan).ToNot(BeClosed())
		g.Eventually(validated.containsAll, timeout).WithArguments(blocks).Should(BeTrue())
	})
}

func TestValidateTransactions(t *testing.T) {
	progressChan := make(chan uint64, 100)
	validated := setup(t, progressChan)

	t.Run("eth transfer transactions", func(t *testing.T) {
		g := gomega.NewWithT(t)

		var blocks []uint64
		for _, address := range testAddresses {
			tx, err := SendEth(address, "0.01")
			require.NoError(t, err)
			t.Logf("Sent tx at block %d", tx.BlockNumber)
			blocks = append(blocks, tx.BlockNumber)
		}

		g.Expect(progressChan).ToNot(BeClosed())
		g.Eventually(validated.containsAll, timeout).WithArguments(blocks).Should(BeTrue())
	})
}

type atomicBlockSet struct {
	blocks map[uint64]struct{}
	sync.Mutex
}

func newBlockSet() *atomicBlockSet {
	return &atomicBlockSet{blocks: make(map[uint64]struct{})}
}

func (set *atomicBlockSet) contains(block uint64) bool {
	set.Lock()
	defer set.Unlock()
	for done := range set.blocks {
		if done == block {
			return true
		}
	}
	return false
}

func (set *atomicBlockSet) containsAll(blocks []uint64) bool {
	for _, block := range blocks {
		if !set.contains(block) {
			return false
		}
	}
	return true
}

func (set *atomicBlockSet) add(block uint64) {
	set.Lock()
	defer set.Unlock()
	set.blocks[block] = struct{}{}
}

func latestBlock(t *testing.T) uint64 {
	db, err := postgres.ConnectSQLX(ctx, dbConfig)
	require.NoError(t, err)
	defer db.Close()

	retriever := ipldeth.NewRetriever(db)
	lastBlock, err := retriever.RetrieveLastBlockNumber()
	require.NoError(t, err)
	return uint64(lastBlock)
}
