# Getting Started (Python)

## Install

```bash
pip install virtengine
```

## Connect

```python
import asyncio
from virtengine import VirtEngineClient

async def main():
    async with VirtEngineClient(grpc_endpoint="localhost:9090") as client:
        print(await client.veid.identity("ve1..."))

asyncio.run(main())
```

## Transactions

Use `VirtEngineClient.sign_and_broadcast` with protobuf messages packed into `Any`.
