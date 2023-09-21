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
	"time"

	"github.com/cerc-io/plugeth-statediff/indexer/database/sql/postgres"
	"github.com/cerc-io/plugeth-statediff/utils"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/viper"
)

type Config struct {
	DBConfig postgres.Config
	DBStats  bool

	ChainConfig *params.ChainConfig
	// Used to trigger writing state diffs for gaps in the index
	Client                *rpc.Client
	FromBlock, Trail      uint64
	RetryInterval         time.Duration
	StateDiffMissingBlock bool
	StateDiffTimeout      time.Duration
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
	c.DBConfig.DatabaseName = viper.GetString("database.name")
	c.DBConfig.Hostname = viper.GetString("database.hostname")
	c.DBConfig.Port = viper.GetInt("database.port")
	c.DBConfig.Username = viper.GetString("database.user")
	c.DBConfig.Password = viper.GetString("database.password")

	c.DBConfig.MaxIdle = viper.GetInt("database.maxIdle")
	c.DBConfig.MaxConns = viper.GetInt("database.maxOpen")
	c.DBConfig.MaxConnLifetime = viper.GetDuration("database.maxLifetime")

	c.DBStats = viper.GetBool("prom.dbStats")

	return nil
}

func (c *Config) setupEth() error {
	var err error
	chainConfigPath := viper.GetString("ethereum.chainConfig")
	if chainConfigPath != "" {
		c.ChainConfig, err = utils.LoadConfig(chainConfigPath)
	} else {
		// read chainID if chain config path not provided
		chainID := viper.GetUint64("ethereum.chainID")
		c.ChainConfig, err = utils.ChainConfig(chainID)
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
	c.FromBlock = viper.GetUint64("validate.fromBlock")
	if c.FromBlock < 1 {
		return fmt.Errorf("starting block height cannot be less than 1")
	}

	c.Trail = viper.GetUint64("validate.trail")
	c.RetryInterval = viper.GetDuration("validate.retryInterval")
	c.StateDiffMissingBlock = viper.GetBool("validate.stateDiffMissingBlock")
	if c.StateDiffMissingBlock {
		c.StateDiffTimeout = viper.GetDuration("validate.stateDiffTimeout")
	}

	return err
}
