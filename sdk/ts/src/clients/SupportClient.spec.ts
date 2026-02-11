import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { SupportClientDeps } from "./SupportClient.ts";
import { SupportClient } from "./SupportClient.ts";
import type {
  MsgAddSupportResponse,
  MsgArchiveSupportRequest,
  MsgCreateSupportRequest,
  MsgUpdateSupportRequest,
} from "./supportTypes.ts";

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

describe("SupportClient", () => {
  let client: SupportClient;
  let deps: SupportClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          support: {
            v1: {
              getSupportRequest: jest.fn<MockFn>().mockResolvedValue({ request: { ticketId: "ticket-1" } }),
              getSupportRequestsBySubmitter: jest.fn<MockFn>().mockResolvedValue({ requests: [{ ticketId: "ticket-1" }] }),
              getSupportResponsesByRequest: jest.fn<MockFn>().mockResolvedValue({ responses: [{ responseId: "resp-1" }] }),
              createSupportRequest: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              updateSupportRequest: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              addSupportResponse: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              archiveSupportRequest: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as SupportClientDeps["sdk"],
    };

    client = new SupportClient(deps);
  });

  it("fetches support request", async () => {
    const request = await client.getSupportRequest("ticket-1");
    expect(request).toBeTruthy();
  });

  it("lists support requests", async () => {
    const requests = await client.listSupportRequests("virt1");
    expect(requests).toHaveLength(1);
  });

  it("fetches support responses", async () => {
    const responses = await client.getSupportResponses("ticket-1");
    expect(responses).toHaveLength(1);
  });

  it("creates support request and returns tx metadata", async () => {
    const result = await client.createSupportRequest({} as MsgCreateSupportRequest);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("updates support request and returns tx metadata", async () => {
    const result = await client.updateSupportRequest({} as MsgUpdateSupportRequest);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("adds support response and returns tx metadata", async () => {
    const result = await client.addSupportResponse({} as MsgAddSupportResponse);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("archives support request and returns tx metadata", async () => {
    const result = await client.archiveSupportRequest({} as MsgArchiveSupportRequest);
    expect(result.transactionHash).toBe("TXHASH");
  });
});
