import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ClientTxResult } from "./types.ts";

export type FactorType =
  | "FACTOR_TYPE_UNSPECIFIED"
  | "FACTOR_TYPE_TOTP"
  | "FACTOR_TYPE_FIDO2"
  | "FACTOR_TYPE_SMS"
  | "FACTOR_TYPE_EMAIL"
  | "FACTOR_TYPE_HARDWARE_KEY";

export interface MFAPolicy {
  address: string;
  enabled: boolean;
  requiredFactors: number;
  allowedFactorTypes: FactorType[];
}

export interface FactorEnrollment {
  factorId: string;
  factorType: FactorType;
  label: string;
  enrolledAt: number;
  lastUsedAt?: number;
}

export interface Challenge {
  challengeId: string;
  factorType: FactorType;
  challengeData?: Uint8Array;
  expiresAt: number;
}

export interface MFAClientDeps {
  sdk: unknown;
}

/**
 * Client for MFA multi-factor authentication module
 */
export class MFAClient extends BaseClient {
  private sdk: unknown;

  constructor(deps: MFAClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get MFA policy for an address
   */
  async getPolicy(_address: string): Promise<MFAPolicy | null> {
    try {
      throw new Error("MFA module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "getPolicy");
    }
  }

  /**
   * List enrolled factors for an address
   */
  async listEnrollments(_address: string): Promise<FactorEnrollment[]> {
    try {
      throw new Error("MFA module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "listEnrollments");
    }
  }

  /**
   * Enroll a new authentication factor
   */
  async enrollFactor(_params: {
    factorType: FactorType;
    label: string;
    publicIdentifier?: Uint8Array;
    initialVerificationProof?: Uint8Array;
  }): Promise<ClientTxResult & { factorId: string }> {
    try {
      throw new Error("MFA module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "enrollFactor");
    }
  }

  /**
   * Create an MFA challenge
   */
  async createChallenge(_factorType: FactorType): Promise<Challenge> {
    try {
      throw new Error("MFA module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "createChallenge");
    }
  }

  /**
   * Verify an MFA challenge response
   */
  async verifyChallenge(_challengeId: string, _response: Uint8Array): Promise<ClientTxResult & { sessionId: string }> {
    try {
      throw new Error("MFA module not yet generated - proto generation needed");
    } catch (error) {
      this.handleQueryError(error, "verifyChallenge");
    }
  }
}
