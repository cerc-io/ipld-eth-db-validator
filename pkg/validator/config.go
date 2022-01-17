package validator

import (
	"fmt"

	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"
	"github.com/spf13/viper"
)

type Config struct {
	dbParams postgres.ConnectionParams
	dbConfig postgres.ConnectionConfig
	DB       *postgres.DB
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	return cfg, cfg.setupDB()
}

func (c *Config) setupDB() error {
	_ = viper.BindEnv("database.name", postgres.DATABASE_NAME)
	_ = viper.BindEnv("database.hostname", postgres.DATABASE_HOSTNAME)
	_ = viper.BindEnv("database.port", postgres.DATABASE_PORT)
	_ = viper.BindEnv("database.user", postgres.DATABASE_USER)
	_ = viper.BindEnv("database.password", postgres.DATABASE_PASSWORD)
	_ = viper.BindEnv("database.maxIdle", postgres.DATABASE_MAX_IDLE_CONNECTIONS)
	_ = viper.BindEnv("database.maxOpen", postgres.DATABASE_MAX_OPEN_CONNECTIONS)
	_ = viper.BindEnv("database.maxLifetime", postgres.DATABASE_MAX_CONN_LIFETIME)

	// DB params
	c.dbParams.Name = viper.GetString("database.name")
	c.dbParams.Hostname = viper.GetString("database.hostname")
	c.dbParams.Port = viper.GetInt("database.port")
	c.dbParams.User = viper.GetString("database.user")
	c.dbParams.Password = viper.GetString("database.password")

	// DB Config
	c.dbConfig.MaxIdle = viper.GetInt("database.maxIdle")
	c.dbConfig.MaxOpen = viper.GetInt("database.maxOpen")
	c.dbConfig.MaxLifetime = viper.GetInt("database.maxLifetime")

	// Create DB
	db, err := NewDB(postgres.DbConnectionString(c.dbParams), postgres.ConnectionConfig{}, node.Info{})
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	c.DB = db
	return nil
}
