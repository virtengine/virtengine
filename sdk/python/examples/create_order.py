import asyncio

from google.protobuf import any_pb2

from virtengine import VirtEngineClient
from virtengine.market.v1beta5 import tx_pb2


async def main() -> None:
    async with VirtEngineClient.from_mnemonic("your mnemonic here") as client:
        msg = tx_pb2.MsgCreateBid()
        msg_any = any_pb2.Any()
        msg_any.Pack(msg)
        response = await client.sign_and_broadcast([msg_any])
        print(response.tx_response.code)


if __name__ == "__main__":
    asyncio.run(main())
