import { beforeEach, describe, expect, it } from "@jest/globals";

import { HPCClient, type HPCClientDeps } from "./HPCClient.ts";

describe("HPCClient", () => {
  let client: HPCClient;

  beforeEach(() => {
    const deps: HPCClientDeps = {
      sdk: {},
    };
    client = new HPCClient(deps);
  });

  describe("constructor", () => {
    it("should create client instance", () => {
      expect(client).toBeInstanceOf(HPCClient);
    });
  });

  describe("Cluster queries", () => {
    describe("getCluster", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.getCluster("cluster-1")).rejects.toThrow(
          "HPC module not yet generated",
        );
      });
    });

    describe("listClusters", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.listClusters()).rejects.toThrow(
          "HPC module not yet generated",
        );
      });

      it("should accept region filter", async () => {
        await expect(client.listClusters({ region: "us-east" })).rejects.toThrow(
          "HPC module not yet generated",
        );
      });
    });
  });

  describe("Offering queries", () => {
    describe("getOffering", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.getOffering("offering-1")).rejects.toThrow(
          "HPC module not yet generated",
        );
      });
    });

    describe("listOfferings", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.listOfferings()).rejects.toThrow(
          "HPC module not yet generated",
        );
      });

      it("should accept filters", async () => {
        await expect(
          client.listOfferings({ clusterId: "cluster-1", activeOnly: true }),
        ).rejects.toThrow("HPC module not yet generated");
      });
    });
  });

  describe("Job queries", () => {
    describe("getJob", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.getJob("job-1")).rejects.toThrow(
          "HPC module not yet generated",
        );
      });
    });

    describe("listJobs", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.listJobs()).rejects.toThrow(
          "HPC module not yet generated",
        );
      });

      it("should accept state filter", async () => {
        await expect(
          client.listJobs({ state: "JOB_STATE_RUNNING" }),
        ).rejects.toThrow("HPC module not yet generated");
      });
    });
  });

  describe("Transaction methods", () => {
    describe("submitJob", () => {
      it("should throw error indicating proto generation needed", async () => {
        const params = {
          offeringId: "offering-1",
          queueName: "default",
          workloadSpec: {
            containerImage: "python:3.11",
            command: ["python", "-c", "print('hello')"],
          },
          resources: {
            cpuCores: 2,
            memoryMb: 4096,
          },
          maxRuntimeSeconds: 3600,
          maxPrice: [{ denom: "uvirt", amount: "1000000" }],
        };
        await expect(client.submitJob(params)).rejects.toThrow(
          "HPC module not yet generated",
        );
      });
    });

    describe("cancelJob", () => {
      it("should throw error indicating proto generation needed", async () => {
        await expect(client.cancelJob("job-1")).rejects.toThrow(
          "HPC module not yet generated",
        );
      });

      it("should accept optional reason", async () => {
        await expect(
          client.cancelJob("job-1", "User requested cancellation"),
        ).rejects.toThrow("HPC module not yet generated");
      });
    });

    describe("registerCluster", () => {
      it("should throw error indicating proto generation needed", async () => {
        const params = {
          name: "Test Cluster",
          region: "us-east",
          totalNodes: 10,
          slurmVersion: "23.02",
        };
        await expect(client.registerCluster(params)).rejects.toThrow(
          "HPC module not yet generated",
        );
      });
    });
  });
});
