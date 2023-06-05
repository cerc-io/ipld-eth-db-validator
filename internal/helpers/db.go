package helpers

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/jmoiron/sqlx"
)

var TestDBConfig, _ = postgres.TestConfig.WithEnv()

// SetupDB is use to setup a db for watcher tests
func SetupDB() *sqlx.DB {
	db, err := postgres.ConnectSQLX(context.Background(), TestDBConfig)
	if err != nil {
		panic(err)
	}
	return db
}

// TearDownDB is used to tear down the watcher dbs after tests
func TearDownDB(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	statements := []string{
		`TRUNCATE nodes`,
		`TRUNCATE ipld.blocks`,
		`TRUNCATE eth.header_cids`,
		`TRUNCATE eth.uncle_cids`,
		`TRUNCATE eth.transaction_cids`,
		`TRUNCATE eth.receipt_cids`,
		`TRUNCATE eth.state_cids`,
		`TRUNCATE eth.storage_cids`,
		`TRUNCATE eth.log_cids`,
		`TRUNCATE eth_meta.watched_addresses`,
	}
	for _, stm := range statements {
		if _, err = tx.Exec(stm); err != nil {
			return fmt.Errorf("error executing `%s`: %w", stm, err)
		}
	}
	return tx.Commit()
}
