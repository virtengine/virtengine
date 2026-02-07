"""Market module helpers."""

from __future__ import annotations

import grpc

from virtengine.market.v1beta5 import query_pb2, query_pb2_grpc


class MarketModule:
    """Market module queries."""

    def __init__(self, channel: grpc.aio.Channel):
        self._query = query_pb2_grpc.QueryStub(channel)

    async def orders(self, state: str | None = None) -> query_pb2.QueryOrdersResponse:
        request = query_pb2.QueryOrdersRequest(state=state or "")
        return await self._query.Orders(request)

    async def order(self, order_id: str) -> query_pb2.QueryOrderResponse:
        request = query_pb2.QueryOrderRequest(order_id=order_id)
        return await self._query.Order(request)

    async def leases(self) -> query_pb2.QueryLeasesResponse:
        request = query_pb2.QueryLeasesRequest()
        return await self._query.Leases(request)

    async def lease(self, lease_id: str) -> query_pb2.QueryLeaseResponse:
        request = query_pb2.QueryLeaseRequest(lease_id=lease_id)
        return await self._query.Lease(request)

    async def params(self) -> query_pb2.QueryParamsResponse:
        request = query_pb2.QueryParamsRequest()
        return await self._query.Params(request)
