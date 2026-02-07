import asyncio

from virtengine import VirtEngineClient


async def main() -> None:
    async with VirtEngineClient() as client:
        subscriber = client.events()
        async for event in subscriber.subscribe("tm.event='NewBlock'"):
            print(event)


if __name__ == "__main__":
    asyncio.run(main())
