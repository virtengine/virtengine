import Long from "long";

import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";
import { IdentityTier } from "../generated/protos/virtengine/veid/v1/types.ts";
import type {
  IdentityRecord,
  IdentityScore,
  IdentityScope,
  ScopeType,
  VerificationStatus,
} from "../generated/protos/virtengine/veid/v1/types.ts";
import type {
  MsgCreateIdentityWallet,
  MsgRequestVerification,
  MsgUploadScope,
} from "../generated/protos/virtengine/veid/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

export interface VEIDClientDeps {
  sdk: ChainNodeSDK;
}

export interface EligibilityResult {
  eligible: boolean;
  requiredScore: number;
  currentScore: number;
  requiredTier: IdentityTier;
  currentTier: IdentityTier;
  missingScopes?: ScopeType[];
}

export interface ScopeListOptions {
  scopeType?: ScopeType;
  statusFilter?: VerificationStatus;
}

/**
 * Client for VEID identity verification module
 */
export class VEIDClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: VEIDClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get identity record for an address
   */
  async getIdentity(address: string): Promise<IdentityRecord | null> {
    const cacheKey = `veid:identity:${address}`;
    const cached = this.getCached<IdentityRecord>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.veid.v1.getIdentity({ accountAddress: address });
      if (!result.found) return null;
      this.setCached(cacheKey, result.identity);
      return result.identity;
    } catch (error) {
      this.handleQueryError(error, "getIdentity");
    }
  }

  /**
   * Get identity score for an address
   */
  async getScore(address: string): Promise<IdentityScore | null> {
    const cacheKey = `veid:score:${address}`;
    const cached = this.getCached<IdentityScore>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.veid.v1.getIdentityScore({ accountAddress: address });
      if (!result.found) return null;
      this.setCached(cacheKey, result.score);
      return result.score;
    } catch (error) {
      this.handleQueryError(error, "getScore");
    }
  }

  /**
   * Verify if an address is eligible for an offering
   */
  async verifyEligibility(address: string, offeringId: string): Promise<EligibilityResult> {
    try {
      const [scoreResult, offeringResult] = await Promise.all([
        this.sdk.virtengine.veid.v1.getIdentityScore({ accountAddress: address }),
        this.sdk.virtengine.hpc.v1.getOffering({ offeringId }),
      ]);

      const currentScore = scoreResult.found ? scoreResult.score.score : 0;
      const currentTier = scoreResult.found ? scoreResult.score.tier : IdentityTier.IDENTITY_TIER_UNVERIFIED;
      const requiredScore = offeringResult.offering?.requiredIdentityThreshold ?? 0;

      return {
        eligible: scoreResult.found && currentScore >= requiredScore,
        requiredScore,
        currentScore,
        requiredTier: IdentityTier.IDENTITY_TIER_UNVERIFIED,
        currentTier,
        missingScopes: [],
      };
    } catch (error) {
      this.handleQueryError(error, "verifyEligibility");
    }
  }

  /**
   * List scopes for an address
   */
  async listScopes(
    address: string,
    options?: ListOptions & ScopeListOptions,
  ): Promise<IdentityScope[]> {
    const cacheKey = `veid:scopes:${address}:${options?.scopeType ?? ""}:${options?.statusFilter ?? ""}:${options?.limit ?? ""}:${options?.offset ?? ""}:${options?.cursor ?? ""}`;
    const cached = this.getCached<IdentityScope[]>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.veid.v1.getScopes({
        accountAddress: address,
        scopeType: options?.scopeType ?? 0,
        statusFilter: options?.statusFilter ?? 0,
        pagination: toPageRequest(options),
      });
      this.setCached(cacheKey, result.scopes);
      return result.scopes;
    } catch (error) {
      this.handleQueryError(error, "listScopes");
    }
  }

  /**
   * Upload an encrypted scope for verification
   */
  async uploadScope(
    params: MsgUploadScope,
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { scopeId: string; status: VerificationStatus; uploadedAt: Long }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.veid.v1.uploadScope(params, txOptions), options);

      return {
        ...txResult,
        scopeId: response.scopeId,
        status: response.status,
        uploadedAt: response.uploadedAt,
      };
    } catch (error) {
      this.handleQueryError(error, "uploadScope");
    }
  }

  /**
   * Request verification of a scope
   */
  async requestVerification(
    params: MsgRequestVerification,
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { scopeId: string; status: VerificationStatus; requestedAt: Long }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.veid.v1.requestVerification(params, txOptions), options);

      return {
        ...txResult,
        scopeId: response.scopeId,
        status: response.status,
        requestedAt: response.requestedAt,
      };
    } catch (error) {
      this.handleQueryError(error, "requestVerification");
    }
  }

  /**
   * Create a new identity wallet
   */
  async createIdentityWallet(
    params: MsgCreateIdentityWallet,
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { walletId: string }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.veid.v1.createIdentityWallet(params, txOptions), options);

      return {
        ...txResult,
        walletId: response.walletId,
      };
    } catch (error) {
      this.handleQueryError(error, "createIdentityWallet");
    }
  }
}
