import Long from "long";

import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";
import type {
  Challenge,
  ChallengeResponse,
  FactorEnrollment,
  FactorEnrollmentStatus,
  FactorMetadata,
  FactorType,
  MFAPolicy,
  MFAProof,
  ClientInfo,
  SensitiveTransactionType,
} from "../generated/protos/virtengine/mfa/v1/types.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

export interface MFAClientDeps {
  sdk: ChainNodeSDK;
}

export interface EnrollmentFilters {
  factorTypeFilter?: FactorType;
  statusFilter?: FactorEnrollmentStatus;
}

/**
 * Client for MFA multi-factor authentication module
 */
export class MFAClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: MFAClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get MFA policy for an address
   */
  async getPolicy(address: string): Promise<MFAPolicy | null> {
    const cacheKey = `mfa:policy:${address}`;
    const cached = this.getCached<MFAPolicy>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.mfa.v1.getMFAPolicy({ address });
      if (!result.found) return null;
      this.setCached(cacheKey, result.policy);
      return result.policy;
    } catch (error) {
      this.handleQueryError(error, "getPolicy");
    }
  }

  /**
   * List enrolled factors for an address
   */
  async listEnrollments(
    address: string,
    options?: ListOptions & EnrollmentFilters,
  ): Promise<FactorEnrollment[]> {
    const cacheKey = `mfa:enrollments:${address}:${options?.factorTypeFilter ?? ""}:${options?.statusFilter ?? ""}:${options?.limit ?? ""}:${options?.offset ?? ""}:${options?.cursor ?? ""}`;
    const cached = this.getCached<FactorEnrollment[]>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.mfa.v1.getFactorEnrollments({
        address,
        factorTypeFilter: options?.factorTypeFilter ?? 0,
        statusFilter: options?.statusFilter ?? 0,
        pagination: toPageRequest(options),
      });
      this.setCached(cacheKey, result.enrollments);
      return result.enrollments;
    } catch (error) {
      this.handleQueryError(error, "listEnrollments");
    }
  }

  /**
   * Enroll a new authentication factor
   */
  async enrollFactor(
    params: {
      sender: string;
      factorType: FactorType;
      label: string;
      publicIdentifier?: Uint8Array;
      metadata?: FactorMetadata;
      initialVerificationProof?: Uint8Array;
    },
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { factorId: string; status: FactorEnrollmentStatus }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.mfa.v1.enrollFactor(
          {
            sender: params.sender,
            factorType: params.factorType,
            label: params.label,
            publicIdentifier: params.publicIdentifier ?? new Uint8Array(),
            metadata: params.metadata,
            initialVerificationProof: params.initialVerificationProof ?? new Uint8Array(),
          },
          txOptions,
        ), options);

      return {
        ...txResult,
        factorId: response.factorId,
        status: response.status,
      };
    } catch (error) {
      this.handleQueryError(error, "enrollFactor");
    }
  }

  /**
   * Create an MFA challenge
   */
  async createChallenge(
    params: {
      sender: string;
      factorType: FactorType;
      transactionType: SensitiveTransactionType;
      factorId?: string;
      clientInfo?: ClientInfo;
    },
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { challengeId: string; challengeData: Uint8Array; expiresAt: Long }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.mfa.v1.createChallenge(
          {
            sender: params.sender,
            factorType: params.factorType,
            factorId: params.factorId ?? "",
            transactionType: params.transactionType,
            clientInfo: params.clientInfo,
          },
          txOptions,
        ), options);

      return {
        ...txResult,
        challengeId: response.challengeId,
        challengeData: response.challengeData,
        expiresAt: response.expiresAt,
      };
    } catch (error) {
      this.handleQueryError(error, "createChallenge");
    }
  }

  /**
   * Verify an MFA challenge response
   */
  async verifyChallenge(
    params: {
      sender: string;
      challengeId: string;
      response: ChallengeResponse;
    },
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { verified: boolean; sessionId: string; sessionExpiresAt: Long; remainingFactors: FactorType[] }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.mfa.v1.verifyChallenge(
          {
            sender: params.sender,
            challengeId: params.challengeId,
            response: params.response,
          },
          txOptions,
        ), options);

      return {
        ...txResult,
        verified: response.verified,
        sessionId: response.sessionId,
        sessionExpiresAt: response.sessionExpiresAt,
        remainingFactors: response.remainingFactors,
      };
    } catch (error) {
      this.handleQueryError(error, "verifyChallenge");
    }
  }

  /**
   * Get a challenge by ID
   */
  async getChallenge(challengeId: string): Promise<Challenge | null> {
    try {
      const result = await this.sdk.virtengine.mfa.v1.getChallenge({ challengeId });
      return result.found ? result.challenge : null;
    } catch (error) {
      this.handleQueryError(error, "getChallenge");
    }
  }

  /**
   * Update MFA policy for an account
   */
  async setPolicy(
    params: { sender: string; policy: MFAPolicy; mfaProof?: MFAProof },
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { success: boolean }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.mfa.v1.setMFAPolicy(
          {
            sender: params.sender,
            policy: params.policy,
            mfaProof: params.mfaProof,
          },
          txOptions,
        ), options);

      return {
        ...txResult,
        success: response.success,
      };
    } catch (error) {
      this.handleQueryError(error, "setPolicy");
    }
  }
}
