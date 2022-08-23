# ipld-eth-db-validator

> `ipld-eth-db-validator` performs validation checks on indexed Ethereum IPLD objects in a Postgres database:
> * Attempt to apply transactions in each block and validate resultant block hash
> * Check referential integrity between IPLD blocks and index tables

## Setup

Build the binary:

```bash
make build
```

## Configuration

An example config file:

```toml
[database]
  # db credentials
  name     = "vulcanize_public" # DATABASE_NAME
  hostname = "localhost"        # DATABASE_HOSTNAME
  port     = 5432               # DATABASE_PORT
  user     = "vdbm"             # DATABASE_USER
  password = "..."              # DATABASE_PASSWORD

[validate]
  # block height to initiate database validation at
  blockHeight = 1      # VALIDATE_BLOCK_HEIGHT  (default: 1)
  # number of blocks to trail behind the head
  trail  = 16         # VALIDATE_TRAIL  (default: 16)
  # sleep interval after validator has caught up to (head-trail) height (in sec)
  sleepInterval = 10  # VALIDATE_SLEEP_INTERVAL (default: 10)

  # whether to perform a statediffing call on a missing block
  stateDiffMissingBlock = true # (default: false)
  # statediffing call timeout period (in sec)
  stateDiffTimeout = 240 # (default: 240)

[ethereum]
  # node info
  # path to json chain config (optional)
  chainConfig = ""            # ETH_CHAIN_CONFIG
  # eth chain id for config (overridden by chainConfig)
  chainID = "1"               # ETH_CHAIN_ID (default: 1)
  # http RPC endpoint URL for a statediffing node
  httpPath = "localhost:8545" # ETH_HTTP_PATH

[prom]
  # prometheus metrics
  metrics = true        # PROM_METRICS    (default: false)
  http = true           # PROM_HTTP       (default: false)
  httpAddr = "0.0.0.0"  # PROM_HTTP_ADDR  (default: 127.0.0.1)
  httpPort = "9001"     # PROM_HTTP_PORT  (default: 9001)
  dbStats = true        # PROM_DB_STATS   (default: false)

[log]
  # log level (trace, debug, info, warn, error, fatal, panic)
  level = "info"  # LOG_LEVEL (default: info)
  # file path for logging, leave unset to log to stdout
  file  = ""      # LOG_FILE_PATH
```


* The validation process trails behind the latest block number in the database by config parameter `validate.trail`.

* If the validator has caught up to (head-trail) height, it waits for a configured time interval (`validate.sleepInterval`) before again querying the database.

* If the validator encounters a missing block (gap) in the database, it makes a `writeStateDiffAt` call to the configured statediffing endpoint (`ethereum.httpPath`) if `validate.stateDiffMissingBlock` is set to `true`. Here it is assumed that the statediffing node pointed to is writing out to the database.

### Local Setup

* Create a chain config file `chain.json` according to chain config in genesis json file used by local geth.

  Example:

  ```json
  {
    "chainId": 41337,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "clique": {
      "period": 5,
      "epoch": 30000
    }
  }
  ```

  Provide the path to the above file in the config.

## Usage

* Create / update the config file (refer to example config above).

* Run validator:

  ```bash
  ./ipld-eth-db-validator stateValidator --config=<config path>
  ```

  Example:

  ```bash
  ./ipld-eth-db-validator stateValidator --config=environments/example.toml
  ```

## Monitoring

* Enable metrics using config parameters `prom.metrics` and `prom.http`.
* `ipld-eth-db-validator` exposes following prometheus metrics at `/metrics` endpoint:
  * `last_validated_block`: Last validated block number.
  * DB stats if `prom.dbStats` set to `true`.

## Tests

* Follow [Test Instructions](./test/README.md) to run unit and integration tests locally.
