"""VEID module helpers."""

from __future__ import annotations

from typing import List, Optional

import grpc

from virtengine.veid.v1 import query_pb2, query_pb2_grpc


class VEIDModule:
    """VEID module queries."""

    def __init__(self, channel: grpc.aio.Channel):
        self._query = query_pb2_grpc.QueryStub(channel)

    async def identity(self, account_address: str) -> query_pb2.QueryIdentityResponse:
        request = query_pb2.QueryIdentityRequest(account_address=account_address)
        return await self._query.Identity(request)

    async def identity_record(self, account_address: str) -> query_pb2.QueryIdentityRecordResponse:
        request = query_pb2.QueryIdentityRecordRequest(account_address=account_address)
        return await self._query.IdentityRecord(request)

    async def scope(self, account_address: str, scope_id: str) -> query_pb2.QueryScopeResponse:
        request = query_pb2.QueryScopeRequest(account_address=account_address, scope_id=scope_id)
        return await self._query.Scope(request)

    async def scopes(self, account_address: str) -> query_pb2.QueryScopesResponse:
        request = query_pb2.QueryScopesRequest(account_address=account_address)
        return await self._query.Scopes(request)

    async def scopes_by_type(
        self, account_address: str, scope_type: int
    ) -> query_pb2.QueryScopesByTypeResponse:
        request = query_pb2.QueryScopesByTypeRequest(
            account_address=account_address, scope_type=scope_type
        )
        return await self._query.ScopesByType(request)

    async def approved_clients(self) -> query_pb2.QueryApprovedClientsResponse:
        request = query_pb2.QueryApprovedClientsRequest()
        return await self._query.ApprovedClients(request)

    async def params(self) -> query_pb2.QueryParamsResponse:
        request = query_pb2.QueryParamsRequest()
        return await self._query.Params(request)
