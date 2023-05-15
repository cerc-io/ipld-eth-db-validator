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
		`DELETE FROM nodes`,
		`DELETE FROM ipld.blocks`,
		`DELETE FROM eth.header_cids`,
		`DELETE FROM eth.uncle_cids`,
		`DELETE FROM eth.transaction_cids`,
		`DELETE FROM eth.receipt_cids`,
		`DELETE FROM eth.state_cids`,
		`DELETE FROM eth.storage_cids`,
		`DELETE FROM eth.log_cids`,
		`DELETE FROM eth_meta.watched_addresses`,
	}
	for _, stm := range statements {
		if _, err = tx.Exec(stm); err != nil {
			return fmt.Errorf("error executing `%s`: %w", stm, err)
		}
	}
	return tx.Commit()
}
