# Getting Started (Python)

## Install

> **UNVERIFIED — not published to PyPI.** `pip install virtengine` returns 404 from
> `https://pypi.org/pypi/virtengine/json`; the publish workflow
> (`.github/workflows/sdk-publish.yaml`) fires only on `release: published` and
> the repository has never published a release. See
> [the SDK README](../README.md) for the verified install path.

```bash
# from the repository root
pip install ./sdk/python
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
