# VirtEngine Chain SDK

[![Lint Status](https://github.com/virtengine/chain-sdk/actions/workflows/lint.yaml/badge.svg)](https://github.com/virtengine/chain-sdk/actions/workflows/lint.yaml)
[![Test Status](https://github.com/virtengine/chain-sdk/actions/workflows/tests.yaml/badge.svg)](https://github.com/virtengine/chain-sdk/actions/workflows/tests.yaml)

## Overview

This repository is a development gateway to the VirtEngine Blockchain.
It aims following:

- Define data types and API via [protobuf](./proto)
  - VirtEngine Blockchain and it's stores, aka [node](./proto/node)
  - VirtEngine provider Interface, aka [provider](./proto/provider)
- Define data types and API (both REST and GRPC) of VirtEngine provider Interface
- Provide official reference clients for supported [programming languages](#supported-languages)

## Supported languages

### Golang

[This implementation](./go) provider all necessary code-generation as well as client defining VirtEngine Blockchain
There are a few packages this implementation exports. All packages available via Vanity URLs which are hosted as [Github Pages](https://github.com/virtengine/vanity).

#### Go package

Source code is located within [go](./go) directory

Contains all the types, clients and utilities necessary to communicate with VirtEngine Blockchain

```go
import "github.com/virtengine/virtengine/sdk/go"
```

##### Migrate package

Depending on difference in API and stores between current and previous versions of the blockchain, there may be a **migrate** package. It is intended to be used by [node](https://github.com/virtengine/virtengine) only.

```go
import "github.com/virtengine/virtengine/sdk/go/node/migrate"
```

#### SDL package

Reference implementation of the SDL.

```go
import "github.com/virtengine/virtengine/sdk/go/sdl"
```

#### CLI package

CLI package which combines improved version of cli clients from node](https://github.com/virtengine/virtengine) and [cosmos-sdk](https://github.com/cosmos/cosmos-sdk)

```go
import "github.com/virtengine/virtengine/sdk/go/cli"
```

### TS

Source code is located within [ts](./ts) directory

## Protobuf

All protobuf definitions are located within [proto](./proto) directory.

This repository consolidates gRPC API definitions for the [VirtEngine Node](https://github.com/virtengine/virtengine). It also includes related code generation.

Currently, two `buf` packages are defined, with potential future publication to BSR based on demand:

- **Node Package**: `buf.build/virtengine/virtengine`
- **Provider Package**: `buf.build/virtengine/provider`

Proto documentation is available for:

- [Node](docs/proto/node.md)
- [Provider](docs/proto/provider.md)

Documentation in swagger format combining both node and provider packages can be located [here](./docs/swagger-ui/swagger.yaml)

### How to run protobuf codegen

If there is a need to run regenerate protobuf (in case of API or documentation changes):

1. Install [direnv](https://direnv.net) and hook it to the [shell](https://direnv.net/docs/hook.html)
   - **MacOS**
   ```shell
   brew install make direnv
   ```
2. Allow direnv within project

   ```shell
   direnv allow
   ```

3. Run codegen. This will
   - Install all required tools into local cache
   - Make sure you setup vendor

   ```shell
   make modvendor
   ```

   - generate changes to all [supported programming languages](#supported-languages)

   ```shell
   make proto-gen
   ```

   - to run codegen for specific language use `make proto-gen-<lang>`. For example

   ```shell
   make proto-gen-go
   ```

## Releases

Releases indicate changes to the repository itself. API versions are defined within each module.

## Contributing

Please submit issues via the [support repository](https://github.com/virtengine/support/issues) and tag them with `repo/chain-sdk`. All pull requests must be associated with an open issue in the support repository.
