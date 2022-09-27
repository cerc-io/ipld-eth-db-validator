package integration_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipld-eth-db-validator/pkg/validator"
	integration "github.com/vulcanize/ipld-eth-db-validator/test"

	"github.com/cerc-io/ipld-eth-server/v4/pkg/shared"
	ethServerIntegration "github.com/cerc-io/ipld-eth-server/v4/test"
)

const (
	blockNum               = 1
	trail                  = 0
	validatorSleepInterval = uint(5)
)

var (
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

var _ = Describe("Integration test", func() {
	ctx := context.Background()

	var contract *ethServerIntegration.ContractDeployed
	var err error
	sleepInterval := 2 * time.Second
	timeout := 4 * time.Second

	db := shared.SetupDB()
	cfg := validator.Config{
		DB:            db,
		BlockNum:      blockNum,
		Trail:         trail,
		SleepInterval: validatorSleepInterval,
		ChainCfg:      validator.IntegrationTestChainConfig,
	}
	validationProgressChan := make(chan uint64)
	service := validator.NewService(&cfg, validationProgressChan)

	wg := new(sync.WaitGroup)

	It("test init", func() {
		wg.Add(1)
		go service.Start(ctx, wg)

		// Deploy a dummy contract as the first contract might get deployed at block number 0
		_, _ = ethServerIntegration.DeployContract()
		time.Sleep(sleepInterval)
	})

	defer It("test teardown", func() {
		service.Stop()
		wg.Wait()

		Expect(validationProgressChan).To(BeClosed())
	})

	Describe("Validate state", func() {
		It("performs validation on contract deployment", func() {
			contract, err = integration.DeployTestContract()
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(sleepInterval)

			Expect(validationProgressChan).ToNot(BeClosed())
			Eventually(validationProgressChan, timeout).Should(Receive(Equal(uint64(contract.BlockNumber))))
		})

		It("performs validation on contract transactions", func() {
			for i := 0; i < 10; i++ {
				res, txErr := integration.PutTestValue(contract.Address, i, i)
				Expect(txErr).ToNot(HaveOccurred())
				time.Sleep(sleepInterval)

				Expect(validationProgressChan).ToNot(BeClosed())
				Eventually(validationProgressChan, timeout).Should(Receive(Equal(uint64(res.BlockNumber))))
			}
		})

		It("performs validation on eth transfer transactions", func() {
			for _, address := range testAddresses {
				tx, txErr := ethServerIntegration.SendEth(address, "0.01")
				Expect(txErr).ToNot(HaveOccurred())
				time.Sleep(sleepInterval)

				Expect(validationProgressChan).ToNot(BeClosed())
				Eventually(validationProgressChan, timeout).Should(Receive(Equal(uint64(tx.BlockNumber))))
			}
		})
	})
})
