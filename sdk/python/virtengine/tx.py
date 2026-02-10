"""Transaction building and broadcasting helpers."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable, List, Optional, Sequence, Union

import grpc
from google.protobuf import any_pb2

from cosmos.tx.v1beta1 import service_pb2, service_pb2_grpc, tx_pb2
from cosmos.tx.signing.v1beta1 import signing_pb2
from cosmos.base.v1beta1 import coin_pb2
from cosmos.crypto.secp256k1 import keys_pb2

from virtengine.wallet import Wallet


@dataclass(frozen=True)
class FeeConfig:
    amount: Sequence[coin_pb2.Coin]
    gas_limit: int


class TxBuilder:
    """Build and sign Cosmos SDK transactions."""

    def __init__(self, chain_id: str, gas_limit: int = 200000, gas_price: str = "0.025uve"):
        self.chain_id = chain_id
        self.gas_limit = gas_limit
        self.gas_price = gas_price

    def _default_fee(self) -> FeeConfig:
        denom = self.gas_price.lstrip("0123456789.")
        amount_str = self.gas_price[: len(self.gas_price) - len(denom)]
        try:
            amount = float(amount_str)
        except ValueError:
            amount = 0.025
        fee_amount = int(amount * self.gas_limit)
        return FeeConfig(
            amount=[coin_pb2.Coin(denom=denom or "uve", amount=str(fee_amount))],
            gas_limit=self.gas_limit,
        )

    def build(
        self,
        messages: Sequence[any_pb2.Any],
        signer: Wallet,
        account_number: int,
        sequence: int,
        memo: str = "",
        fee: Optional[Union[FeeConfig, tx_pb2.Fee, dict]] = None,
    ) -> tx_pb2.TxRaw:
        """Build and sign a TxRaw for the provided messages."""
        if not messages:
            raise ValueError("At least one message is required")

        tx_body = tx_pb2.TxBody(messages=list(messages), memo=memo)

        pubkey = keys_pb2.PubKey(key=signer.public_key)
        pubkey_any = any_pb2.Any()
        pubkey_any.Pack(pubkey)

        mode_info = tx_pb2.ModeInfo(
            single=tx_pb2.ModeInfo.Single(mode=signing_pb2.SIGN_MODE_DIRECT)
        )
        signer_info = tx_pb2.SignerInfo(
            public_key=pubkey_any,
            mode_info=mode_info,
            sequence=sequence,
        )

        if fee is None:
            fee_config = self._default_fee()
            fee_msg = tx_pb2.Fee(amount=fee_config.amount, gas_limit=fee_config.gas_limit)
        elif isinstance(fee, tx_pb2.Fee):
            fee_msg = fee
        elif isinstance(fee, FeeConfig):
            fee_msg = tx_pb2.Fee(amount=list(fee.amount), gas_limit=fee.gas_limit)
        elif isinstance(fee, dict):
            fee_msg = tx_pb2.Fee(
                amount=[coin_pb2.Coin(**c) for c in fee.get("amount", [])],
                gas_limit=int(fee.get("gas_limit", self.gas_limit)),
            )
        else:
            raise TypeError("Unsupported fee type")

        auth_info = tx_pb2.AuthInfo(signer_infos=[signer_info], fee=fee_msg)

        sign_doc = tx_pb2.SignDoc(
            body_bytes=tx_body.SerializeToString(),
            auth_info_bytes=auth_info.SerializeToString(),
            chain_id=self.chain_id,
            account_number=account_number,
        )

        signature = signer.sign(sign_doc.SerializeToString())

        return tx_pb2.TxRaw(
            body_bytes=tx_body.SerializeToString(),
            auth_info_bytes=auth_info.SerializeToString(),
            signatures=[signature],
        )


class TxBroadcaster:
    """Broadcast transactions via gRPC."""

    def __init__(self, channel: grpc.aio.Channel):
        self._stub = service_pb2_grpc.ServiceStub(channel)

    async def broadcast(
        self,
        tx: tx_pb2.TxRaw,
        mode: service_pb2.BroadcastMode = service_pb2.BROADCAST_MODE_SYNC,
    ) -> service_pb2.BroadcastTxResponse:
        request = service_pb2.BroadcastTxRequest(tx_bytes=tx.SerializeToString(), mode=mode)
        return await self._stub.BroadcastTx(request)
