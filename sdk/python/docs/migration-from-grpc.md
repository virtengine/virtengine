# Migration from raw gRPC (Python)

The SDK wraps generated protobuf clients and provides helper methods for
transaction signing and broadcasting. If you previously used gRPC stubs directly:

1. Replace direct `QueryStub` usage with `VirtEngineClient` module helpers.
2. Use `VirtEngineClient.sign_and_broadcast` instead of manual tx assembly.
3. Keep using generated protobuf messages for request/response payloads.
