# VirtEngine Chain SDK

[![SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml/badge.svg?branch=develop)](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml?query=branch%3Adevelop)
[![Portal CI](https://github.com/virtengine/virtengine/actions/workflows/portal-ci.yaml/badge.svg?branch=develop)](https://github.com/virtengine/virtengine/actions/workflows/portal-ci.yaml?query=branch%3Adevelop)

| Package | Gated by | What runs |
| --- | --- | --- |
| [`sdk/go`](./go) | [SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml) | `go test ./...` + `golangci-lint` |
| [`sdk/ts`](./ts) | [SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml) | `npm run lint`, `npm run build` |
| [`sdk/ts`](./ts) tests | [Protobuf and Module Contract Gate](https://github.com/virtengine/virtengine/actions/workflows/proto-generation.yaml) | `npm run lint`, `npm test -- --runInBand`, `npm run build` (all 40 suites) |
| [`sdk/portal`](./portal) | [Portal CI](https://github.com/virtengine/virtengine/actions/workflows/portal-ci.yaml) (job `SDK Portal`) | `pnpm type-check`, `pnpm test`, `pnpm build` |
| [`sdk/python`](./python) | **none — see below** | `pytest tests/` (measured green, not yet wired) |
| [`sdk/rust`](./rust) | **none — see below** | — |

> **Known gap: `sdk/python` and `sdk/rust` have no CI.** Neither is a supported
> release contract — `sdk/generation/toolchain.json:46-47` declares that no
> Python protobuf SDK is published ("generation fails closed") and that the Rust
> templates are experimental ("no Rust protobuf SDK is a release output"), and
> `sdk/generation/generate.sh:333` refuses both by name. Gating them would mean
> gating output the build deliberately does not produce. Their current state:
>
> - **`sdk/python` — collectible and green, but still ungated.** Both defects
>   this block used to report are fixed:
>
>   1. *The `google/` shadowing package is deleted.* It was a 0-byte
>      `__init__.py` plus generated `*_pb2` modules and no `descriptor.py`, so it
>      shadowed the installed `google.protobuf` whenever `sdk/python` was first on
>      `sys.path`. `google/api/*` now comes from the `googleapis-common-protos`
>      dependency instead, and the `google` entry was dropped from the
>      `pyproject.toml` package list.
>   2. *The gencode/runtime version floor is corrected.* The checked-in gencode is
>      mixed-version — of the 194 tracked `*_pb2.py` files, 179 declare
>      `# Protobuf Python Version: 6.33.5` and 15 declare `7.36.2`. Protobuf
>      runtimes ≥5 validate gencode at import and hard-fail when the gencode is
>      newer, so the runtime must sit at or above the newest gencode in the tree;
>      the floor is therefore `protobuf = ">=7.36.2,<8"`, not the previous
>      `^5.29.6`.
>
>   Measured on Python 3.11 and 3.12: `pytest tests/ --collect-only` exits 0 with
>   5 tests collected, and `pytest tests/` passes 5/5. The un-gated state is a
>   coverage decision, not a blocker — there is nothing red to wire up yet. The
>   reason it is not simply added to [SDK CI](https://github.com/virtengine/virtengine/actions/workflows/sdk-ci.yaml)
>   is a packaging one: this SDK has no lockfile and no supported generation
>   contract, so a gate would resolve fresh dependency ranges on every run rather
>   than test a pinned tree. Wiring it is a deliberate follow-up, not an oversight.
> - **`sdk/rust` — knowingly uncompilable, and deliberately ungated.**
>   `cargo check --all-targets` fails with 6 errors, measured:
>
>   | Error | Site |
>   | --- | --- |
>   | `expected identifier, found keyword 'mod'` | `src/proto/mod.rs:123` (`pub mod mod {`) |
>   | `E0428`: `query_client` redefined | `src/proto/cosmos.auth.v1beta1.tonic.rs:3` |
>   | `E0428`: `query_server` redefined | `src/proto/cosmos.auth.v1beta1.tonic.rs:348` |
>   | `E0428`: `service_client` redefined | `src/proto/cosmos.tx.v1beta1.tonic.rs:3` |
>   | `E0428`: `service_server` redefined | `src/proto/cosmos.tx.v1beta1.tonic.rs:312` |
>   | `recursion limit reached while expanding 'stringify!'` | `src/proto/cometbft.abci.v2.rs:18` (`prost::Oneof` derive) |
>
>   The `pub mod mod` entry is not merely a keyword error — it is
>   `pub mod mod { include!("mod.rs"); }`, i.e. a self-include. There is no
>   `package mod` and no `mod.proto` anywhere in the repository, so this module
>   is orphan generated output with no regeneration path: neither
>   `sdk/buf.gen.rust-sdk.yaml` nor any make target is wired to
>   `sdk/rust/src/proto/`. It is left in place rather than hand-patched, because
>   editing a `@generated` file by hand would hide the real defect (a generator
>   that was never run for this tree) behind a one-line cosmetic fix.
>
>   `sdk/rust` also cannot simply be deleted: the hand-written `src/client.rs`,
>   `src/tx.rs`, `src/modules/*.rs` all import `crate::proto::…`, so the
>   generated tree is load-bearing for the crate even though it does not compile.
>   Making it build requires regenerating the tree from a supported Rust template
>   first — which is the same reason it is not a release output.
>
> Decision owner for both: SDK surface policy is operator/chain-core. Re-verify
> the numbers above before quoting them; they are not asserted by CI.

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

### Published SDK packages

Only `sdk/ts` is published to a public registry. The other SDKs are **source-only**
and are deliberately not published:

| SDK | Published? | Registry | Install |
| --- | --- | --- | --- |
| `sdk/go` | yes (Go modules) | `pkg.go.dev` via module path | `go get github.com/virtengine/virtengine/sdk/go/...` |
| `sdk/ts` | yes, **manually and rarely** | npm `@virtengine/chain-sdk` | `npm install @virtengine/chain-sdk@alpha` |
| `sdk/python` | **no** — source-only | — | build from this repo |
| `sdk/rust` | **no** — source-only | — | build from this repo |
| `sdk/portal` | **no** — source-only | — | internal portal workspace |

`sdk/python` and `sdk/rust` are source-only by operator decision, not by oversight.
`.github/workflows/sdk-publish.yaml` used to trigger on `release: types: [published]`
and run `poetry publish` / `cargo publish`; because the repository has only ever had a
2021 **draft** release, that event never fired and the path was never exercised. Cutting
a real first release would have fired a live, irreversible publish as an unattended side
effect of tagging. The workflow is now `workflow_dispatch`-only and its job refuses to
publish. Do not reintroduce a `release: published` trigger without a human decision.

### `sdk/release-please-config.json` is NOT in use

`sdk/release-please-config.json` declares a `ts` package (`@virtengine/chain-sdk`,
prerelease alpha). **No workflow in this repository consumes it** — verified: no file
under `.github/` references `release-please`. It is retained as intent for a future
release-please rollout, not as a description of current behaviour. It does not version
the package and nothing reads `changelog-path` from it.

### Publishing `@virtengine/chain-sdk` (current manual path)

There is **no automated npm publish path** in this repository. The only recipe is the
`release-ts` target in `sdk/make/release-ts.mk:1-2`, which runs `npm run release`
(build, then test, then `npm publish --tag alpha`). The root Makefile has no
`release-ts` target, so `make release-ts` from the repo root fails with
`make: *** No rule to make target 'release-ts'.  Stop.`

```shell
# From the sdk/ directory. NOTE: sdk/Makefile requires direnv and stops with
# "No direnv in <PATH>" if it is absent, so `direnv allow` must have run first.
cd sdk && direnv allow && make release-ts
```

Before running it, note three things:

1. It **publishes**. `npm run release` ends in `npm publish --tag alpha`, so a dry run
   is not available on this path.
2. It requires npm publish credentials in your environment (`NPM_TOKEN`, or an
   interactive login). The repository configures no `registry-url` and holds no npm
   token secret, because no workflow publishes to npm.
3. It also runs `modvendor` and `buf` as prerequisites, so it mutates the working tree
   before publishing.

`npm publish` is refused by an unconfigured or already-published version, so the
`package.json` version must be bumped first. The current published version and the
in-repo version are both `1.0.0-alpha.20`; npm has exactly one version, first published
2026-02-12, and it has not moved since. Consumers pinning `@alpha` are therefore on a
package that is months behind this repository — confirm intent before publishing again.

Automating this path requires a credential decision that a bot must not make
unattended: npm trusted publishing (OIDC, no long-lived token) versus a stored
`NPM_TOKEN`, and whether alpha tags should be auto-published at all.

## Contributing

Please submit issues on this repository: <https://github.com/virtengine/virtengine/issues>.
The previous pointer to a separate `virtengine/support` repository
(`gh api repos/virtengine/support` → 404) has been removed, along with the
"all pull requests must be associated with an open issue in the support
repository" requirement, which no such repository backs. See
[CONTRIBUTING.md](../CONTRIBUTING.md) at the repository root for the actual
contribution process.
