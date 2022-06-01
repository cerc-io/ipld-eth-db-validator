package integration_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipld-eth-db-validator/pkg/validator"

	"github.com/vulcanize/ipld-eth-server/v3/pkg/shared"
	integration "github.com/vulcanize/ipld-eth-server/v3/test"
)

const (
	trail                  = 0
	validatorSleepInterval = uint(5)
)

var _ = Describe("Integration test", func() {
	ctx := context.Background()

	var contract *integration.ContractDeployed
	var contractErr error
	sleepInterval := 5 * time.Second

	db := shared.SetupDB()
	validationProgressChan := make(chan uint64)
	service := validator.NewService(db, 1, trail, validatorSleepInterval, validator.IntegrationTestChainConfig, validationProgressChan)

	wg := new(sync.WaitGroup)

	It("test init", func() {
		wg.Add(1)
		go service.Start(ctx, wg)
	})

	defer It("test teardown", func() {
		service.Stop()
		wg.Wait()

		Expect(validationProgressChan).To(BeClosed())
	})

	Describe("Validate state", func() {
		BeforeEach(func() {
			// Deploy a dummy contract as the first contract might get deployed at block number 0
			_, _ = integration.DeployContract()
			time.Sleep(sleepInterval)

			contract, contractErr = integration.DeployContract()
			time.Sleep(sleepInterval)
		})

		It("performs state root validation", func() {
			Expect(contractErr).ToNot(HaveOccurred())

			Expect(validationProgressChan).ToNot(BeClosed())
			Eventually(validationProgressChan).Should(Receive(Equal(uint64(contract.BlockNumber))))
		})
	})
})
