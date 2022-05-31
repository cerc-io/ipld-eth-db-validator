package integration_test

import (
	"context"
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

	Describe("Validate state", func() {
		BeforeEach(func() {
			// Deploy a dummy contract as the first contract might get deployed at block number 0
			_, _ = integration.DeployContract()
			time.Sleep(sleepInterval)

			contract, contractErr = integration.DeployContract()
			time.Sleep(sleepInterval)
		})

		It("Validate state root", func() {
			Expect(contractErr).ToNot(HaveOccurred())

			db := shared.SetupDB()
			srvc := validator.NewService(db, uint64(contract.BlockNumber), trail, validatorSleepInterval, validator.IntegrationTestChainConfig)
			srvc.Start(ctx, nil)
		})
	})
})
