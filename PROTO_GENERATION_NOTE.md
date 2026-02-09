# Proto Generation Note

## Issue
The protobuf generation pipeline (`buf generate`) is not generating Go code for the new domain verification messages added to `sdk/proto/node/virtengine/provider/v1beta4/msg.proto` and events in `event.proto`.

## New Messages Added
- `MsgRequestDomainVerification` / `MsgRequestDomainVerificationResponse`
- `MsgConfirmDomainVerification` / `MsgConfirmDomainVerificationResponse` 
- `MsgRevokeDomainVerification` / `MsgRevokeDomainVerificationResponse`
- `VerificationMethod` enum

## New Events Added
- `EventProviderDomainVerificationRequested`
- `EventProviderDomainVerificationConfirmed`
- `EventProviderDomainVerificationRevoked`

## To Complete Implementation

1. **Regenerate Protobuf Files:**
   ```bash
   cd sdk
   buf generate proto/node --template buf.gen.gogo.yaml
   ```
   
   Expected output files:
   - `sdk/go/node/provider/v1beta4/msg.pb.go` (should contain new message structs)
   - `sdk/go/node/provider/v1beta4/event.pb.go` (should contain new event structs)

2. **After successful generation, the following files have placeholder implementations that need the generated types:**
   - `sdk/go/node/provider/v1beta4/msgs.go` - Has ValidateBasic implementations ready
   - `x/provider/handler/server.go` - Has handler methods ready
   - `x/provider/keeper/domain_verification.go` - Has keeper methods ready
   - `sdk/go/cli/provider_tx.go` - Has CLI commands ready

## Current Status

All implementation logic is complete, but waiting on protobuf generation to create the actual message/event type definitions. The logic has been written to use these types once they're generated.

## What Works Now

- Keeper logic for domain verification (request, confirm, revoke)
- Handler wiring for new messages 
- CLI commands structure
- Tests (will need generated types to run)

## What Needs Proto Generation

- Message type definitions
- Event type definitions  
- gRPC service definitions

## Alternative if buf continues having issues

Use protoc directly:
```bash
protoc \
  --gocosmos_out=. \
  --grpc-gateway_out=. \
  -I sdk/proto/node \
  -I ~/.buf/cache \
  sdk/proto/node/virtengine/provider/v1beta4/msg.proto \
  sdk/proto/node/virtengine/provider/v1beta4/event.proto
```
