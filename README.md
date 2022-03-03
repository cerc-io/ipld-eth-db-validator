- [Validator-README](#validator-readme)
- [Overview](#overview)
- [Intention for the Validator](#intention-for-the-validator)
  - [Edge Cases](#edge-cases)
- [Instructions for Testing](#instructions-for-testing)
- [Code Overview](#code-overview)
- [Known Bugs](#known-bugs)
- [Tests on 03/03/22](#tests-on-03-03-22)
  - [Set Up](#set-up)
  - [Testing Failures](#testing-failures)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

# Overview

This repository contains the validator. The purpose of the validator is to ensure that the data in the Core Postgres database match the data on the blockchain.

# Intention for the Validator

The perfect scenario for the validator is as follows:

1. The validator will have the capacity to perform historical checks for the Core Postgres database. Users can contain these historical checks to specified configurations (block range).
2. The validator will validate a certain number of trailing blocks, `t`, trailing the head, `n`. Therefore the validator will constantly perform real-time validation starting at `n` and ending at `n - t`.
3. The validator validates the IPLD blocks in the Core Database; it will update the core database to indicate that the validator validated it.

## Edge Cases

We must consider the following edge cases for the validator.

- There are three different data types that the validator must account for.

# Instructions for Testing

To run the test, do the following:

1. Make sure `GOPATH` is set in your `~/.bashrc` or `~/.bash_profile`: `export GOPATH=$(go env GOPATH)`
2. `./scripts/run_integration_test.sh`

# Code Overview

This section will provide some insight into specific files and their purpose.

- `validator_test/chain_maker.go` - This file contains the code for creating a “test” blockchain.
- `validator_test/validator_test.go` - This file contains testing to validate the validator. It leverages `chain_maker.go` to create a blockchain to validate.
- `pkg/validator/validator.go` - This file contains most of the core logic for the validator.

# Known Bugs

1. The validator is improperly handling missing headers from the database.
   1. Scenario
      1. The IPLD blocks from the mock blockchain are inserted into the Postgres Data.
      2. The validator runs, and all tests pass.
      3. Users manually remove the last few rows from the database.
      4. The validator runs, and all tests pass - This behavior is neither expected nor wanted.

# Tests on 03/03/22

The tests highlighted below were conducted to validate the initial behavior of the validator.

## Set Up

Below are the steps utilized to set up the test environment.

1. Run the `scripts/run_integration_test.sh` script.
   1. First comment outline 130 to 133 from `validator_test/validator_test.go`
2. Once the code has completed running, comment out lines 55 to 126, 38 to 40, and 42 to 44.
   1. Make the following change `db, err = setupDB() --> db, _ = setupDB()`
3. Run the following command: `ginkgo -r validator_test/ -v`
   1. All tests should pass

## Testing Failures

Once we had populated the database, we tested for failures.

1. Removing a Transaction from `transaction_cids` - If we removed a transaction from the database and ran the test, the test would fail. **This is the expected behavior.**
2. Removing Headers from `eth.header_cids`
   1. If we removed a header block sandwiched between two header blocks, the test would fail (For example, we removed the entry for block 4, and the block range is 1-10). **This is the expected behavior.**
   2. If we removed the tail block(s) from the table, the test would pass (For example, we remove the entry for blocks 8, 9, 10, and the block range is 1-10). **This is _not_ the expected behavior.**
