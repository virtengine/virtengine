import Long from "long";

import type {
  MsgProposeMeasurement,
  MsgRegisterEnclaveIdentity,
  MsgRevokeMeasurement,
  MsgRotateEnclaveIdentity,
} from "../generated/protos/virtengine/enclave/v1/tx.ts";
import type {
  AttestedScoringResult,
  EnclaveIdentity,
  KeyRotationRecord,
  MeasurementRecord,
  ValidatorKeyInfo,
} from "../generated/protos/virtengine/enclave/v1/types.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";

export interface EnclaveClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Enclave module
 */
export class EnclaveClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: EnclaveClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  async getEnclaveIdentity(validatorAddress: string): Promise<EnclaveIdentity | null> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getEnclaveIdentity({ validatorAddress });
      return result.identity ?? null;
    } catch (error) {
      this.handleQueryError(error, "getEnclaveIdentity");
    }
  }

  async listActiveValidatorEnclaveKeys(options?: ListOptions): Promise<EnclaveIdentity[]> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getActiveValidatorEnclaveKeys({
        pagination: toPageRequest(options),
      });
      return result.identities;
    } catch (error) {
      this.handleQueryError(error, "listActiveValidatorEnclaveKeys");
    }
  }

  async getCommitteeEnclaveKeys(committeeEpoch: number | Long): Promise<EnclaveIdentity[]> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getCommitteeEnclaveKeys({
        committeeEpoch: Long.fromValue(committeeEpoch),
      });
      return result.identities;
    } catch (error) {
      this.handleQueryError(error, "getCommitteeEnclaveKeys");
    }
  }

  async listMeasurementAllowlist(options?: ListOptions): Promise<MeasurementRecord[]> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getMeasurementAllowlist({
        pagination: toPageRequest(options),
      });
      return result.measurements;
    } catch (error) {
      this.handleQueryError(error, "listMeasurementAllowlist");
    }
  }

  async getMeasurement(measurementHash: string): Promise<MeasurementRecord | null> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getMeasurement({ measurementHash });
      return result.measurement ?? null;
    } catch (error) {
      this.handleQueryError(error, "getMeasurement");
    }
  }

  async getKeyRotation(validatorAddress: string): Promise<KeyRotationRecord | null> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getKeyRotation({ validatorAddress });
      return result.rotation ?? null;
    } catch (error) {
      this.handleQueryError(error, "getKeyRotation");
    }
  }

  async getValidKeySet(): Promise<ValidatorKeyInfo[]> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getValidKeySet({});
      return result.validatorKeys;
    } catch (error) {
      this.handleQueryError(error, "getValidKeySet");
    }
  }

  async getAttestedResult(blockHeight: number | Long, scopeId: string): Promise<AttestedScoringResult | null> {
    try {
      const result = await this.sdk.virtengine.enclave.v1.getAttestedResult({
        blockHeight: Long.fromValue(blockHeight),
        scopeId,
      });
      return result.result ?? null;
    } catch (error) {
      this.handleQueryError(error, "getAttestedResult");
    }
  }

  async registerEnclaveIdentity(
    params: MsgRegisterEnclaveIdentity,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.enclave.v1.registerEnclaveIdentity(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "registerEnclaveIdentity");
    }
  }

  async rotateEnclaveIdentity(
    params: MsgRotateEnclaveIdentity,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.enclave.v1.rotateEnclaveIdentity(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "rotateEnclaveIdentity");
    }
  }

  async proposeMeasurement(
    params: MsgProposeMeasurement,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.enclave.v1.proposeMeasurement(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "proposeMeasurement");
    }
  }

  async revokeMeasurement(
    params: MsgRevokeMeasurement,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.enclave.v1.revokeMeasurement(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "revokeMeasurement");
    }
  }
}
