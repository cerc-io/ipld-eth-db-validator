package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	// imported to register env vars with viper
	_ "github.com/cerc-io/ipld-eth-db-validator/v5/cmd"
	"github.com/cerc-io/ipld-eth-db-validator/v5/integration"
	"github.com/cerc-io/ipld-eth-db-validator/v5/internal/helpers"
	"github.com/cerc-io/ipld-eth-db-validator/v5/pkg/validator"
)

const (
	timeout            = 20 * time.Minute
	pollInterval       = time.Second
	progressBufferSize = 200
)

var (
	testAddresses = []string{
		"0x1111111111111111111111111111111111111112",
		"0x1ca7c995f8eF0A2989BbcE08D5B7Efe50A584aa1",
		"0x9a4b666af23a2cdb4e5538e1d222a445aeb82134",
		"0xF7C7AEaECD2349b129d5d15790241c32eeE4607B",
		"0x992b6E9BFCA1F7b0797Cee10b0170E536EAd3532",
	}

	// Track the blocks validated on this chain
	lastValidated uint64
	validated     = newBlockSet()

	ctx = context.Background()
	wg  sync.WaitGroup
)

func init() {
	gomega.SetDefaultEventuallyTimeout(timeout)
	gomega.SetDefaultEventuallyPollingInterval(pollInterval)
}

func setup(t *testing.T, progressChan chan uint64) {
	cfg, err := validator.NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	// set the default DB config to the testing defaults
	cfg.DBConfig, _ = helpers.TestDBConfig.WithEnv()
	// update the start block if we have already validated past it
	if lastValidated > cfg.FromBlock {
		cfg.FromBlock = lastValidated
	}

	service, err := validator.NewService(cfg, progressChan)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for block := range progressChan {
			validated.add(block)
			lastValidated = block
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
}

func TestValidateContracts(t *testing.T) {
	progressChan := make(chan uint64, progressBufferSize)
	setup(t, progressChan)

	contract, err := integration.DeployTestContract()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("contract deployment", func(t *testing.T) {
		g := gomega.NewWithT(t)
		t.Logf("Deployed contract at block %d", contract.BlockNumber)

		g.Expect(progressChan).ToNot(BeClosed())
		g.Eventually(validated.contains, timeout).WithArguments(contract.BlockNumber).Should(BeTrue())
	})

	t.Run("contract method calls", func(t *testing.T) {
		g := gomega.NewWithT(t)

		var blocks []uint64
		for i := 0; i < 5; i++ {
			res, err := integration.PutTestValue(contract.Address, i)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("Put() called at block %d", res.BlockNumber)
			blocks = append(blocks, res.BlockNumber)
		}

		g.Expect(progressChan).ToNot(BeClosed())
		g.Eventually(validated.containsAll, timeout).WithArguments(blocks).Should(BeTrue())
	})
}

func TestValidateTransactions(t *testing.T) {
	progressChan := make(chan uint64, progressBufferSize)
	setup(t, progressChan)

	t.Run("ETH transfer transactions", func(t *testing.T) {
		g := gomega.NewWithT(t)

		var blocks []uint64
		for _, address := range testAddresses {
			tx, err := integration.SendEth(address, "0.01")
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("Sent tx at block %d", tx.BlockNumber)
			blocks = append(blocks, tx.BlockNumber)
		}

		g.Expect(progressChan).ToNot(BeClosed())
		g.Eventually(validated.containsAll, timeout).WithArguments(blocks).Should(BeTrue())
	})
}
