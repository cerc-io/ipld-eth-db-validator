package integration_test

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Vulcanize/ipld-eth-db-validator/pkg/validator"

	integration "github.com/vulcanize/ipld-eth-server/test"
)

const trail = 0

var randomAddr = common.HexToAddress("0x1C3ab14BBaD3D99F4203bd7a11aCB94882050E6f")

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

			db, _ := setupDB()
			srvc := validator.NewService(db, uint64(contract.BlockNumber), trail, validator.IntegrationTestChainConfig)
			_, err = srvc.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

		})
	})
})

func setupDB() (*postgres.DB, error) {
	uri := postgres.DbConnectionString(postgres.ConnectionParams{
		User:     "vdbm",
		Password: "password",
		Hostname: "localhost",
		Name:     "vulcanize_testing",
		Port:     8077,
	})
	return validator.NewDB(uri, postgres.ConnectionConfig{}, node.Info{})
}
