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

	"github.com/jmoiron/sqlx"
)

var (
	ReferentialIntegrityErr = "referential integrity check failed at block %d, entry for %s not found"
	EntryNotFoundErr        = "entry for %s not found"
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

	err = ValidateLogCIDsRef(db, blockNumber)
	if err != nil {
		return err
	}

	return nil
}

// ValidateHeaderCIDsRef does a reference integrity check on references in eth.header_cids table
func ValidateHeaderCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	err := ValidateIPFSBlocks(db, blockNumber, "eth.header_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateUncleCIDsRef does a reference integrity check on references in eth.uncle_cids table
func ValidateUncleCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var exists bool
	err := db.Get(&exists, UncleCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.uncle_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateTransactionCIDsRef does a reference integrity check on references in eth.header_cids table
func ValidateTransactionCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var exists bool
	err := db.Get(&exists, TransactionCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.transaction_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateReceiptCIDsRef does a reference integrity check on references in eth.receipt_cids table
func ValidateReceiptCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var exists bool
	err := db.Get(&exists, ReceiptCIDsRefTransactionCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.transaction_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.receipt_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStateCIDsRef does a reference integrity check on references in eth.state_cids table
func ValidateStateCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var exists bool
	err := db.Get(&exists, StateCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.state_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStorageCIDsRef does a reference integrity check on references in eth.storage_cids table
func ValidateStorageCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var exists bool
	err := db.Get(&exists, StorageCIDsRefStateCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.state_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.storage_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateLogCIDsRef does a reference integrity check on references in eth.log_cids table
func ValidateLogCIDsRef(db *sqlx.DB, blockNumber uint64) error {
	var exists bool
	err := db.Get(&exists, LogCIDsRefReceiptCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.receipt_cids")
	}

	err = ValidateIPFSBlocks(db, blockNumber, "eth.log_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateIPFSBlocks does a reference integrity check between the given CID table and IPFS blocks table on MHKey and block number
func ValidateIPFSBlocks(db *sqlx.DB, blockNumber uint64, CIDTable string, CIDField string) error {
	var exists bool
	err := db.Get(&exists, fmt.Sprintf(CIDsRefIPLDBlocks, CIDTable, CIDField), blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "ipld.blocks")
	}

	return nil
}
