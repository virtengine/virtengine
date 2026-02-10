"""Escrow module helpers."""

from __future__ import annotations

import grpc

from virtengine.escrow.v1 import query_pb2, query_pb2_grpc


class EscrowModule:
    """Escrow module queries."""

    def __init__(self, channel: grpc.aio.Channel):
        self._query = query_pb2_grpc.QueryStub(channel)

    async def balances(self, owner: str = "") -> query_pb2.QueryBalancesResponse:
        request = query_pb2.QueryBalancesRequest(owner=owner)
        return await self._query.Balances(request)

    async def account(self, address: str) -> query_pb2.QueryAccountResponse:
        request = query_pb2.QueryAccountRequest(address=address)
        return await self._query.Account(request)

    async def params(self) -> query_pb2.QueryParamsResponse:
        request = query_pb2.QueryParamsRequest()
        return await self._query.Params(request)
