# Test Instructions

## Setup

- For running integration tests:

  - Clone [stack-orchestrator](https://github.com/vulcanize/stack-orchestrator) and [go-ethereum](https://github.com/vulcanize/go-ethereum) repositories.

  - Checkout [v4 release](https://github.com/vulcanize/go-ethereum/releases/tag/v1.10.18-statediff-4.0.2-alpha) in go-ethereum repo.

    ```bash
    # In go-ethereum repo.
    git checkout v1.10.18-statediff-4.0.2-alpha
    ```

  - Checkout working commit in stack-orchestrator repo.

    ```bash
    # In stack-orchestrator repo.
    git checkout 418957a1f745c921b21286c13bb033f922a91ae9
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

      # Path to go-ethereum repo.
      vulcanize_go_ethereum=~/go-ethereum

      # Path to contract folder.
      vulcanize_test_contract=~/ipld-eth-db-validator/test/contract

      db_write=true
      ipld_eth_server_db_dependency=access-node
      go_ethereum_db_dependency=access-node

      connecting_db_name=vulcanize_testing_v4
      ```

    - Run stack-orchestrator:

      ```bash
      # In stack-orchestrator root directory.
      cd helper-scripts

      ./wrapper.sh \
      -e docker \
      -d ../docker/latest/docker-compose-db.yml \
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
