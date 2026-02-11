import type {
  MsgAssignModerator,
  MsgEscalateFraudReport,
  MsgRejectFraudReport,
  MsgResolveFraudReport,
  MsgSubmitFraudReport,
  MsgUpdateReportStatus,
} from "../generated/protos/virtengine/fraud/v1/tx.ts";
import type {
  FraudAuditLog,
  FraudReport,
  ModeratorQueueEntry,
} from "../generated/protos/virtengine/fraud/v1/types.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";

export interface FraudClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Fraud module
 */
export class FraudClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: FraudClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  async getFraudReport(reportId: string): Promise<FraudReport | null> {
    try {
      const result = await this.sdk.virtengine.fraud.v1.getFraudReport({ reportId });
      return result.report ?? null;
    } catch (error) {
      this.handleQueryError(error, "getFraudReport");
    }
  }

  async listFraudReports(options?: ListOptions): Promise<FraudReport[]> {
    try {
      const result = await this.sdk.virtengine.fraud.v1.getFraudReports({
        pagination: toPageRequest(options),
      });
      return result.reports;
    } catch (error) {
      this.handleQueryError(error, "listFraudReports");
    }
  }

  async listFraudReportsByReporter(reporter: string, options?: ListOptions): Promise<FraudReport[]> {
    try {
      const result = await this.sdk.virtengine.fraud.v1.getFraudReportsByReporter({
        reporter,
        pagination: toPageRequest(options),
      });
      return result.reports;
    } catch (error) {
      this.handleQueryError(error, "listFraudReportsByReporter");
    }
  }

  async listFraudReportsByReportedParty(reportedParty: string, options?: ListOptions): Promise<FraudReport[]> {
    try {
      const result = await this.sdk.virtengine.fraud.v1.getFraudReportsByReportedParty({
        reportedParty,
        pagination: toPageRequest(options),
      });
      return result.reports;
    } catch (error) {
      this.handleQueryError(error, "listFraudReportsByReportedParty");
    }
  }

  async getAuditLog(reportId: string): Promise<FraudAuditLog[]> {
    try {
      const result = await this.sdk.virtengine.fraud.v1.getAuditLog({ reportId });
      return result.auditLogs;
    } catch (error) {
      this.handleQueryError(error, "getAuditLog");
    }
  }

  async getModeratorQueue(options?: ListOptions): Promise<ModeratorQueueEntry[]> {
    try {
      const result = await this.sdk.virtengine.fraud.v1.getModeratorQueue({
        pagination: toPageRequest(options),
      });
      return result.queueEntries;
    } catch (error) {
      this.handleQueryError(error, "getModeratorQueue");
    }
  }

  async submitFraudReport(params: MsgSubmitFraudReport, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.fraud.v1.submitFraudReport(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "submitFraudReport");
    }
  }

  async assignModerator(params: MsgAssignModerator, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.fraud.v1.assignModerator(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "assignModerator");
    }
  }

  async updateReportStatus(params: MsgUpdateReportStatus, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.fraud.v1.updateReportStatus(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "updateReportStatus");
    }
  }

  async resolveFraudReport(params: MsgResolveFraudReport, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.fraud.v1.resolveFraudReport(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "resolveFraudReport");
    }
  }

  async rejectFraudReport(params: MsgRejectFraudReport, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.fraud.v1.rejectFraudReport(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "rejectFraudReport");
    }
  }

  async escalateFraudReport(params: MsgEscalateFraudReport, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.fraud.v1.escalateFraudReport(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "escalateFraudReport");
    }
  }
}
