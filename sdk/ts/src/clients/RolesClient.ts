import type { AccountStateRecord, RoleAssignment } from "../generated/protos/virtengine/roles/v1/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ListOptions } from "./types.ts";
import { toPageRequest } from "./types.ts";

export interface RolesClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Roles module (role-based access control)
 */
export class RolesClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: RolesClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get all roles assigned to an address
   */
  async getAccountRoles(address: string): Promise<RoleAssignment[]> {
    const cacheKey = `roles:account:${address}`;
    const cached = this.getCached<RoleAssignment[]>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.roles.v1.getAccountRoles({ address });
      this.setCached(cacheKey, result.roles);
      return result.roles;
    } catch (error) {
      this.handleQueryError(error, "getAccountRoles");
    }
  }

  /**
   * Check if an address has a specific role
   */
  async hasRole(address: string, role: string): Promise<boolean> {
    try {
      const result = await this.sdk.virtengine.roles.v1.getHasRole({ address, role });
      return result.hasRole;
    } catch (error) {
      this.handleQueryError(error, "hasRole");
    }
  }

  /**
   * List all members with a specific role
   */
  async listRoleMembers(role: string, options?: ListOptions): Promise<RoleAssignment[]> {
    const cacheKey = `roles:members:${role}:${options?.limit ?? ""}:${options?.offset ?? ""}:${options?.cursor ?? ""}`;
    const cached = this.getCached<RoleAssignment[]>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.roles.v1.getRoleMembers({
        role,
        pagination: toPageRequest(options),
      });
      this.setCached(cacheKey, result.members);
      return result.members;
    } catch (error) {
      this.handleQueryError(error, "listRoleMembers");
    }
  }

  /**
   * Get full account state including roles and flags
   */
  async getAccountState(address: string): Promise<AccountStateRecord | null> {
    const cacheKey = `roles:state:${address}`;
    const cached = this.getCached<AccountStateRecord>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.roles.v1.getAccountState({ address });
      if (result.accountState) {
        this.setCached(cacheKey, result.accountState);
      }
      return result.accountState ?? null;
    } catch (error) {
      this.handleQueryError(error, "getAccountState");
    }
  }
}
