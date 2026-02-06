import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import { MemoryCache } from "../utils/cache.ts";
import type { RolesClientDeps } from "./RolesClient.ts";
import { RolesClient } from "./RolesClient.ts";

type MockFn = (...args: unknown[]) => Promise<unknown>;

describe("RolesClient", () => {
  let client: RolesClient;
  let deps: RolesClientDeps;

  beforeEach(() => {
    deps = {
      sdk: {
        virtengine: {
          roles: {
            v1: {
              getAccountRoles: jest.fn<MockFn>().mockResolvedValue({
                roles: [
                  { address: "virt1abc", role: "provider", grantedAt: "2024-01-01T00:00:00Z" },
                  { address: "virt1abc", role: "validator", grantedAt: "2024-01-02T00:00:00Z" },
                ],
              }),
              getHasRole: jest.fn<MockFn>().mockResolvedValue({ hasRole: true }),
              getRoleMembers: jest.fn<MockFn>().mockResolvedValue({
                members: [
                  { address: "virt1abc", role: "provider", grantedAt: "2024-01-01T00:00:00Z" },
                  { address: "virt1def", role: "provider", grantedAt: "2024-01-03T00:00:00Z" },
                ],
              }),
              getAccountState: jest.fn<MockFn>().mockResolvedValue({
                accountState: {
                  address: "virt1abc",
                  state: 1,
                  reason: "",
                  modifiedBy: "",
                  modifiedAt: 0,
                  previousState: 0,
                },
              }),
            },
          },
        },
      } as unknown as RolesClientDeps["sdk"],
    };
    client = new RolesClient(deps);
  });

  it("should create client instance", () => {
    expect(client).toBeInstanceOf(RolesClient);
  });

  it("fetches account roles", async () => {
    const roles = await client.getAccountRoles("virt1abc");
    expect(roles).toHaveLength(2);
    expect(roles[0].role).toBe("provider");
    expect(deps.sdk.virtengine.roles.v1.getAccountRoles).toHaveBeenCalledWith({ address: "virt1abc" });
  });

  it("caches account roles on subsequent calls", async () => {
    const cache = new MemoryCache({ ttlMs: 30000 });
    const client2 = new RolesClient(deps, { enableCaching: true, cache });
    await client2.getAccountRoles("virt1abc");
    await client2.getAccountRoles("virt1abc");
    expect(deps.sdk.virtengine.roles.v1.getAccountRoles).toHaveBeenCalledTimes(1);
  });

  it("checks if address has a role", async () => {
    const hasRole = await client.hasRole("virt1abc", "provider");
    expect(hasRole).toBe(true);
    expect(deps.sdk.virtengine.roles.v1.getHasRole).toHaveBeenCalledWith({
      address: "virt1abc",
      role: "provider",
    });
  });

  it("returns false when address does not have role", async () => {
    (deps.sdk.virtengine.roles.v1.getHasRole as jest.Mock<MockFn>)
      .mockResolvedValueOnce({ hasRole: false });
    const hasRole = await client.hasRole("virt1abc", "admin");
    expect(hasRole).toBe(false);
  });

  it("lists role members", async () => {
    const members = await client.listRoleMembers("provider");
    expect(members).toHaveLength(2);
    expect(deps.sdk.virtengine.roles.v1.getRoleMembers).toHaveBeenCalledWith(
      expect.objectContaining({ role: "provider" }),
    );
  });

  it("lists role members with pagination", async () => {
    await client.listRoleMembers("provider", { limit: 10, offset: 0 });
    expect(deps.sdk.virtengine.roles.v1.getRoleMembers).toHaveBeenCalledWith(
      expect.objectContaining({ role: "provider", pagination: expect.anything() }),
    );
  });

  it("fetches account state", async () => {
    const state = await client.getAccountState("virt1abc");
    expect(state?.state).toBe(1);
    expect(state?.address).toBe("virt1abc");
    expect(deps.sdk.virtengine.roles.v1.getAccountState).toHaveBeenCalledWith({ address: "virt1abc" });
  });

  it("returns null for missing account state", async () => {
    (deps.sdk.virtengine.roles.v1.getAccountState as jest.Mock<MockFn>)
      .mockResolvedValueOnce({ accountState: undefined });
    const state = await client.getAccountState("virt1missing");
    expect(state).toBeNull();
  });

  it("caches account state on subsequent calls", async () => {
    const cache = new MemoryCache({ ttlMs: 30000 });
    const client2 = new RolesClient(deps, { enableCaching: true, cache });
    await client2.getAccountState("virt1abc");
    await client2.getAccountState("virt1abc");
    expect(deps.sdk.virtengine.roles.v1.getAccountState).toHaveBeenCalledTimes(1);
  });
});
