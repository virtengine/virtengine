# VirtEngine Python SDK

Python SDK for interacting with the VirtEngine chain via gRPC.

## Install

> **UNVERIFIED — not published to PyPI.** `pip install virtengine` does not work:
> `https://pypi.org/pypi/virtengine/json` returns **404**. The package is built by
> `.github/workflows/sdk-publish.yaml` on the `release: published` event, and
> `gh release list -R virtengine/virtengine` shows only `0.1.0` — a **draft** from
> 2021 — so that workflow has never run. Install from a checkout instead
> (see task `t_9d3b0f86`):

```bash
# from the repository root
pip install ./sdk/python
```

The distribution name and version are declared in
[`sdk/python/pyproject.toml`](../python/pyproject.toml) (`name = "virtengine"`,
`version = "0.1.0"`).

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
