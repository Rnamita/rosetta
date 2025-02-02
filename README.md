# Celo Rosetta

[![CircleCI](https://circleci.com/gh/celo-org/rosetta/tree/master.svg?style=shield)](https://circleci.com/gh/celo-org/rosetta/tree/master)
[![License](https://img.shields.io/github/license/celo-org/rosetta.svg)](https://github.com/celo-org/rosetta/blob/master/LICENSE.txt)

A monitoring server for celo-blockchain

## What is Celo Rosetta?

Celo Rosetta is an RPC server that exposes an API to:

- Query Celo's Blockchain
- Obtain Balance Changing Operations
- Construct Airgapped Transactions

With a special focus on getting balance change operations, Celo Rosetta provides an easy way to obtain changes that are not easily queryable using
the celo-blockchain rpc; such as:

- Gas Fee distribution
- Gold transfers (internal & external). Taking in account Tobin Tax
- Epoch Rewards Distribution
- LockedGold & Election Operations

## RPC endpoints

Rosetta exposes the following endpoints:

- `POST /network/list`: Get List of Available Networks
- `POST /network/status`: Get Network Status
- `POST /network/options`: Get Network Options
- `POST /block`: Get a Block
- `POST /block/transaction`: Get a Block Transaction
- `POST /mempool`: Get All Mempool Transactions
- `POST /mempool/transaction`: Get a Mempool Transaction
- `POST /account/balance`: Get an Account Balance
- `POST /construction/metadata`: Get Transaction Construction Metadata
- `POST /construction/submit`: Submit a Signed Transaction

For an understanding of inputs & outputs check [servicer.go](./service/rpc/servicer.go)

## Command line arguments

The main command is `rosetta run`, whose arguments are:

```txt
Usage:
  rosetta run [flags]

Flags:
      --datadir string            datadir to use
      --geth.binary string        Path to the celo-blockchain binary
      --geth.bootnodes string     Bootnodes to use (separated by ,)
      --geth.cache string         Memory (in MB) allocated to geth's internal caching
      --geth.gcmode string        Geth garbage collection mode (full, archive) (default "full")
      --geth.genesis string       (Optional) path to the genesis.json, for use with custom chains
      --geth.ipcpath string       Path to the geth ipc file
      --geth.logfile string       Path to logs file
      --geth.maxpeers string      Maximum number of network peers (network disabled if set to 0) (default: 1100) (default "1100")
      --geth.network string       Network to use, either 'mainnet', 'alfajores', or 'baklava'
      --geth.publicip string      Public Ip to configure geth (sometimes required for discovery)
      --geth.rpcaddr string       Geth HTTP-RPC server listening interface (default "127.0.0.1")
      --geth.rpcport string       Geth HTTP-RPC server listening port (default "8545")
      --geth.rpcvhosts string     Geth comma separated list of virtual hostnames from which to accept requests (default "localhost")
      --geth.staticnodes string   StaticNode to use (separated by ,)
      --geth.syncmode string      Geth blockchain sync mode (fast, full, light) (default "fast")
      --geth.verbosity string     Geth log verbosity (number between [1-5])
  -h, --help                      help for run
      --rpc.address string        Listening address for http server
      --rpc.port uint             Listening port for http server (default 8080)
      --rpc.reqTimeout duration   Timeout for requests to this service, this also controls the timeout sent to the blockchain node for trace transaction requests (default 2m0s)
```

Every argument can be defined using environment variables using `ROSETTA_` prefix; and replacing `.` for `_`; for example:

```sh
ROSETTA_DATADIR="/my/dir"
ROSETTA_GETH_NETWORK="alfajores"
```

Note that from Rosetta `v0.8.4` onwards, it is no longer necessary to pass in either `--geth.bootnodes` or `--geth.staticnodes`, as the geth flag `--alfajores`, `--baklava`, or no flag (for mainnet) will be set automatically, which sets the geth bootnodes appropriately. These flags may still optionally be used but are not recommended if there is not a specific reason to do so.

## Running the Rosetta RPC Server

Running the Rosetta RPC Server from scratch will take some time to sync, since it runs a full archive node in the background. While it may be possible to run the Construction API in the future with a non-archive node, this is still required by the Rosetta spec for the Data API implementation in order to perform balance reconciliation.

### Version 1: Running from `rosetta` source code

You will need the following three repositories cloned locally:

- `rosetta` (this repo)
- [`celo-blockchain`](https://github.com/celo-org/celo-blockchain)

You also need the following dependencies to be met:

- `go >= 1.15`
- `golangci` ([installation instructions](https://golangci-lint.run/usage/install/#local-installation)) (linter dependency for the Makefile)

#### Running Rosetta

Prerequisites:

- Checkout the rosetta version that you want and run `make all`.
- Find the `celo-blockchain` version in the rosetta `go.mod` file. Look for a line containing `github.com/celo-org/celo-blockchain` the version comes after separated by a space.
- Checkout `celo-blockchain` at the version specified in rosetta's `go.mod` and run `make geth`
- Replace `<NETWORK>` with one of `alfajores` (developer testnet), `baklava` (validator testnet) or `mainnet`.
- Replace `<PATH-TO-DATADIR>` below, which is the location for the celo-blockchain data directory (the directory does not need to exist before passing it in).
The data directory is network specific so when switching networks you will also need to change the data directory.

Then run:

```bash
go run main.go run \
  --geth.network <NETWORK> \
  --geth.binary ../celo-blockchain/build/bin/geth \
  --geth.syncmode full \
  --geth.gcmode archive \
  --datadir <PATH-TO-DATADIR>
```

You should start to see continuous output looking something like this:

```sh
INFO [01-28|14:09:03.434] Press CTRL-C to stop the process
--nousb --rpc --rpcaddr 127.0.0.1 --rpcport 8545 --rpcvhosts localhost --syncmode full --gcmode archive --rpcapi eth,net,web3,debug,admin,personal --ipcpath <YourPathToRosetta>/rosetta/envs/alfajores/celo/geth.ipc --light.serve 0 --light.maxpeers 0 --maxpeers 1100 --consoleformat term
INFO [01-28|14:09:05.110] Detected Chain Parameters                chainId=44787 epochSize=17280
INFO [01-28|14:09:05.120] Starting httpServer                      listen_address=:8080
INFO [01-28|14:09:05.120] Resuming operation from last persisted  block srv=celo-monitor block=0
INFO [01-28|14:09:05.121] SubscriptionFetchMode:Start              srv=celo-monitor pipe=header_listener start=1

...

INFO [01-28|14:09:25.731] Stored 1000 blocks                       srv=celo-monitor pipe=persister       block=1000 registryUpdates=0
```

You can stop the service and restart by re-running just the last command above (`go run main.go` ... )

### Version 2: Running Rosetta Docker Image

Prerequisites:

- [Install](https://docs.docker.com/engine/install/) and run `docker` (tested with version `19.03.12`)

Rosetta is released as a docker image: `us.gcr.io/celo-testnet/rosetta`. All versions can be found on the [registry page](https://us.gcr.io/celo-testnet/rosetta). Within the docker image, we pack the `rosetta` binary and also the `geth` binary from `celo-blockchain`. Rosetta will run both.

The command below runs the Celo Rosetta RPC server for `alfajores`:

```bash
export RELEASE="latest"  # or specify a release version
# folder for rosetta to use as data directory (saves rosetta.db & celo-blockchain datadir)
export DATADIR="${PWD}/datadir"
mkdir $DATADIR
docker pull us.gcr.io/celo-testnet/rosetta:$RELEASE
docker run --name rosetta --rm \
  -v "${DATADIR}:/data" \
  -p 8080:8080 \
  us.gcr.io/celo-testnet/rosetta:$RELEASE \
  run \
  --geth.network alfajores \
  --geth.syncmode full \
  --geth.gcmode archive

```

To run this for a different network, change the `geth.network` flag from `alfajores` to `mainnet` or `baklava`.

## Airgap Client Guide

The Celo Rosetta Airgap module is designed to facilitate signing transactions, parameterized by contemporaenous network metadata, in an offline context.

Examples of this metadata include:

- network wide state like "gas price minimum"
- argument specific state like vote amount "effect on validator priority queue"

```js
AirGapServer {
  ObtainMetadata(TxArgs): TxMetadata
  SubmitTx(Tx): Status
}

AirGapClient {
  ConstructTxFromMetadata(TxMetadata): Tx
  SignTx(Tx, PrivateKey): Tx
}
```

### Custody: Staking and Voting

For a documentation resource, please see the [custody docs](https://docs.celo.org/developer-guide/overview/integrations/custody).

For a code resource, please see the [examples](./examples/airgap/main.go).

## Developer Guide

### Setup

In addition to the dependencies listed above under the instructions for running from `rosetta` source code, you also need:

- `openapi-generator` To re-generate rpc scaffold ([install link](https://openapi-generator.tech))

### Build Commands

Important commands:

- `make all`: Builds project (compiles all modules), same as `go build ./...`
- `make test` or `go test ./...` to run unit tests

### Interaction with Celo Core Contracts

Rosetta uses [kliento](https://github.com/celo-org/kliento) to interact with the necessary Celo Core Contracts.

## How to run rosetta-cli-checks

- Install the [`rosetta-cli`](https://github.com/coinbase/rosetta-cli) according to the instructions. (Note that on Mac, installing the `rosetta-cli` to `/usr/local/bin` or adding its location to you `$PATH` will allow you to call `rosetta-cli` directly on the command line rather than needing to provide the path to the executable). Current testing has been done with `v0.5.16` of the `rosetta-cli`.
- Run the Rosetta service in the background for the respective network (currently only alfajores for both Data and Construction checks)
- Run the CLI checks for alfajores as follows:

```sh
# alfajores; specify construction or data
rosetta-cli check:construction --configuration-file PATH/TO/rosetta/rosetta-cli-conf/testnet/cli-config.json
```

_Note that running the checks to completion will take a long time if this is the first time you are running Rosetta locally. Under the hood, the service is syncing a full archive node, which takes time (likely a couple of days on a normal laptop). The construction service needs to reach the tip before submitting transactions. The data checks will take a while to complete as well (likely a couple of days on a normal laptop with the current settings) as they reconcile balances for the entire chain._

### How to generate `bootstrap_balances.json`

This is only necessary for running the data checks if it has not already been created for the particular network. Here's how to generate this for alfajores (for another network, specify the appropriate genesis block URL and output path):

```sh
go run examples/generate_balances/main.go \
  https://storage.googleapis.com/genesis_blocks/alfajores \
  rosetta-cli-conf/testnet/bootstrap_balances.json
```
