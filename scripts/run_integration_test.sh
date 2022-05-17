set -e
set -o xtrace

export ETH_FORWARD_ETH_CALLS=false
export DB_WRITE=true
export ETH_HTTP_PATH=""
export ETH_PROXY_ON_ERROR=false

# Clear up existing docker images and volume.
docker-compose down --remove-orphans --volumes

# Build and start the containers.
docker-compose -f docker-compose.yml  up -d ipld-eth-db dapptools contract

export PGPASSWORD=password
export DATABASE_USER=vdbm
export DATABASE_PORT=8077
export DATABASE_PASSWORD=password
export DATABASE_HOSTNAME=127.0.0.1
export DATABASE_NAME=vulcanize_testing

# Wait for containers to be up and execute the integration test.
while [ "$(curl -s -o /dev/null -w ''%{http_code}'' localhost:8545)" != "200" ]; do echo "waiting for geth-statediff..." && sleep 5; done && \
        make integrationtest
