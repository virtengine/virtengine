import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";
import type { Account } from "../generated/protos/virtengine/escrow/types/v1/account.ts";
import type { Payment } from "../generated/protos/virtengine/escrow/types/v1/payment.ts";
import type { MsgAccountDeposit } from "../generated/protos/virtengine/escrow/v1/msg.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

export interface EscrowClientDeps {
  sdk: ChainNodeSDK;
}

export interface EscrowAccountFilters {
  state?: string;
  xid?: string;
}

export interface EscrowPaymentFilters {
  state?: string;
  xid?: string;
}

/**
 * Client for Escrow module (payment escrow management)
 */
export class EscrowClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: EscrowClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get escrow account by ID (xid)
   */
  async getAccount(xid: string): Promise<Account | null> {
    try {
      const result = await this.sdk.virtengine.escrow.v1.getAccounts({
        xid,
        state: "",
        pagination: toPageRequest({ limit: 1 }),
      });
      return result.accounts[0] ?? null;
    } catch (error) {
      this.handleQueryError(error, "getAccount");
    }
  }

  /**
   * List escrow accounts
   */
  async listAccounts(options?: ListOptions & EscrowAccountFilters): Promise<Account[]> {
    try {
      const result = await this.sdk.virtengine.escrow.v1.getAccounts({
        state: options?.state ?? "",
        xid: options?.xid ?? "",
        pagination: toPageRequest(options),
      });
      return result.accounts;
    } catch (error) {
      this.handleQueryError(error, "listAccounts");
    }
  }

  /**
   * Get payments for an escrow account
   */
  async getPayments(xid: string, options?: ListOptions): Promise<Payment[]> {
    try {
      const result = await this.sdk.virtengine.escrow.v1.getPayments({
        xid,
        state: "",
        pagination: toPageRequest(options),
      });
      return result.payments;
    } catch (error) {
      this.handleQueryError(error, "getPayments");
    }
  }

  /**
   * Deposit funds into an escrow account
   */
  async deposit(params: MsgAccountDeposit, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.escrow.v1.accountDeposit(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "deposit");
    }
  }
}
