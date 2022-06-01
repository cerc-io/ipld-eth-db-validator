package validator

// Queries to validate referential integrity in the indexed data:
// At the given block number,
// In each table, for each (would be) foreign key reference, perform left join with the referenced table on the foreign key fields.
// Select rows where there are no matching rows in the referenced table.
// If any such rows exist, there are missing entries in the referenced table.

const (
	CIDsRefIPLDBlocks = `SELECT EXISTS (
						SELECT *
						FROM %[1]s
						LEFT JOIN public.blocks ON (
							%[1]s.%[2]s = blocks.key
							AND %[1]s.block_number = blocks.block_number
						)
						WHERE
							%[1]s.block_number = $1
							AND blocks.key IS NULL
					)`

	UncleCIDsRefHeaderCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.uncle_cids
						LEFT JOIN eth.header_cids ON (
							uncle_cids.header_id = header_cids.block_hash
							AND uncle_cids.block_number = header_cids.block_number
						)
						WHERE
							uncle_cids.block_number = $1
							AND header_cids.block_hash IS NULL
					)`

	TransactionCIDsRefHeaderCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.transaction_cids
						LEFT JOIN eth.header_cids ON (
							transaction_cids.header_id = header_cids.block_hash
							AND transaction_cids.block_number = header_cids.block_number
						)
						WHERE
							transaction_cids.block_number = $1
							AND header_cids.block_hash IS NULL
					)`

	ReceiptCIDsRefTransactionCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.receipt_cids
						LEFT JOIN eth.transaction_cids ON (
							receipt_cids.tx_id = transaction_cids.tx_hash
							AND receipt_cids.block_number = transaction_cids.block_number
						)
						WHERE
							receipt_cids.block_number = $1
							AND transaction_cids.tx_hash IS NULL
					)`

	StateCIDsRefHeaderCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.state_cids
						LEFT JOIN eth.header_cids ON (
							state_cids.header_id = header_cids.block_hash
							AND state_cids.block_number = header_cids.block_number
						)
						WHERE
							state_cids.block_number = $1
							AND header_cids.block_hash IS NULL
					)`

	StorageCIDsRefStateCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.storage_cids
						LEFT JOIN eth.state_cids ON (
							storage_cids.state_path = state_cids.state_path
							AND storage_cids.header_id = state_cids.header_id
							AND storage_cids.block_number = state_cids.block_number
						)
						WHERE
							storage_cids.block_number = $1
							AND state_cids.state_path IS NULL
					)`

	StateAccountsRefStateCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.state_accounts
						LEFT JOIN eth.state_cids ON (
							state_accounts.state_path = state_cids.state_path
							AND state_accounts.header_id = state_cids.header_id
							AND state_accounts.block_number = state_cids.block_number
						)
						WHERE
							state_accounts.block_number = $1
							AND state_cids.state_path IS NULL
					)`

	AccessListElementsRefTransactionCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.access_list_elements
						LEFT JOIN eth.transaction_cids ON (
							access_list_elements.tx_id = transaction_cids.tx_hash
							AND access_list_elements.block_number = transaction_cids.block_number
						)
						WHERE
							access_list_elements.block_number = $1
							AND transaction_cids.tx_hash IS NULL
					)`

	LogCIDsRefReceiptCIDs = `SELECT EXISTS (
						SELECT *
						FROM eth.log_cids
						LEFT JOIN eth.receipt_cids ON (
							log_cids.rct_id = receipt_cids.tx_id
							AND log_cids.block_number = receipt_cids.block_number
						)
						WHERE
							log_cids.block_number = $1
							AND receipt_cids.tx_id IS NULL
					)`
)
