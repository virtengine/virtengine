"""WebSocket event subscription helpers."""

from __future__ import annotations

import asyncio
import json
from typing import AsyncIterator, Optional

import websockets


class EventSubscriber:
    """Subscribe to Tendermint RPC events via WebSocket."""

    def __init__(self, ws_endpoint: str):
        self.ws_endpoint = ws_endpoint

    async def subscribe(self, query: str, timeout: Optional[float] = None) -> AsyncIterator[dict]:
        request = {
            "jsonrpc": "2.0",
            "id": "1",
            "method": "subscribe",
            "params": {"query": query},
        }
        async with websockets.connect(self.ws_endpoint) as websocket:
            await websocket.send(json.dumps(request))
            while True:
                if timeout:
                    message = await asyncio.wait_for(websocket.recv(), timeout=timeout)
                else:
                    message = await websocket.recv()
                yield json.loads(message)
