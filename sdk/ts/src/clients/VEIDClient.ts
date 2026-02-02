import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ClientTxResult } from "./types.ts";

// VEID Types - these would come from generated protos
export type ScopeType =
  | "SCOPE_TYPE_UNSPECIFIED"
  | "SCOPE_TYPE_FACE"
  | "SCOPE_TYPE_DOCUMENT"
  | "SCOPE_TYPE_LIVENESS"
  | "SCOPE_TYPE_BIOMETRIC";

export type VerificationStatus =
  | "VERIFICATION_STATUS_UNSPECIFIED"
  | "VERIFICATION_STATUS_PENDING"
  | "VERIFICATION_STATUS_VERIFIED"
  | "VERIFICATION_STATUS_REJECTED"
  | "VERIFICATION_STATUS_BORDERLINE";

export type IdentityTier =
  | "IDENTITY_TIER_UNSPECIFIED"
  | "IDENTITY_TIER_NONE"
  | "IDENTITY_TIER_BASIC"
  | "IDENTITY_TIER_VERIFIED"
  | "IDENTITY_TIER_PREMIUM";

export interface Identity {
  address: string;
  walletId: string;
  score: number;
  tier: IdentityTier;
  scopes: ScopeReference[];
  createdAt: number;
  updatedAt: number;
}

export interface ScopeReference {
  scopeId: string;
  scopeType: ScopeType;
  status: VerificationStatus;
  envelopeHash: Uint8Array;
  addedAt: number;
}

export interface ScoreInfo {
  address: string;
  score: number;
  tier: IdentityTier;
  scoreVersion: string;
  lastUpdated: number;
}

export interface UploadScopeParams {
  scopeId: string;
  scopeType: ScopeType;
  encryptedPayload: {
    recipientFingerprints: string[];
    algorithm: string;
    ciphertext: Uint8Array;
    nonces: Uint8Array[];
    senderPubKey: Uint8Array;
    senderSignature: Uint8Array;
  };
  salt: Uint8Array;
  deviceFingerprint: string;
  clientId: string;
  clientSignature: Uint8Array;
  userSignature: Uint8Array;
  payloadHash: Uint8Array;
  captureTimestamp: number;
  geoHint?: string;
}

export interface EligibilityResult {
  eligible: boolean;
  requiredScore: number;
  currentScore: number;
  requiredTier: IdentityTier;
  currentTier: IdentityTier;
  missingScopes?: ScopeType[];
}

export interface VEIDClientDeps {
  sdk: unknown; // The generated SDK instance
}

/**
 * Client for VEID identity verification module
 */
export class VEIDClient extends BaseClient {
  private sdk: unknown;

  constructor(deps: VEIDClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get identity wallet for an address
   */
  async getIdentity(address: string): Promise<Identity | null> {
    const cacheKey = `veid:identity:${address}`;
    const cached = this.getCached<Identity>(cacheKey);
    if (cached) return cached;

    try {
      // This would call the actual SDK once VEID is generated
      // const result = await this.sdk.virtengine.veid.v1.getIdentityWallet({ address });
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "getIdentity");
    }
  }

  /**
   * Get identity score for an address
   */
  async getScore(address: string): Promise<ScoreInfo | null> {
    const cacheKey = `veid:score:${address}`;
    const cached = this.getCached<ScoreInfo>(cacheKey);
    if (cached) return cached;

    try {
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "getScore");
    }
  }

  /**
   * Verify if an address is eligible for an offering
   */
  async verifyEligibility(_address: string, _offeringId: string): Promise<EligibilityResult> {
    try {
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "verifyEligibility");
    }
  }

  /**
   * List scopes for an address
   */
  async listScopes(_address: string, _options?: { scopeType?: ScopeType }): Promise<ScopeReference[]> {
    try {
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "listScopes");
    }
  }

  /**
   * Upload an encrypted scope for verification
   */
  async uploadScope(_params: UploadScopeParams): Promise<ClientTxResult> {
    try {
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "uploadScope");
    }
  }

  /**
   * Request verification of a scope
   */
  async requestVerification(_scopeId: string): Promise<ClientTxResult> {
    try {
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "requestVerification");
    }
  }

  /**
   * Create a new identity wallet
   */
  async createIdentityWallet(_params: {
    bindingSignature: Uint8Array;
    bindingPubKey: Uint8Array;
    metadata?: Record<string, string>;
  }): Promise<ClientTxResult & { walletId: string }> {
    try {
      throw new Error("VEID module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "createIdentityWallet");
    }
  }
}
