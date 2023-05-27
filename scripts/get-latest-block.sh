#!/bin/bash

set -eu

geth_endpoint="$1"

latest_block_hex=$(curl -s $geth_endpoint -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}' | \
    jq -r .result)

printf "%d" $latest_block_hex
