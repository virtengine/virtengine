import { beforeEach, describe, expect, it, jest } from "@jest/globals";
import Long from "long";

import type { HPCClientDeps } from "./HPCClient.ts";
import { HPCClient } from "./HPCClient.ts";

type MockFn = (...args: unknown[]) => Promise<unknown>;

const txResponse = () => ({
  height: 1,
  transactionHash: "ABC",
  code: 0,
  rawLog: "",
  gasWanted: 100,
  gasUsed: 90,
  data: new Uint8Array(),
  events: [],
  eventsRaw: [],
  msgResponses: [],
});

describe("HPCClient", () => {
  let client: HPCClient;
  let deps: HPCClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          hpc: {
            v1: {
              getCluster: jest.fn<MockFn>().mockResolvedValue({
                cluster: { clusterId: "cluster-1" },
              }),
              getClusters: jest.fn<MockFn>().mockResolvedValue({
                clusters: [{ clusterId: "cluster-1" }],
              }),
              getOffering: jest.fn<MockFn>().mockResolvedValue({
                offering: { offeringId: "offering-1", requiredIdentityThreshold: 50, maxRuntimeSeconds: Long.ZERO },
              }),
              getOfferings: jest.fn<MockFn>().mockResolvedValue({
                offerings: [{ offeringId: "offering-1" }],
              }),
              getOfferingsByCluster: jest.fn<MockFn>().mockResolvedValue({
                offerings: [{ offeringId: "offering-2" }],
              }),
              getJob: jest.fn<MockFn>().mockResolvedValue({
                job: { jobId: "job-1" },
              }),
              getJobs: jest.fn<MockFn>().mockResolvedValue({
                jobs: [{ jobId: "job-1" }],
              }),
              getJobsByCustomer: jest.fn<MockFn>().mockResolvedValue({
                jobs: [{ jobId: "job-2" }],
              }),
              getJobsByProvider: jest.fn<MockFn>().mockResolvedValue({
                jobs: [{ jobId: "job-3" }],
              }),
              submitJob: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({ jobId: "job-1", escrowId: "escrow-1" });
              }),
              cancelJob: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              registerCluster: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({ clusterId: "cluster-99" });
              }),
            },
          },
        },
      } as unknown as HPCClientDeps["sdk"],
    };
    client = new HPCClient(deps);
  });

  it("should create client instance", () => {
    expect(client).toBeInstanceOf(HPCClient);
  });

  it("fetches a cluster", async () => {
    const result = await client.getCluster("cluster-1");
    expect(result?.clusterId).toBe("cluster-1");
    expect(deps.sdk.virtengine.hpc.v1.getCluster).toHaveBeenCalledWith({ clusterId: "cluster-1" });
  });

  it("lists clusters", async () => {
    const result = await client.listClusters({ region: "us-east" });
    expect(result).toHaveLength(1);
    expect(deps.sdk.virtengine.hpc.v1.getClusters).toHaveBeenCalled();
  });

  it("lists offerings by cluster when filter present", async () => {
    const result = await client.listOfferings({ clusterId: "cluster-1" });
    expect(result[0].offeringId).toBe("offering-2");
  });

  it("lists jobs by customer when customer filter present", async () => {
    const result = await client.listJobs({ customerAddress: "virt1abc" });
    expect(result[0].jobId).toBe("job-2");
  });

  it("submits job and returns tx metadata", async () => {
    const result = await client.submitJob({
      customerAddress: "virt1abc",
      offeringId: "offering-1",
      queueName: "default",
      workloadSpec: undefined,
      resources: undefined,
      dataReferences: [],
      encryptedInputsPointer: "",
      encryptedOutputsPointer: "",
      maxRuntimeSeconds: Long.ZERO,
      maxPrice: [],
    });
    expect(result.jobId).toBe("job-1");
    expect(result.transactionHash).toBe("ABC");
  });

  it("cancels job and returns tx metadata", async () => {
    const result = await client.cancelJob({ requesterAddress: "virt1abc", jobId: "job-1", reason: "" });
    expect(result.transactionHash).toBe("ABC");
  });

  it("registers cluster and returns cluster id", async () => {
    const result = await client.registerCluster({
      providerAddress: "virt1provider",
      name: "Test",
      description: "",
      region: "us-east",
      partitions: [],
      totalNodes: 1,
      clusterMetadata: undefined,
      slurmVersion: "23.02",
    });
    expect(result.clusterId).toBe("cluster-99");
  });
});
