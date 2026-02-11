import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgDeleteReview, MsgSubmitReview } from "../generated/protos/virtengine/review/v1/tx.ts";
import type { ReviewClientDeps } from "./ReviewClient.ts";
import { ReviewClient } from "./ReviewClient.ts";

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

describe("ReviewClient", () => {
  let client: ReviewClient;
  let deps: ReviewClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          review: {
            v1: {
              submitReview: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              deleteReview: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as ReviewClientDeps["sdk"],
    };

    client = new ReviewClient(deps);
  });

  it("submits review", async () => {
    const result = await client.submitReview({} as MsgSubmitReview);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("deletes review", async () => {
    const result = await client.deleteReview({} as MsgDeleteReview);
    expect(result.transactionHash).toBe("TXHASH");
  });
});
