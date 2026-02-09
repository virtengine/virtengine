# Provider Module (x/provider) — AGENTS Guide

## Package Overview
- Purpose: chain module that owns provider registration, lifecycle management, domain verification, and provider public-key management.
- Use when: Updating on-chain provider state or provider-facing governance logic; use the TypeScript SDK in `sdk/ts` for off-chain clients.
- Key entry points:
  - `AppModuleBasic` and `AppModule` for module wiring in the app (`x/provider/module.go:44`, `x/provider/module.go:49`).
  - `keeper.IKeeper` for cross-module reads/writes (`x/provider/keeper/keeper.go:15`).
  - `handler.NewMsgServerImpl` for MsgServer registration (`x/provider/handler/server.go:34`).
  - Module constants and store prefixes in `sdk/go/node/provider/v1beta4/key.go:3`.

## Architecture
- Entry points:
  - Module wiring and registration: `x/provider/module.go:44`.
  - Keeper implementation: `x/provider/keeper/keeper.go:15`.
  - MsgServer (tx handlers): `x/provider/handler/server.go:34`.
  - Domain verification flow: `x/provider/keeper/domain_verification.go:47`.
  - SDK types and keys: `sdk/go/node/provider/v1beta4/key.go:3` and `sdk/go/node/provider/v1beta4/types.go:11`.
- Directory layout:
  - `x/provider/keeper/` — store access, domain verification, public-key management.
  - `x/provider/handler/` — gRPC MsgServer implementation and error mapping.
  - `x/provider/query/` — query service wiring.
  - `x/provider/types/daemon/` — provider-daemon transport models.
  - `sdk/go/node/provider/v1beta4/` — protobuf-generated types, errors, and keys.

## Core Concepts
- Provider lifecycle: `CreateProvider`, `UpdateProvider`, `DeleteProvider` validate input, enforce constraints, emit typed events (`x/provider/handler/server.go:47`).
- Domain verification uses DNS TXT records with `_virtengine-verification` prefix and token expiry; verification state is stored in-module and emits events (`x/provider/keeper/domain_verification.go:17`, `x/provider/keeper/domain_verification.go:81`).
- Public-key management supports `ed25519`, `x25519`, and `secp256k1` key types (`sdk/go/node/provider/v1beta4/key.go:13`) and is exposed via `keeper.IKeeper` (`x/provider/keeper/keeper.go:26`).
- Error handling uses Cosmos SDK error wrapping and module-registered errors for deterministic codes (`x/provider/handler/server.go:18`).

## Usage Examples

### Wiring the module in the app
```go
import (
  provider "github.com/virtengine/virtengine/x/provider"
  providerkeeper "github.com/virtengine/virtengine/x/provider/keeper"
  providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

providerKeeper := providerkeeper.NewKeeper(appCodec, keys[providertypes.StoreKey])
providerModule := provider.NewAppModule(
  appCodec,
  providerKeeper,
  accountKeeper,
  bankKeeper,
  marketKeeper,
  veidKeeper,
  mfaKeeper,
)
moduleManager.RegisterModules(providerModule)
```

### Using the keeper for domain verification
```go
record, err := providerKeeper.GenerateDomainVerificationToken(ctx, providerAddr, "example.com")
if err != nil {
  return err
}
// DNS TXT record is expected at: _virtengine-verification.example.com
err = providerKeeper.VerifyProviderDomain(ctx, providerAddr)
```

## Implementation Patterns
- Add new provider messages by updating protobuf definitions in `sdk/go/node/provider/v1beta4/`, re-generating code, and extending MsgServer in `x/provider/handler/server.go:47`.
- Keep state mutations inside the keeper; MsgServer should validate and delegate to keeper methods (`x/provider/keeper/keeper.go:70`).
- Add new domain verification logic in `x/provider/keeper/domain_verification.go:47` and wire new events in `sdk/go/node/provider/v1beta4/event.pb.go`.
- Tests live alongside keepers and handlers; mirror error scenarios and event emission in tests under `x/provider/keeper/*_test.go` and `x/provider/handler/handler_test.go`.
- Anti-patterns:
  - Do not bypass `ValidateBasic()` in MsgServer handlers (`x/provider/handler/server.go:50`).
  - Do not write directly to KVStore outside the keeper (`x/provider/keeper/keeper.go:70`).

## API Reference
- `provider.NewAppModule(codec.Codec, keeper.IKeeper, govtypes.AccountKeeper, bankkeeper.Keeper, mkeeper.IKeeper, veidkeeper.IKeeper, mfakeeper.IKeeper) AppModule` (`x/provider/module.go:110`).
- `keeper.NewKeeper(codec.BinaryCodec, storetypes.StoreKey) IKeeper` (`x/provider/keeper/keeper.go:48`).
- `keeper.IKeeper` key methods:
  - `Get`, `Create`, `Update`, `Delete` (`x/provider/keeper/keeper.go:18`).
  - `SetProviderPublicKey`, `RotateProviderPublicKey` (`x/provider/keeper/keeper.go:29`).
  - `GenerateDomainVerificationToken`, `VerifyProviderDomain` (`x/provider/keeper/keeper.go:35`).
- MsgServer methods: `CreateProvider`, `UpdateProvider`, `DeleteProvider`, `GenerateDomainVerificationToken`, `VerifyProviderDomain` (`x/provider/handler/server.go:47`).

## Dependencies & Environment
- Depends on Cosmos SDK modules (`x/market`, `x/veid`, `x/mfa`) via keeper interfaces (`x/provider/module.go:53`).
- Uses DNS lookups for domain verification (`x/provider/keeper/domain_verification.go:95`); ensure DNS access in integration tests.
- No package-specific environment variables.

## Configuration
- No module parameters; verification settings are compile-time constants (`x/provider/keeper/domain_verification.go:26`).
- DNS verification uses `_virtengine-verification` TXT records (`x/provider/keeper/domain_verification.go:33`).

## Testing
- Unit tests:
  - `x/provider/handler/handler_test.go`
  - `x/provider/keeper/*_test.go`
  - `x/provider/types/daemon/*_test.go`
- Recommended commands:
  - `go test ./x/provider/... -count=1`
  - `go test ./sdk/go/node/provider/v1beta4 -count=1`
<<<<<<< HEAD

## Troubleshooting
- Domain verification fails unexpectedly
  - Cause: TXT record missing or cached with old token.
  - Fix: Re-issue token and verify with DNS tools before retrying.
=======
>>>>>>> 757ceb4f (docs(provider): add AGENTS guides for core packages)
