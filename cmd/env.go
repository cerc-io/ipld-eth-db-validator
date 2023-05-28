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

package cmd

import (
	"github.com/spf13/viper"
)

const (
	LOG_LEVEL     = "LOG_LEVEL"
	LOG_FILE_PATH = "LOG_FILE_PATH"

	PROM_METRICS   = "PROM_METRICS"
	PROM_HTTP      = "PROM_HTTP"
	PROM_HTTP_ADDR = "PROM_HTTP_ADDR"
	PROM_HTTP_PORT = "PROM_HTTP_PORT"
	PROM_DB_STATS  = "PROM_DB_STATS"

	DATABASE_NAME     = "DATABASE_NAME"
	DATABASE_HOSTNAME = "DATABASE_HOSTNAME"
	DATABASE_PORT     = "DATABASE_PORT"
	DATABASE_USER     = "DATABASE_USER"
	DATABASE_PASSWORD = "DATABASE_PASSWORD"

	DATABASE_MAX_IDLE_CONNECTIONS = "DATABASE_MAX_IDLE_CONNECTIONS"
	DATABASE_MAX_OPEN_CONNECTIONS = "DATABASE_MAX_OPEN_CONNECTIONS"
	DATABASE_MAX_CONN_LIFETIME    = "DATABASE_MAX_CONN_LIFETIME"

	ETH_CHAIN_CONFIG = "ETH_CHAIN_CONFIG"
	ETH_CHAIN_ID     = "ETH_CHAIN_ID"
	ETH_HTTP_PATH    = "ETH_HTTP_PATH"

	VALIDATE_FROM_BLOCK              = "VALIDATE_FROM_BLOCK"
	VALIDATE_TRAIL                   = "VALIDATE_TRAIL"
	VALIDATE_RETRY_INTERVAL          = "VALIDATE_RETRY_INTERVAL"
	VALIDATE_STATEDIFF_MISSING_BLOCK = "VALIDATE_STATEDIFF_MISSING_BLOCK"
	VALIDATE_STATEDIFF_TIMEOUT       = "VALIDATE_STATEDIFF_TIMEOUT"
)

// Bind env vars
func init() {
	viper.BindEnv("log.level", LOG_LEVEL)
	viper.BindEnv("log.file", LOG_FILE_PATH)

	viper.BindEnv("prom.metrics", PROM_METRICS)
	viper.BindEnv("prom.http", PROM_HTTP)
	viper.BindEnv("prom.httpAddr", PROM_HTTP_ADDR)
	viper.BindEnv("prom.httpPort", PROM_HTTP_PORT)
	viper.BindEnv("prom.dbStats", PROM_DB_STATS)

	viper.BindEnv("database.name", DATABASE_NAME)
	viper.BindEnv("database.hostname", DATABASE_HOSTNAME)
	viper.BindEnv("database.port", DATABASE_PORT)
	viper.BindEnv("database.user", DATABASE_USER)
	viper.BindEnv("database.password", DATABASE_PASSWORD)

	viper.BindEnv("database.maxIdle", DATABASE_MAX_IDLE_CONNECTIONS)
	viper.BindEnv("database.maxOpen", DATABASE_MAX_OPEN_CONNECTIONS)
	viper.BindEnv("database.maxLifetime", DATABASE_MAX_CONN_LIFETIME)

	viper.BindEnv("ethereum.chainConfig", ETH_CHAIN_CONFIG)
	viper.BindEnv("ethereum.chainID", ETH_CHAIN_ID)
	viper.BindEnv("ethereum.httpPath", ETH_HTTP_PATH)

	viper.BindEnv("validate.fromBlock", VALIDATE_FROM_BLOCK)
	viper.BindEnv("validate.trail", VALIDATE_TRAIL)
	viper.BindEnv("validate.retryInterval", VALIDATE_RETRY_INTERVAL)
	viper.BindEnv("validate.stateDiffMissingBlock", VALIDATE_STATEDIFF_MISSING_BLOCK)
	viper.BindEnv("validate.stateDiffTimeout", VALIDATE_STATEDIFF_TIMEOUT)
}
