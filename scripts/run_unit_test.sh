#!/bin/bash

set -e

# Clear up existing docker images and volume.
docker-compose down --remove-orphans --volumes

# Spin up TimescaleDB
docker-compose -f docker-compose.yml up -d migrations ipld-eth-db
sleep 45

# Run unit tests
go clean -testcache
PGPASSWORD=password DATABASE_USER=vdbm DATABASE_PORT=8077 DATABASE_PASSWORD=password DATABASE_HOSTNAME=127.0.0.1 DATABASE_NAME=vulcanize_testing make test

# Clean up
docker-compose down --remove-orphans --volumes
