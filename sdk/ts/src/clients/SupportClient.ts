import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type {
  MsgAddSupportResponse,
  MsgArchiveSupportRequest,
  MsgCreateSupportRequest,
  MsgUpdateSupportRequest,
  SupportRequest,
  SupportResponse,
} from "./supportTypes.ts";
import type { ChainNodeSDK, ClientTxResult } from "./types.ts";
import { withTxResult } from "./types.ts";

export interface SupportClientDeps {
  sdk: ChainNodeSDK;
}

export interface SupportQueryOptions {
  viewerAddress?: string;
  viewerKeyId?: string;
}

interface SupportSDK {
  virtengine: {
    support: {
      v1: {
        getSupportRequest: (request: {
          ticketId: string;
          viewerAddress?: string;
          viewerKeyId?: string;
        }) => Promise<{ request?: SupportRequest }>;
        getSupportRequestsBySubmitter: (request: {
          submitterAddress: string;
          status?: string;
          viewerAddress?: string;
          viewerKeyId?: string;
        }) => Promise<{ requests: SupportRequest[] }>;
        getSupportResponsesByRequest: (request: {
          ticketId: string;
          viewerAddress?: string;
          viewerKeyId?: string;
        }) => Promise<{ responses: SupportResponse[] }>;
        createSupportRequest: (request: MsgCreateSupportRequest, options?: TxCallOptions) => Promise<unknown>;
        updateSupportRequest: (request: MsgUpdateSupportRequest, options?: TxCallOptions) => Promise<unknown>;
        addSupportResponse: (request: MsgAddSupportResponse, options?: TxCallOptions) => Promise<unknown>;
        archiveSupportRequest: (request: MsgArchiveSupportRequest, options?: TxCallOptions) => Promise<unknown>;
      };
    };
  };
}

/**
 * Client for Support module
 */
export class SupportClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: SupportClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  private get supportSdk() {
    return (this.sdk as unknown as SupportSDK).virtengine.support.v1;
  }

  async getSupportRequest(ticketId: string, options?: SupportQueryOptions): Promise<SupportRequest | null> {
    try {
      const result = await this.supportSdk.getSupportRequest({
        ticketId,
        viewerAddress: options?.viewerAddress,
        viewerKeyId: options?.viewerKeyId,
      });
      return result.request ?? null;
    } catch (error) {
      this.handleQueryError(error, "getSupportRequest");
    }
  }

  async listSupportRequests(
    submitterAddress: string,
    status?: string,
    options?: SupportQueryOptions,
  ): Promise<SupportRequest[]> {
    try {
      const result = await this.supportSdk.getSupportRequestsBySubmitter({
        submitterAddress,
        status,
        viewerAddress: options?.viewerAddress,
        viewerKeyId: options?.viewerKeyId,
      });
      return result.requests;
    } catch (error) {
      this.handleQueryError(error, "listSupportRequests");
    }
  }

  async getSupportResponses(ticketId: string, options?: SupportQueryOptions): Promise<SupportResponse[]> {
    try {
      const result = await this.supportSdk.getSupportResponsesByRequest({
        ticketId,
        viewerAddress: options?.viewerAddress,
        viewerKeyId: options?.viewerKeyId,
      });
      return result.responses;
    } catch (error) {
      this.handleQueryError(error, "getSupportResponses");
    }
  }

  async createSupportRequest(
    params: MsgCreateSupportRequest,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.supportSdk.createSupportRequest(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "createSupportRequest");
    }
  }

  async updateSupportRequest(
    params: MsgUpdateSupportRequest,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.supportSdk.updateSupportRequest(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "updateSupportRequest");
    }
  }

  async addSupportResponse(
    params: MsgAddSupportResponse,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.supportSdk.addSupportResponse(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "addSupportResponse");
    }
  }

  async archiveSupportRequest(
    params: MsgArchiveSupportRequest,
    options?: TxCallOptions,
  ): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.supportSdk.archiveSupportRequest(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "archiveSupportRequest");
    }
  }
}
