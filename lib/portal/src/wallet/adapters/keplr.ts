import type {
  AminoSignDoc,
  AminoSignResponse,
  DirectSignDoc,
  DirectSignResponse,
  WalletAccount,
  WalletChainInfo,
  WalletSignOptions,
} from "../types";
import { BaseWalletAdapter } from "./base";

interface KeplrLike {
  enable(chainId: string | string[]): Promise<void>;
  getKey(
    chainId: string,
  ): Promise<{ bech32Address: string; pubKey: Uint8Array; algo: string }>;
  signAmino(
    chainId: string,
    signer: string,
    signDoc: AminoSignDoc,
    signOptions?: WalletSignOptions,
  ): Promise<AminoSignResponse>;
  signDirect(
    chainId: string,
    signer: string,
    signDoc: DirectSignDoc,
    signOptions?: WalletSignOptions,
  ): Promise<DirectSignResponse>;
  signArbitrary?(
    chainId: string,
    signer: string,
    data: string | Uint8Array,
  ): Promise<{ signature: string; pub_key: { value: string } }>;
  experimentalSuggestChain?(chainInfo: Record<string, unknown>): Promise<void>;
}

interface OfflineSigner {
  getAccounts(): Promise<
    Array<{ address: string; algo: string; pubkey: Uint8Array }>
  >;
}

declare global {
  interface Window {
    keplr?: KeplrLike;
    getOfflineSigner?: (
      chainId: string,
      signOptions?: WalletSignOptions,
    ) => OfflineSigner;
    getOfflineSignerAuto?: (
      chainId: string,
      signOptions?: WalletSignOptions,
    ) => Promise<OfflineSigner>;
  }
}

export class KeplrAdapter extends BaseWalletAdapter {
  readonly type = "keplr" as const;
  readonly name = "Keplr";
  readonly icon = "https://wallet.keplr.app/keplr-logo.svg";

  private get keplr(): KeplrLike | undefined {
    if (typeof window === "undefined") return undefined;
    return window.keplr;
  }

  isAvailable(): boolean {
    return !!this.keplr;
  }

  async connect(chainInfo: WalletChainInfo): Promise<WalletAccount[]> {
    const keplr = this.keplr;
    if (!keplr) {
      throw new Error("Keplr wallet is not installed");
    }

    if (keplr.experimentalSuggestChain) {
      await keplr.experimentalSuggestChain(this.toKeplrChainInfo(chainInfo));
    }

    await keplr.enable(chainInfo.chainId);

    const accounts = await this.getAccounts(chainInfo);
    if (accounts.length === 0) {
      throw new Error("No Keplr accounts available");
    }

    this.setAccounts(accounts);
    return accounts;
  }

  async getAccounts(chainInfo: WalletChainInfo): Promise<WalletAccount[]> {
    const keplr = this.keplr;
    if (!keplr) {
      throw new Error("Keplr wallet is not installed");
    }

    const signer = window.getOfflineSignerAuto
      ? await window.getOfflineSignerAuto(chainInfo.chainId)
      : window.getOfflineSigner?.(chainInfo.chainId);

    if (signer) {
      const accounts = await signer.getAccounts();
      return accounts.map((account) => ({
        address: account.address,
        pubKey: account.pubkey,
        algo: account.algo,
      }));
    }

    const key = await keplr.getKey(chainInfo.chainId);
    return [
      {
        address: key.bech32Address,
        pubKey: key.pubKey,
        algo: key.algo,
      },
    ];
  }

  async signAmino(
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc,
    signOptions?: WalletSignOptions,
  ): Promise<AminoSignResponse> {
    const keplr = this.keplr;
    if (!keplr) {
      throw new Error("Keplr wallet is not connected");
    }

    return keplr.signAmino(chainId, signerAddress, signDoc, signOptions);
  }

  async signDirect(
    chainId: string,
    signerAddress: string,
    signDoc: DirectSignDoc,
  ): Promise<DirectSignResponse> {
    const keplr = this.keplr;
    if (!keplr) {
      throw new Error("Keplr wallet is not connected");
    }

    return keplr.signDirect(chainId, signerAddress, signDoc);
  }

  async signArbitrary(
    chainId: string,
    signerAddress: string,
    data: string | Uint8Array,
  ): Promise<{ signature: string; pubKey: Uint8Array }> {
    const keplr = this.keplr;
    if (!keplr?.signArbitrary) {
      throw new Error("Keplr signArbitrary not available");
    }

    const result = await keplr.signArbitrary(chainId, signerAddress, data);
    return {
      signature: result.signature,
      pubKey: this.base64ToBytes(result.pub_key.value),
    };
  }

  private toKeplrChainInfo(
    chainInfo: WalletChainInfo,
  ): Record<string, unknown> {
    return {
      chainId: chainInfo.chainId,
      chainName: chainInfo.chainName,
      rpc: chainInfo.rpcEndpoint,
      rest: chainInfo.restEndpoint,
      bip44: { coinType: chainInfo.bip44?.coinType ?? 118 },
      bech32Config: chainInfo.bech32Config,
      currencies: chainInfo.currencies,
      feeCurrencies: chainInfo.feeCurrencies,
      stakeCurrency: chainInfo.stakeCurrency,
      features: chainInfo.features ?? [],
    };
  }
}
