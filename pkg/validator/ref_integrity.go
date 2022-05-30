package validator

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ValidateReferentialIntegrity validates referential integrity at the given height
func ValidateReferentialIntegrity(db *sqlx.DB, blockNumber uint64) error {

	err := ValidateHeaderCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateUncleCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateTransactionCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateReceiptCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateStateCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateStorageCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateStateAccountsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateAccessListElementsRef(db, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateLogCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	return nil
}

// ValidateHeaderCIDsRef does a reference integrity check on references in eth.header_cids table
func ValidateHeaderCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	err := ValidateIPFSBlocks(db, blockNumber, "eth.header_cids", "mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateUncleCIDsRef does a reference integrity check on references in eth.uncle_cids table
func ValidateUncleCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, UncleCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.uncle_cids", "mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateTransactionCIDsRef does a reference integrity check on references in eth.header_cids table
func ValidateTransactionCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, TransactionCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.transaction_cids", "mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateReceiptCIDsRef does a reference integrity check on references in eth.receipt_cids table
func ValidateReceiptCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, ReceiptCIDsRefTransactionCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.transaction_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.receipt_cids", "leaf_mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStateCIDsRef does a reference integrity check on references in eth.state_cids table
func ValidateStateCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, StateCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.state_cids", "mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStorageCIDsRef does a reference integrity check on references in eth.storage_cids table
func ValidateStorageCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, StorageCIDsRefStateCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.state_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.storage_cids", "mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStateAccountsRef does a reference integrity check on references in eth.state_accounts table
func ValidateStateAccountsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, StateAccountsRefStateCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.state_cids")
	}

	return nil
}

// ValidateAccessListElementsRef does a reference integrity check on references in eth.access_list_elements table
func ValidateAccessListElementsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, AccessListElementsRefTransactionCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.transaction_cids")
	}

	return nil
}

// ValidateLogCIDsRef does a reference integrity check on references in eth.log_cids table
func ValidateLogCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var count int
	err := db.Get(&count, LogCIDsRefReceiptCIDs, blockNumber)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.receipt_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.log_cids", "leaf_mh_key")
	if err != nil {
		return err
	}

	return nil
}

// ValidateIPFSBlocks does a reference integrity check between the given CID table and IPFS blocks table on MHKey and block number
func ValidateIPFSBlocks(db *sqlx.DB, blockNumber uint64, CIDTable string, mhKeyField string) error {
	var count int
	err := db.Get(&count, fmt.Sprintf(CIDsRefIPLDBlocks, CIDTable, mhKeyField), blockNumber)
	if err != nil {
		return err
	}

	if count != 0 {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "public.blocks")
	}

	return nil
}
