module github.com/Vulcanize/ipld-eth-db-validator

go 1.17

require (
	github.com/ethereum/go-ethereum v1.10.14
	github.com/jmoiron/sqlx v1.2.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.0
	github.com/vulcanize/ipfs-ethdb v0.0.6
	github.com/vulcanize/ipld-eth-server v0.3.9
)

replace (
	github.com/ethereum/go-ethereum v1.10.14 => github.com/vulcanize/go-ethereum v1.10.14-statediff-0.0.29
	github.com/vulcanize/ipld-eth-server v0.3.9 => github.com/vulcanize/ipld-eth-server v0.0.0-20220117121622-5ea788912bd3
)
