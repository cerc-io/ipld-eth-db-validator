# Test Insructions

## Setup

- For running integration tests:

  - Clone [stack-orchestrator](https://github.com/vulcanize/stack-orchestrator) and [go-ethereum](https://github.com/vulcanize/go-ethereum) repositories.

  - Checkout [v3 release](https://github.com/vulcanize/go-ethereum/releases/tag/v1.10.17-statediff-3.2.1) in go-ethereum repo.

    ```bash
    # In go-ethereum repo.
    git checkout v1.10.17-statediff-3.2.1
    ```

  - Checkout working commit in stack-orchestrator repo.

    ```bash
    # In stack-orchestrator repo.
    git checkout 3bb1796a59827fb755410c5ce69fac567a0f832b
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
