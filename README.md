# Introduction
DCRM wallet service is a distributed key generation and distributed signature service that can serve as a distributed custodial solution.

# Prerequisites
1. VPS server with 1 CPU and 2G mem
2. Golang ^1.12

# Setting Up
## Clone The Repository
To get started, launch your terminal and download the latest version of the SDK.
```
mkdir -p $GOPATH/src/github.com/fsn-dev

cd $GOPATH/src/github.com/fsn-dev

git clone https://github.com/fsn-dev/dcrm-walletService.git
```
## Build
Next compile the code.  Make sure you are in /walletService directory.
```
cd walletService && make
```
## config file
cmd/conf.toml (bin/cmd/conf.toml)

## Run
Open access to the APIs by running the compiled code. 
```
./bin/cmd/gdcrm
```
The `gdcrm` will provide rpc service, the default RPC port is port 4449.

Note: 
Before call RPC API, please wait at least 5 minutes after running the node which need to prepare dcrm env.

# Front-end

After running the dcrm wallet rpc service, we can use [SMPCWallet](https://github.com/fsn-dev/SMPCWallet/releases) to connect service. Use this front-end to create managed account which support BTC/ETH/FSN.


