import { CosmostationAdapter } from "./CosmostationAdapter.ts";
import { KeplrAdapter } from "./KeplrAdapter.ts";
import { LeapAdapter } from "./LeapAdapter.ts";
import type { ChainInfo, WalletAccount, WalletAdapter } from "./types.ts";

export type WalletType = "keplr" | "leap" | "cosmostation";

export interface WalletManagerOptions {
  autoConnect?: boolean;
  preferredWallet?: WalletType;
}

export class WalletManager {
  private adapters: Map<WalletType, WalletAdapter>;
  private currentAdapter: WalletAdapter | null = null;
  private chainInfo: ChainInfo | null = null;

  constructor(_options?: WalletManagerOptions) {
    this.adapters = new Map<WalletType, WalletAdapter>();
    this.adapters.set("keplr", new KeplrAdapter());
    this.adapters.set("leap", new LeapAdapter());
    this.adapters.set("cosmostation", new CosmostationAdapter());
  }

  getAvailableWallets(): WalletType[] {
    const available: WalletType[] = [];
    for (const [type, adapter] of this.adapters) {
      if (adapter.isAvailable()) {
        available.push(type);
      }
    }
    return available;
  }

  getAdapter(type: WalletType): WalletAdapter | undefined {
    return this.adapters.get(type);
  }

  getCurrentAdapter(): WalletAdapter | null {
    return this.currentAdapter;
  }

  getChainInfo(): ChainInfo | null {
    return this.chainInfo;
  }

  async connect(type: WalletType, chainInfo: ChainInfo): Promise<WalletAccount> {
    const adapter = this.adapters.get(type);
    if (!adapter) {
      throw new Error(`Unknown wallet type: ${type}`);
    }

    if (!adapter.isAvailable()) {
      throw new Error(`${adapter.name} wallet is not available`);
    }

    await adapter.connect(chainInfo);
    this.currentAdapter = adapter;
    this.chainInfo = chainInfo;

    if (!adapter.account) {
      throw new Error("Failed to get account after connection");
    }

    return adapter.account;
  }

  async disconnect(): Promise<void> {
    if (this.currentAdapter) {
      await this.currentAdapter.disconnect();
      this.currentAdapter = null;
    }
  }

  isConnected(): boolean {
    return this.currentAdapter?.state === "connected";
  }

  getAddress(): string {
    if (!this.currentAdapter) {
      throw new Error("No wallet connected");
    }
    return this.currentAdapter.getAddress();
  }
}
