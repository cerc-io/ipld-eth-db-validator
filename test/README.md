# Test Instructions

## Setup

- For running integration tests:

  - Clone [stack-orchestrator](https://github.com/vulcanize/stack-orchestrator), [go-ethereum](https://github.com/vulcanize/go-ethereum) and [ipld-eth-db](https://github.com/vulcanize/ipld-eth-db) repositories.

  - Checkout [v4 release](https://github.com/vulcanize/ipld-eth-db/releases/tag/v4.2.0-alpha) in ipld-eth-db repo.

    ```bash
    # In ipld-eth-db repo.
    git checkout v4.2.0-alpha
    ```

  - Checkout [v4 release](https://github.com/vulcanize/go-ethereum/releases/tag/v1.10.19-statediff-4.1.0-alpha) in go-ethereum repo.

    ```bash
    # In go-ethereum repo.
    git checkout v1.10.19-statediff-4.1.0-alpha
    ```

  - Checkout working commit in stack-orchestrator repo.

    ```bash
    # In stack-orchestrator repo.
    git checkout f2fd766f5400fcb9eb47b50675d2e3b1f2753702
    ```

## Run

- Run unit tests:

  ```bash
  # In ipld-eth-db-validator root directory.
  ./scripts/run_unit_test.sh
  ```

- Run integration tests:

  - In stack-orchestrator repo:

    - Create config file:

      ```bash
      cd helper-scripts

      ./create-config.sh
      ```

      A `config.sh` will be created in the root directory.

    - Update/Edit the config file `config.sh`:

      ```bash
      #!/bin/bash

      # Path to ipld-eth-server repo.
      vulcanize_ipld_eth_db=~/ipld-eth-db/

      # Path to go-ethereum repo.
      vulcanize_go_ethereum=~/go-ethereum

      # Path to contract folder.
      vulcanize_test_contract=~/ipld-eth-db-validator/test/contract

      genesis_file_path='start-up-files/go-ethereum/genesis.json'
      db_write=true
      ```

    - Run stack-orchestrator:

      ```bash
      # In stack-orchestrator root directory.
      cd helper-scripts

      ./wrapper.sh \
      -e docker \
      -d ../docker/local/docker-compose-db-sharding.yml \
      -d ../docker/local/docker-compose-go-ethereum.yml \
      -d ../docker/local/docker-compose-contract.yml \
      -v remove \
      -p ../config.sh
      ```

  - Run tests:

    ```bash
    # In ipld-eth-db-validator root directory.
    ./scripts/run_integration_test.sh
    ```
