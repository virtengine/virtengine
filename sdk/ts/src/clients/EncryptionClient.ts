import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult } from "./types.ts";
import { withTxResult } from "./types.ts";
import type { EncryptedPayloadEnvelope, RecipientKeyRecord } from "../generated/protos/virtengine/encryption/v1/types.ts";
import type { QueryValidateEnvelopeResponse } from "../generated/protos/virtengine/encryption/v1/query.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

export interface EncryptionClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Encryption module (key management and envelope validation)
 */
export class EncryptionClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: EncryptionClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get recipient keys for an address
   */
  async getRecipientKey(address: string): Promise<RecipientKeyRecord[]> {
    const cacheKey = `encryption:keys:${address}`;
    const cached = this.getCached<RecipientKeyRecord[]>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.encryption.v1.getRecipientKey({ address });
      this.setCached(cacheKey, result.keys);
      return result.keys;
    } catch (error) {
      this.handleQueryError(error, "getRecipientKey");
    }
  }

  /**
   * Get a specific key by fingerprint
   */
  async getKeyByFingerprint(fingerprint: string): Promise<RecipientKeyRecord | null> {
    const cacheKey = `encryption:key:${fingerprint}`;
    const cached = this.getCached<RecipientKeyRecord>(cacheKey);
    if (cached) return cached;

    try {
      const result = await this.sdk.virtengine.encryption.v1.getKeyByFingerprint({ fingerprint });
      if (result.key) {
        this.setCached(cacheKey, result.key);
      }
      return result.key ?? null;
    } catch (error) {
      this.handleQueryError(error, "getKeyByFingerprint");
    }
  }

  /**
   * Validate an encrypted envelope
   */
  async validateEnvelope(envelope: EncryptedPayloadEnvelope): Promise<QueryValidateEnvelopeResponse> {
    try {
      return await this.sdk.virtengine.encryption.v1.getValidateEnvelope({ envelope });
    } catch (error) {
      this.handleQueryError(error, "validateEnvelope");
    }
  }

  /**
   * Register a new encryption key
   */
  async registerKey(
    params: {
      sender: string;
      publicKey: Uint8Array;
      algorithmId: string;
      label?: string;
    },
    options?: TxCallOptions,
  ): Promise<ClientTxResult & { fingerprint: string }> {
    try {
      const { response, txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.encryption.v1.registerRecipientKey(
          {
            sender: params.sender,
            publicKey: params.publicKey,
            algorithmId: params.algorithmId,
            label: params.label ?? "",
          },
          txOptions,
        ), options);

      return {
        ...txResult,
        fingerprint: response.keyFingerprint,
      };
    } catch (error) {
      this.handleQueryError(error, "registerKey");
    }
  }

  /**
   * Update the label of a registered encryption key
   */
  async updateKeyLabel(
    params: {
      sender: string;
      keyFingerprint: string;
      label: string;
    },
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.encryption.v1.updateKeyLabel(
          {
            sender: params.sender,
            keyFingerprint: params.keyFingerprint,
            label: params.label,
          },
          txOptions,
        ), options);

      return txResult;
    } catch (error) {
      this.handleQueryError(error, "updateKeyLabel");
    }
  }

  /**
   * Revoke an encryption key
   */
  async revokeKey(
    params: { sender: string; keyFingerprint: string },
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.encryption.v1.revokeRecipientKey(
          {
            sender: params.sender,
            keyFingerprint: params.keyFingerprint,
          },
          txOptions,
        ), options);

      return txResult;
    } catch (error) {
      this.handleQueryError(error, "revokeKey");
    }
  }
}
