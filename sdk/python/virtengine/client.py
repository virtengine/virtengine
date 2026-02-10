"""Main client for interacting with VirtEngine."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Optional, Sequence

import grpc
from google.protobuf import any_pb2

from cosmos.auth.v1beta1 import auth_pb2, query_pb2, query_pb2_grpc
from cosmos.tx.v1beta1 import service_pb2

from virtengine.events import EventSubscriber
from virtengine.modules import EscrowModule, MarketModule, ModuleRegistry, VEIDModule
from virtengine.tx import TxBroadcaster, TxBuilder
from virtengine.wallet import Wallet


@dataclass(frozen=True)
class AccountInfo:
    account_number: int
    sequence: int
    address: str


class VirtEngineClient:
    """Main client for interacting with VirtEngine blockchain."""

    def __init__(
        self,
        grpc_endpoint: str = "localhost:9090",
        rest_endpoint: str = "http://localhost:1317",
        chain_id: str = "virtengine-1",
        wallet: Optional[Wallet] = None,
        rpc_ws_endpoint: Optional[str] = None,
    ):
        self.grpc_endpoint = grpc_endpoint
        self.rest_endpoint = rest_endpoint
        self.chain_id = chain_id
        self.wallet = wallet
        self.rpc_ws_endpoint = rpc_ws_endpoint or self._default_ws_endpoint(rest_endpoint)

        self._channel = grpc.aio.insecure_channel(grpc_endpoint)
        self._auth_query = query_pb2_grpc.QueryStub(self._channel)

        self.veid = VEIDModule(self._channel)
        self.market = MarketModule(self._channel)
        self.escrow = EscrowModule(self._channel)
        self.modules = ModuleRegistry(self._channel)

        self._tx_builder = TxBuilder(chain_id)
        self._tx_broadcaster = TxBroadcaster(self._channel)

    @staticmethod
    def _default_ws_endpoint(rest_endpoint: str) -> str:
        if rest_endpoint.startswith("https://"):
            return "wss://" + rest_endpoint[len("https://") :] + "/websocket"
        if rest_endpoint.startswith("http://"):
            return "ws://" + rest_endpoint[len("http://") :] + "/websocket"
        return f"ws://{rest_endpoint}/websocket"

    @classmethod
    def from_mnemonic(
        cls,
        mnemonic: str,
        grpc_endpoint: str = "localhost:9090",
        **kwargs,
    ) -> "VirtEngineClient":
        wallet = Wallet.from_mnemonic(mnemonic)
        return cls(grpc_endpoint=grpc_endpoint, wallet=wallet, **kwargs)

    async def sign_and_broadcast(
        self,
        messages: Sequence[any_pb2.Any],
        memo: str = "",
        fee: Optional[dict] = None,
        mode: service_pb2.BroadcastMode = service_pb2.BROADCAST_MODE_SYNC,
    ) -> service_pb2.BroadcastTxResponse:
        if not self.wallet:
            raise ValueError("Wallet required for signing")

        account = await self.get_account(self.wallet.address)
        tx = self._tx_builder.build(
            messages=messages,
            signer=self.wallet,
            account_number=account.account_number,
            sequence=account.sequence,
            memo=memo,
            fee=fee,
        )
        return await self._tx_broadcaster.broadcast(tx, mode=mode)

    async def get_account(self, address: str) -> AccountInfo:
        request = query_pb2.QueryAccountRequest(address=address)
        response = await self._auth_query.Account(request)
        base_account = self._unpack_account(response.account)
        return AccountInfo(
            account_number=base_account.account_number,
            sequence=base_account.sequence,
            address=base_account.address,
        )

    @staticmethod
    def _unpack_account(account_any: any_pb2.Any) -> auth_pb2.BaseAccount:
        base_account = auth_pb2.BaseAccount()
        if account_any.Is(base_account.DESCRIPTOR):
            account_any.Unpack(base_account)
            return base_account
        module_account = auth_pb2.ModuleAccount()
        if account_any.Is(module_account.DESCRIPTOR):
            account_any.Unpack(module_account)
            if module_account.base_account is not None:
                return module_account.base_account
        raise ValueError("Unsupported account type")

    def events(self) -> EventSubscriber:
        return EventSubscriber(self.rpc_ws_endpoint)

    async def close(self) -> None:
        await self._channel.close()

    async def __aenter__(self) -> "VirtEngineClient":
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb) -> None:
        await self.close()
