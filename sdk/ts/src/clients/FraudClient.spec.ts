import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type {
  MsgAssignModerator,
  MsgEscalateFraudReport,
  MsgRejectFraudReport,
  MsgResolveFraudReport,
  MsgSubmitFraudReport,
  MsgUpdateReportStatus,
} from "../generated/protos/virtengine/fraud/v1/tx.ts";
import type { FraudClientDeps } from "./FraudClient.ts";
import { FraudClient } from "./FraudClient.ts";

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

describe("FraudClient", () => {
  let client: FraudClient;
  let deps: FraudClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          fraud: {
            v1: {
              getFraudReport: jest.fn<MockFn>().mockResolvedValue({ report: { reportId: "report-1" } }),
              getFraudReports: jest.fn<MockFn>().mockResolvedValue({ reports: [{ reportId: "report-1" }] }),
              getFraudReportsByReporter: jest.fn<MockFn>().mockResolvedValue({ reports: [{ reportId: "report-1" }] }),
              getFraudReportsByReportedParty: jest.fn<MockFn>().mockResolvedValue({ reports: [{ reportId: "report-1" }] }),
              getAuditLog: jest.fn<MockFn>().mockResolvedValue({ auditLogs: [{ action: "created" }] }),
              getModeratorQueue: jest.fn<MockFn>().mockResolvedValue({ queueEntries: [{ reportId: "report-1" }] }),
              submitFraudReport: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              assignModerator: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              updateReportStatus: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              resolveFraudReport: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              rejectFraudReport: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              escalateFraudReport: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as FraudClientDeps["sdk"],
    };

    client = new FraudClient(deps);
  });

  it("fetches fraud report", async () => {
    const report = await client.getFraudReport("report-1");
    expect(report).toBeTruthy();
  });

  it("lists fraud reports", async () => {
    const reports = await client.listFraudReports();
    expect(reports).toHaveLength(1);
  });

  it("lists fraud reports by reporter", async () => {
    const reports = await client.listFraudReportsByReporter("virt1");
    expect(reports).toHaveLength(1);
  });

  it("lists fraud reports by reported party", async () => {
    const reports = await client.listFraudReportsByReportedParty("virt1");
    expect(reports).toHaveLength(1);
  });

  it("fetches audit log", async () => {
    const entries = await client.getAuditLog("report-1");
    expect(entries).toHaveLength(1);
  });

  it("fetches moderator queue", async () => {
    const reports = await client.getModeratorQueue();
    expect(reports).toHaveLength(1);
  });

  it("submits fraud report", async () => {
    const result = await client.submitFraudReport({} as MsgSubmitFraudReport);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("assigns moderator", async () => {
    const result = await client.assignModerator({} as MsgAssignModerator);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("updates report status", async () => {
    const result = await client.updateReportStatus({} as MsgUpdateReportStatus);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("resolves fraud report", async () => {
    const result = await client.resolveFraudReport({} as MsgResolveFraudReport);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("rejects fraud report", async () => {
    const result = await client.rejectFraudReport({} as MsgRejectFraudReport);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("escalates fraud report", async () => {
    const result = await client.escalateFraudReport({} as MsgEscalateFraudReport);
    expect(result.transactionHash).toBe("TXHASH");
  });
});
