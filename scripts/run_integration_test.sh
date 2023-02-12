#!/bin/bash

set -ex

# export PGPASSWORD=password
# export DATABASE_USER=vdbm
# export DATABASE_PORT=8077
# export DATABASE_PASSWORD=password
# export DATABASE_HOSTNAME=127.0.0.1
# export DATABASE_NAME=cerc_testing

# # Wait for containers to be up and execute the integration test.
# while [ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:8545)" != "200" ]; do echo "waiting for geth-statediff..." && sleep 5; done && \
#         make integrationtest


# laconic-so --verbose --local-stack --stack fixturenet-eth-loaded build-containers --exclude cerc/ipld-eth-server,cerc/keycloak,cerc/tx-spammer

# laconic-so --verbose --local-stack --stack fixturenet-eth-loaded deploy --exclude tx-spammer,fixturenet-eth-metrics,keycloak,ipld-eth-server up

# ETH_PORT=$(docker-port.sh fixturenet-eth-geth-1-1 8545)
# docker run -it -p 3000:3000 -e ETH_ADDR="http://127.0.0.1:$ETH_PORT" cerc/test-contract

# Build and deploy a cluster with only what we need from the stack
laconic-so --verbose --stack fixturenet-eth-loaded setup-repositories \
    --include cerc-io/go-ethereum,cerc-io/ipld-eth-db

laconic-so --verbose --stack fixturenet-eth-loaded build-containers \
    --include cerc/fixturenet-eth-geth,cerc/fixturenet-eth-lighthouse,cerc/ipld-eth-db

laconic-so --verbose --stack fixturenet-eth-loaded deploy \
    --include fixturenet-eth,ipld-eth-db \
    --cluster test up

# echo "Waiting for geth..."
# until [[ "$(docker inspect test-fixturenet-eth-geth-1-1 | jq -r '.[0].State.Status')" = 'running' ]]
# do
#     sleep 1
# done

# docker cp test-fixturenet-eth-geth-1-1:/opt/testnet/build/el/accounts.csv .
# DEPLOYER_PK="$(head -1 accounts.csv | cut -d',' -f3)"

# Read an account key so we can send from a funded account
DEPLOYER_PK="$(docker exec test-fixturenet-eth-geth-1-1 cat /opt/testnet/build/el/accounts.csv | cut -d',' -f3)"

# Build and run the deployment server
docker build ./test/contract -t cerc/test-contract

docker run -d --rm -i -p 3000:3000 --network test_default --name=test-deployer \
    -e ETH_ADDR=http://fixturenet-eth-geth-1:8545 \
    -e ETH_CHAIN_ID=1212 \
    -e DEPLOYER_PRIVATE_KEY=$DEPLOYER_PK \
    cerc/test-contract

go test -v ./integration
