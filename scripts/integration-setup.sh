#!/bin/bash

set -ex

# Prevent conflicting tty output
export BUILDKIT_PROGRESS=plain
# By default assume we are running in the project root
export CERC_REPO_BASE_DIR="${CERC_REPO_BASE_DIR:-..}"

CONFIG_DIR=$(readlink -f "${CONFIG_DIR:-$(mktemp -d)}")

laconic_so="${LACONIC_SO:-laconic-so} --stack fixturenet-eth-loaded --quiet"

set -x

# Build and deploy a cluster with only what we need from the stack
$laconic_so setup-repositories \
    --exclude cerc-io/ipld-eth-server,cerc-io/tx-spammer \
    --branches-file ./test/stack-refs.yml

$laconic_so build-containers \
    --exclude cerc/ipld-eth-server,cerc/keycloak,cerc/tx-spammer

$laconic_so deploy \
    --include fixturenet-eth,ipld-eth-db \
    --cluster test up

set +x

# Get IPv4 endpoint of geth file server
bootnode_endpoint=$(docker port test-fixturenet-eth-bootnode-geth-1 9898 | head -1)
geth_endpoint="$(docker port test-fixturenet-eth-geth-1-1 8545 | head -1)"

# Extract the chain config and ID from genesis file
curl -s $bootnode_endpoint/geth.json | jq '.config' > "$CONFIG_DIR/chain.json"

# export PGPASSWORD=password
# QUERY_BLOCKS_EXIST='SELECT exists(SELECT block_number FROM ipld.blocks LIMIT 1);'

# echo "Waiting until we have some data written..."
# until [[ "$(psql -qtA cerc_testing -h localhost -U vdbm -p 8077 -c "$QUERY_BLOCKS_EXIST")" = 't' ]]; do
#     sleep 1
# done

# Output vars if we are running on Github
if [[ -n "$GITHUB_ENV" ]]; then
    echo ETH_CHAIN_ID="$(jq '.chainId' $CONFIG_DIR/chain.json)" >> "$GITHUB_ENV"
    echo ETH_CHAIN_CONFIG="$CONFIG_DIR/chain.json" >> "$GITHUB_ENV"
    echo ETH_HTTP_PATH=$geth_endpoint >> "$GITHUB_ENV"
    # Read a private key so we can send from a funded account
    echo DEPLOYER_PRIVATE_KEY="$(curl -s $bootnode_endpoint/accounts.csv | head -1 | cut -d',' -f3)" >> "$GITHUB_ENV"
fi
