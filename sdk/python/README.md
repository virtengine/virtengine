# VirtEngine Python SDK

Python SDK for interacting with the VirtEngine chain via gRPC.

## Install

```bash
pip install virtengine
```

## Quick start

```python
import asyncio
from virtengine import VirtEngineClient

async def main():
    async with VirtEngineClient(grpc_endpoint="localhost:9090") as client:
        identity = await client.veid.identity("ve1...")
        print(identity)

asyncio.run(main())
```

## Transactions

```python
from google.protobuf import any_pb2
from virtengine.veid.v1 import tx_pb2

msg = tx_pb2.MsgSubmitScope(
    scope_type=1,
    encrypted_data=b"...",
    client_sig=b"...",
)
msg_any = any_pb2.Any()
msg_any.Pack(msg)

response = await client.sign_and_broadcast([msg_any])
print(response.tx_response.code)
```

## Events

```python
subscriber = client.events()
async for event in subscriber.subscribe("tm.event='NewBlock'"):
    print(event)
```

## Development

```bash
poetry install
poetry run pytest
```
