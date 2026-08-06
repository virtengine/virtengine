# Reproducible Protobuf and Contract Generation

VirtEngine has one supported contract source graph: `sdk/proto/node` and `sdk/proto/provider`, resolved by `sdk/buf.yaml` and the commit/digest-pinned `sdk/buf.lock`.

## Supported commands

- `./scripts/proto-generate.sh all` — build the pinned Linux image and regenerate every supported output.
- `./scripts/proto-generate.sh go|descriptor|openapi|ts|inventory` — regenerate one output class.
- `./scripts/verify-proto-generation.sh` — regenerate twice and require byte-identical checked-in artifacts.
- `./scripts/verify-modules.sh` — verify every workspace/tool module, required checksums, and SDK vendor synchronization.
- `VE_VERIFY_EMPTY_CACHE=1 ./scripts/verify-modules.sh` — additionally download all module graphs into an empty temporary cache.
- `./scripts/proto-generate-wsl.sh <mode>` — native Linux/WSL equivalent used only when Docker is unavailable. It builds Go plugins from the isolated tool module and downloads Node only after checking the pinned SHA-256.

Do not invoke `buf generate` directly and do not use host-discovered plugins for release artifacts.

## Pinned environment

`sdk/generation/toolchain.json` is the machine-readable tool manifest. `sdk/generation/Dockerfile` pins its base image by digest and verifies downloaded protoc/Node archives by SHA-256. Go generators are built from `sdk/generation/go.mod` and `go.sum`. TypeScript dependencies are installed with `npm ci` from `sdk/ts/package-lock.json`.

The first online run fills explicit caches. A subsequent run may operate from those caches. Generation never mutates application `go.mod` or `go.sum`; dependency updates require a separate reviewed change. Python and Rust protobuf SDKs are not declared release outputs and fail closed instead of silently producing stale clients.

## Checked-in outputs

- Go messages/services: `sdk/go/node/**/*.pb.go`
- Go REST gateways: `sdk/go/node/**/*.pb.gw.go`
- Descriptor set and digest: `sdk/artifacts/proto/virtengine.binpb*`
- Module/service/route/output inventory: `sdk/artifacts/proto/inventory.json*`
- Protobuf-derived OpenAPI: `api/openapi/virtengine-proto.swagger.json`
- TypeScript contracts: `sdk/ts/src/generated/**`

`api/openapi/portal_api.yaml` remains the manual provider portal API because it represents non-protobuf HTTP services. `api/openapi/virtengine-api.yaml` remains curated product documentation; blockchain operations are authoritative and drift-checked in the generated Swagger document.

## Compatibility and migrations

Before changing a field type/number, enum numeric value, persisted message, genesis shape, type URL, service/method name, or HTTP route:

1. add/update a binary or JSON fixture under `tests/compatibility`;
2. run Buf breaking detection against the release descriptor baseline;
3. preserve existing field numbers and type URLs, or ship an approved deterministic migration;
4. regenerate all outputs and compile Go and TypeScript consumers;
5. run live gRPC/REST gateway parity tests.

Task 84A's `MsgSubmitConsensusVerification` and vote-extension messages are generated from `sdk/proto/node/virtengine/veid/v1/tx.proto`; their field numbers must not be edited manually.
