package validator

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/vulcanize/ipld-eth-server/v3/pkg/shared"
)

var (
	DATABASE_NAME                 = "DATABASE_NAME"
	DATABASE_HOSTNAME             = "DATABASE_HOSTNAME"
	DATABASE_PORT                 = "DATABASE_PORT"
	DATABASE_USER                 = "DATABASE_USER"
	DATABASE_PASSWORD             = "DATABASE_PASSWORD"
	DATABASE_MAX_IDLE_CONNECTIONS = "DATABASE_MAX_IDLE_CONNECTIONS"
	DATABASE_MAX_OPEN_CONNECTIONS = "DATABASE_MAX_OPEN_CONNECTIONS"
	DATABASE_MAX_CONN_LIFETIME    = "DATABASE_MAX_CONN_LIFETIME"
)

var IntegrationTestChainConfig = &params.ChainConfig{
	ChainID:             big.NewInt(4),
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
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	return cfg, cfg.setupDB()
}

func (c *Config) setupDB() error {
	_ = viper.BindEnv("database.name", DATABASE_NAME)
	_ = viper.BindEnv("database.hostname", DATABASE_HOSTNAME)
	_ = viper.BindEnv("database.port", DATABASE_PORT)
	_ = viper.BindEnv("database.user", DATABASE_USER)
	_ = viper.BindEnv("database.password", DATABASE_PASSWORD)
	_ = viper.BindEnv("database.maxIdle", DATABASE_MAX_IDLE_CONNECTIONS)
	_ = viper.BindEnv("database.maxOpen", DATABASE_MAX_OPEN_CONNECTIONS)
	_ = viper.BindEnv("database.maxLifetime", DATABASE_MAX_CONN_LIFETIME)

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
	return nil
}
