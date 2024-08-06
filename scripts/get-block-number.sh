#!/bin/bash

set -eu

geth_endpoint="${1:-$ETH_HTTP_PATH}"

latest_block_hex=$(curl -s $geth_endpoint -X POST -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":42}' | \
    python3 -c 'import json, sys; print(int(json.load(sys.stdin)["result"], 16))' \
)

printf "%d" $latest_block_hex
