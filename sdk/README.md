# VirtEngine Chain SDK

[![SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml/badge.svg?branch=develop)](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml?query=branch%3Adevelop)
[![Portal CI](https://github.com/virtengine/virtengine/actions/workflows/portal-ci.yaml/badge.svg?branch=develop)](https://github.com/virtengine/virtengine/actions/workflows/portal-ci.yaml?query=branch%3Adevelop)

| Package | Gated by | What runs |
| --- | --- | --- |
| [`sdk/go`](./go) | [SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml) | `go test ./...` + `golangci-lint` |
| [`sdk/ts`](./ts) | [SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml) | `npm run lint`, `npm run build` |
| [`sdk/ts`](./ts) tests | [Protobuf and Module Contract Gate](https://github.com/virtengine/virtengine/actions/workflows/proto-generation.yaml) | `npm run lint`, `npm test -- --runInBand`, `npm run build` (all 40 suites) |
| [`sdk/portal`](./portal) | [Portal CI](https://github.com/virtengine/virtengine/actions/workflows/portal-ci.yaml) (job `SDK Portal`) | `pnpm type-check`, `pnpm test`, `pnpm build` |
| [`sdk/python`](./python) | **none — see below** | — |
| [`sdk/rust`](./rust) | **none — see below** | — |

> **Known gap: `sdk/python` and `sdk/rust` have no CI.** Neither is a supported
> release contract — `sdk/generation/toolchain.json:46-47` declares that no
> Python protobuf SDK is published ("generation fails closed") and that the Rust
> templates are experimental ("no Rust protobuf SDK is a release output"), and
> `sdk/generation/generate.sh:333` refuses both by name. Gating them would mean
> gating output the build deliberately does not produce.
>
> Two concrete defects make them un-gateable today, independent of that policy:
>
> - `sdk/python` — pytest cannot collect. The generated `google/protobuf/`
>   package checked in next to the code is a 0-byte `__init__.py` plus generated
>   `*_pb2` modules with no `descriptor.py`, so it shadows the installed
>   `google.protobuf` whenever `sdk/python` is first on `sys.path`. Measured:
>   `pytest tests/ --collect-only` → 3 collection errors, `ImportError: cannot
>   import name 'descriptor' from 'google.protobuf'`. All 3 test modules error.
> - `sdk/rust` — the checked-in `rust/src/proto/mod.rs:123` contains
>   `pub mod mod {`. `mod` is a Rust keyword, so this cannot compile; the file is
>   marked `@generated` and is consistent with output that was never compiled.
>
> These are tracked as separate work items; this workflow gates only what is
> known green so that a red check always means a real regression.

## Overview

The `sdk/` directory of the [virtengine/virtengine](https://github.com/virtengine/virtengine)
monorepo is a development gateway to the VirtEngine Blockchain.
It aims following:

- Define data types and API via [protobuf](./proto)
  - VirtEngine Blockchain and it's stores, aka [node](./proto/node)
  - VirtEngine provider Interface, aka [provider](./proto/provider)
- Define data types and API (both REST and GRPC) of VirtEngine provider Interface
- Provide official reference clients for supported [programming languages](#supported-languages)

## Supported languages

### Golang

[This implementation](./go) provider all necessary code-generation as well as client defining VirtEngine Blockchain.
There are a few packages this implementation exports. Import them by their full module path
(`github.com/virtengine/virtengine/sdk/go/...`); the previous reference to Vanity URLs hosted as
[GitHub Pages](https://github.com/virtengine/vanity) has been removed because
`gh api repos/virtengine/vanity` returns 404 — that repository does not exist.

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

3. Run codegen from the `sdk/` directory (these targets are defined in
   `sdk/make/codegen.mk` and `sdk/make/mod.mk`, not in the root Makefile).
   This will
   - Install all required tools into local cache
   - Make sure you setup vendor

   ```shell
   make modvendor
   ```

   - generate changes to all [supported programming languages](#supported-languages)

   ```shell
   make proto-gen
   ```

   - to run codegen for a specific module use `make proto-gen-<module>`,
     where `<module>` is one of `go`, `ts`, or `doc` — the exact list is
     `PROTO_GEN_MODS` in `sdk/make/codegen.mk:1-3`. Targets `proto-gen-rust`
     (`codegen.mk:17`) and `proto-gen-pulsar` (`codegen.mk:38`) also exist but are
     **not** part of `PROTO_GEN_MODS`, so bare `make proto-gen` does not run them.
     For example

   ```shell
   make proto-gen-go
   ```

## Releases

Releases indicate changes to the repository itself. API versions are defined within each module.

## Contributing

Please submit issues on this repository: <https://github.com/virtengine/virtengine/issues>.
The previous pointer to a separate `virtengine/support` repository
(`gh api repos/virtengine/support` → 404) has been removed, along with the
"all pull requests must be associated with an open issue in the support
repository" requirement, which no such repository backs. See
[CONTRIBUTING.md](../CONTRIBUTING.md) at the repository root for the actual
contribution process.
