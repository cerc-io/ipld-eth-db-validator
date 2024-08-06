package validator_test

import (
	"context"
	"math/big"

	"github.com/cerc-io/plugeth-statediff/indexer/database/sql/postgres"
	"github.com/ethereum/go-ethereum/params"
	"github.com/jmoiron/sqlx"
)

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
	GrayGlacierBlock:    big.NewInt(0),
	Ethash:              new(params.EthashConfig),
}

var TestDBConfig postgres.Config

func init() {
	var err error
	TestDBConfig, err = postgres.TestConfig.WithEnv()
	if err != nil {
		panic(err)
	}
}

func SetupDB() *sqlx.DB {
	db, err := postgres.ConnectSQLX(context.Background(), TestDBConfig)
	if err != nil {
		panic(err)
	}
	return db
}
