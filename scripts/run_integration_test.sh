#!/bin/bash

set -e
set -o xtrace

export PGPASSWORD=password
export DATABASE_USER=vdbm
export DATABASE_PORT=8066
export DATABASE_PASSWORD=password
export DATABASE_HOSTNAME=127.0.0.1
export DATABASE_NAME=vulcanize_testing_v4

# Wait for containers to be up and execute the integration test.
while [ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:8545)" != "200" ]; do echo "waiting for geth-statediff..." && sleep 5; done && \
        make integrationtest
