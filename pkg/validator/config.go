// VulcanizeDB
// Copyright © 2022 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package validator

import (
	"fmt"
	"math/big"
	"time"

	"github.com/cerc-io/ipld-eth-db-validator/pkg/prom"
	"github.com/cerc-io/ipld-eth-server/v4/pkg/shared"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var IntegrationTestChainConfig = &params.ChainConfig{
	ChainID:             big.NewInt(99),
	HomesteadBlock:      big.NewInt(0),
	EIP150Block:         big.NewInt(0),
	EIP155Block:         big.NewInt(0),
	EIP158Block:         big.NewInt(0),
	ByzantiumBlock:      big.NewInt(0),
	ConstantinopleBlock: big.NewInt(0),
	PetersburgBlock:     big.NewInt(0),
	IstanbulBlock:       big.NewInt(0),
	Clique: &params.CliqueConfig{
		Period: 0,
		Epoch:  30000,
	},
}

var TestChainConfig = &params.ChainConfig{
	ChainID:             big.NewInt(1),
	HomesteadBlock:      big.NewInt(0),
	EIP150Block:         big.NewInt(0),
	EIP155Block:         big.NewInt(0),
	EIP158Block:         big.NewInt(0),
	ByzantiumBlock:      big.NewInt(0),
	ConstantinopleBlock: big.NewInt(0),
	PetersburgBlock:     big.NewInt(0),
	IstanbulBlock:       big.NewInt(0),
	MuirGlacierBlock:    big.NewInt(0),
	BerlinBlock:         big.NewInt(0),
	LondonBlock:         big.NewInt(6),
	ArrowGlacierBlock:   big.NewInt(0),
	Ethash:              new(params.EthashConfig),
}

type Config struct {
	dbConfig postgres.Config
	DB       *sqlx.DB

	ChainCfg              *params.ChainConfig
	Client                *rpc.Client
	StateDiffMissingBlock bool
	StateDiffTimeout      uint

	BlockNum, Trail uint64
	SleepInterval   uint
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	err := cfg.setupDB()
	if err != nil {
		return nil, err
	}

	err = cfg.setupEth()
	if err != nil {
		return nil, err
	}

	err = cfg.setupValidator()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) setupDB() error {
	// DB Config
	c.dbConfig.DatabaseName = viper.GetString("database.name")
	c.dbConfig.Hostname = viper.GetString("database.hostname")
	c.dbConfig.Port = viper.GetInt("database.port")
	c.dbConfig.Username = viper.GetString("database.user")
	c.dbConfig.Password = viper.GetString("database.password")

	c.dbConfig.MaxIdle = viper.GetInt("database.maxIdle")
	c.dbConfig.MaxConns = viper.GetInt("database.maxOpen")
	c.dbConfig.MaxConnLifetime = time.Duration(viper.GetInt("database.maxLifetime"))

	// Create DB
	db, err := shared.NewDB(c.dbConfig.DbConnectionString(), c.dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	c.DB = db

	// Enable DB stats
	if viper.GetBool("prom.dbStats") {
		prom.RegisterDBCollector(c.dbConfig.DatabaseName, c.DB)
	}

	return nil
}

func (c *Config) setupEth() error {
	var err error
	chainConfigPath := viper.GetString("ethereum.chainConfig")
	if chainConfigPath != "" {
		c.ChainCfg, err = statediff.LoadConfig(chainConfigPath)
	} else {
		// read chainID if chain config path not provided
		chainID := viper.GetUint64("ethereum.chainID")
		c.ChainCfg, err = statediff.ChainConfig(chainID)
	}
	if err != nil {
		return err
	}

	// setup a statediffing client
	ethHTTP := viper.GetString("ethereum.httpPath")
	if ethHTTP != "" {
		ethHTTPEndpoint := fmt.Sprintf("http://%s", ethHTTP)
		c.Client, err = rpc.Dial(ethHTTPEndpoint)
	}

	return err
}

func (c *Config) setupValidator() error {
	var err error
	c.BlockNum = viper.GetUint64("validate.blockHeight")
	if c.BlockNum < 1 {
		return fmt.Errorf("block height cannot be less the 1")
	}

	c.Trail = viper.GetUint64("validate.trail")
	c.SleepInterval = viper.GetUint("validate.sleepInterval")
	c.StateDiffMissingBlock = viper.GetBool("validate.stateDiffMissingBlock")
	if c.StateDiffMissingBlock {
		c.StateDiffTimeout = viper.GetUint("validate.stateDiffTimeout")
	}

	return err
}
