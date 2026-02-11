import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgAddPriceEntry } from "../generated/protos/virtengine/oracle/v1/msgs.ts";
import type { OracleClientDeps } from "./OracleClient.ts";
import { OracleClient } from "./OracleClient.ts";

type MockFn = (...args: unknown[]) => Promise<unknown>;

const txResponse = () => ({
  height: 1,
  transactionHash: "TXHASH",
  code: 0,
  rawLog: "",
  gasWanted: 100,
  gasUsed: 90,
  data: new Uint8Array(),
  events: [],
  eventsRaw: [],
  msgResponses: [],
});

describe("OracleClient", () => {
  let client: OracleClient;
  let deps: OracleClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          oracle: {
            v1: {
              getAggregatedPrice: jest.fn<MockFn>().mockResolvedValue({ aggregatedPrice: { denom: "uve", twap: "1.0" } }),
              getPrices: jest.fn<MockFn>().mockResolvedValue({ prices: [{ id: { denom: "uve" } }] }),
              addPriceEntry: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as OracleClientDeps["sdk"],
    };

    client = new OracleClient(deps);
  });

  it("fetches aggregated price", async () => {
    const price = await client.getPrice("uve");
    expect(price).toBeTruthy();
  });

  it("lists price data", async () => {
    const prices = await client.listPrices();
    expect(prices).toHaveLength(1);
  });

  it("gets exchange rate", async () => {
    const rate = await client.getExchangeRate("uve");
    expect(rate).toBe("1.0");
  });

  it("lists supported assets", async () => {
    const assets = await client.listSupportedAssets();
    expect(assets).toContain("uve");
  });

  it("adds price entry and returns tx metadata", async () => {
    const result = await client.addPriceEntry({} as MsgAddPriceEntry);
    expect(result.transactionHash).toBe("TXHASH");
  });
});
