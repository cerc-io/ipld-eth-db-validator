package integration_test

import (
	"context"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vulcanize/ipld-eth-db-validator/pkg/validator"

	"github.com/vulcanize/ipld-eth-server/pkg/eth"
	integration "github.com/vulcanize/ipld-eth-server/test"
)

const trail = 0

var _ = Describe("Integration test", func() {
	directProxyEthCalls, err := strconv.ParseBool(os.Getenv("ETH_FORWARD_ETH_CALLS"))
	Expect(err).To(BeNil())

	Expect(err).ToNot(HaveOccurred())
	ctx := context.Background()

	var contract *integration.ContractDeployed
	var contractErr error
	sleepInterval := 5 * time.Second

	Describe("Validate state", func() {
		BeforeEach(func() {
			if directProxyEthCalls {
				Skip("skipping no-direct-proxy-forwarding integration tests")
			}
			contract, contractErr = integration.DeployContract()
			time.Sleep(sleepInterval)
		})

		It("Validate state root", func() {
			Expect(contractErr).ToNot(HaveOccurred())

			db := eth.SetupTestDB()
			srvc := validator.NewService(db, uint64(contract.BlockNumber), trail, validator.IntegrationTestChainConfig)
			_, err = srvc.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
