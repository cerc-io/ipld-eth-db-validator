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
func ValidateReferentialIntegrity(tx *sqlx.Tx, blockNumber uint64) error {
	err := ValidateHeaderCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateUncleCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateTransactionCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateReceiptCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateStateCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateStorageCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	err = ValidateLogCIDsRef(tx, blockNumber)
	if err != nil {
		return err
	}

	return nil
}

// ValidateHeaderCIDsRef does a reference integrity check on references in eth.header_cids table
func ValidateHeaderCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	err := ValidateIPFSBlocks(tx, blockNumber, "eth.header_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateUncleCIDsRef does a reference integrity check on references in eth.uncle_cids table
func ValidateUncleCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	var exists bool
	err := tx.Get(&exists, UncleCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(tx, blockNumber, "eth.uncle_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateTransactionCIDsRef does a reference integrity check on references in eth.header_cids table
func ValidateTransactionCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	var exists bool
	err := tx.Get(&exists, TransactionCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(tx, blockNumber, "eth.transaction_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateReceiptCIDsRef does a reference integrity check on references in eth.receipt_cids table
func ValidateReceiptCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	var exists bool
	err := tx.Get(&exists, ReceiptCIDsRefTransactionCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.transaction_cids")
	}

	err = ValidateIPFSBlocks(tx, blockNumber, "eth.receipt_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStateCIDsRef does a reference integrity check on references in eth.state_cids table
func ValidateStateCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	var exists bool
	err := tx.Get(&exists, StateCIDsRefHeaderCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.header_cids")
	}

	err = ValidateIPFSBlocks(tx, blockNumber, "eth.state_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateStorageCIDsRef does a reference integrity check on references in eth.storage_cids table
func ValidateStorageCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	var exists bool
	err := tx.Get(&exists, StorageCIDsRefStateCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.state_cids")
	}

	err = ValidateIPFSBlocks(tx, blockNumber, "eth.storage_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateLogCIDsRef does a reference integrity check on references in eth.log_cids table
func ValidateLogCIDsRef(tx *sqlx.Tx, blockNumber uint64) error {
	var exists bool
	err := tx.Get(&exists, LogCIDsRefReceiptCIDs, blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "eth.receipt_cids")
	}

	err = ValidateIPFSBlocks(tx, blockNumber, "eth.log_cids", "cid")
	if err != nil {
		return err
	}

	return nil
}

// ValidateIPFSBlocks does a reference integrity check between the given CID table and IPFS blocks table on MHKey and block number
func ValidateIPFSBlocks(tx *sqlx.Tx, blockNumber uint64, CIDTable string, CIDField string) error {
	var exists bool
	err := tx.Get(&exists, fmt.Sprintf(CIDsRefIPLDBlocks, CIDTable, CIDField), blockNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(ReferentialIntegrityErr, blockNumber, "ipld.blocks")
	}

	return nil
}
