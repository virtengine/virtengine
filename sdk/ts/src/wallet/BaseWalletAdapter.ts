import type { ChainInfo, SignedTx, UnsignedTx, WalletAccount, WalletAdapter, WalletState } from "./types.ts";

export abstract class BaseWalletAdapter implements WalletAdapter {
  abstract readonly name: string;
  abstract readonly icon?: string;

  protected _state: WalletState = "disconnected";
  protected _account: WalletAccount | null = null;
  protected _chainId: string | null = null;

  get state(): WalletState {
    return this._state;
  }

  get account(): WalletAccount | null {
    return this._account;
  }

  abstract isAvailable(): boolean;
  abstract connect(chainInfo: ChainInfo): Promise<void>;
  abstract signTx(tx: UnsignedTx, chainId: string): Promise<SignedTx>;

  async disconnect(): Promise<void> {
    this._state = "disconnected";
    this._account = null;
    this._chainId = null;
  }

  getAddress(): string {
    if (!this._account) {
      throw new Error(`${this.name}: Wallet not connected`);
    }
    return this._account.address;
  }

  protected setState(state: WalletState): void {
    this._state = state;
  }

  protected setAccount(account: WalletAccount | null): void {
    this._account = account;
  }
}
