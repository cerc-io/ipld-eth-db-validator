# Overview

This repository contains the validator. The purpose of the validator is to ensure that the data in the Core Postgres database match the data on the blockchain.

# Intention for the Validator

The perfect scenario for the validator is as follows:

1. The validator will have the capacity to perform historical checks for the Core Postgres database. Users can contain these historical checks to specified configurations (block range).
2. The validator will perform validation for a certain number of trailing blocks, `t`, trailing the head, `n`. Therefore the validator will constantly perform real-time validation starting at `n` and ending at `n - t`.
3. The validator validates IDLP blocks in the Core Database; it will update the core database to indicate that the validator validated the block.

## Edge Cases

We must consider the following edge cases for the validator.

- There are three different data types that the validator must account for.

# Instructions for Testing

To run the test, do the following:

1. Make sure `GOPATH` is set in your `~/.bashrc` or `~/.bash_profile`: `export GOPATH=$(go env GOPATH)`
2. `./scripts/run_integration_test.sh`
