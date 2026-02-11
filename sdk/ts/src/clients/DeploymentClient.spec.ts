import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import type { MsgCloseDeployment, MsgCreateDeployment, MsgUpdateDeployment } from "../generated/protos/virtengine/deployment/v1beta4/deploymentmsg.ts";
import type { DeploymentClientDeps } from "./DeploymentClient.ts";
import { DeploymentClient } from "./DeploymentClient.ts";

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

describe("DeploymentClient", () => {
  let client: DeploymentClient;
  let deps: DeploymentClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          deployment: {
            v1beta4: {
              getDeployment: jest.fn<MockFn>().mockResolvedValue({ deployment: { id: {} } }),
              getDeployments: jest.fn<MockFn>().mockResolvedValue({ deployments: [{ deployment: { id: {} } }] }),
              getGroup: jest.fn<MockFn>().mockResolvedValue({ group: { groupId: {} } }),
              createDeployment: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              updateDeployment: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
              closeDeployment: jest.fn<MockFn>().mockImplementation((...args: unknown[]) => {
                const options = args[1] as Record<string, (...a: unknown[]) => void> | undefined;
                options?.afterBroadcast?.(txResponse());
                return Promise.resolve({});
              }),
            },
          },
        },
      } as unknown as DeploymentClientDeps["sdk"],
    };

    client = new DeploymentClient(deps);
  });

  it("fetches a deployment", async () => {
    const deployment = await client.getDeployment(
      {} as Parameters<DeploymentClient["getDeployment"]>[0],
    );
    expect(deployment).toBeTruthy();
  });

  it("lists deployments", async () => {
    const deployments = await client.listDeployments();
    expect(deployments).toHaveLength(1);
  });

  it("fetches deployment group", async () => {
    const group = await client.getDeploymentGroup(
      {} as Parameters<DeploymentClient["getDeploymentGroup"]>[0],
    );
    expect(group).toBeTruthy();
  });

  it("creates deployment and returns tx metadata", async () => {
    const result = await client.createDeployment({} as MsgCreateDeployment);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("updates deployment and returns tx metadata", async () => {
    const result = await client.updateDeployment({} as MsgUpdateDeployment);
    expect(result.transactionHash).toBe("TXHASH");
  });

  it("closes deployment and returns tx metadata", async () => {
    const result = await client.closeDeployment({} as MsgCloseDeployment);
    expect(result.transactionHash).toBe("TXHASH");
  });
});
