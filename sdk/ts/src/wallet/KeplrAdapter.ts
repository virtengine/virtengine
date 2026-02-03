import { BaseWalletAdapter } from "./BaseWalletAdapter.ts";
import type { ChainInfo, KeplrChainInfo, KeplrLike, KeplrSignDoc, SignedTx, UnsignedTx, WalletWindow } from "./types.ts";

export class KeplrAdapter extends BaseWalletAdapter {
  readonly name = "Keplr";
  readonly icon = "https://wallet.keplr.app/keplr-logo.svg";

  private get keplr(): KeplrLike | undefined {
    return typeof window !== "undefined" ? (window as WalletWindow).keplr : undefined;
  }

  isAvailable(): boolean {
    return !!this.keplr;
  }

  async connect(chainInfo: ChainInfo): Promise<void> {
    const keplr = this.keplr;
    if (!keplr) {
      throw new Error("Keplr wallet is not installed");
    }

    this.setState("connecting");

    try {
      // Suggest chain if experimentalSuggestChain is available
      if (keplr.experimentalSuggestChain) {
        await keplr.experimentalSuggestChain(this.toKeplrChainInfo(chainInfo));
      }

      await keplr.enable(chainInfo.chainId);

      const key = await keplr.getKey(chainInfo.chainId);

      this.setAccount({
        address: key.bech32Address,
        pubKey: key.pubKey,
        algo: key.algo,
      });

      this._chainId = chainInfo.chainId;
      this.setState("connected");
    } catch (error) {
      this.setState("error");
      throw error;
    }
  }

  async signTx(tx: UnsignedTx, chainId: string): Promise<SignedTx> {
    const keplr = this.keplr;
    if (!keplr || !this._account) {
      throw new Error("Keplr wallet is not connected");
    }

    const signDoc = this.createSignDoc(tx, chainId);
    const result = await keplr.signDirect(chainId, this._account.address, signDoc);

    return {
      bodyBytes: this.base64ToBytes(result.signed.bodyBytes),
      authInfoBytes: this.base64ToBytes(result.signed.authInfoBytes),
      signatures: [this.base64ToBytes(result.signature.signature)],
    };
  }

  async signArbitrary(chainId: string, data: string | Uint8Array): Promise<{ signature: Uint8Array; pubKey: Uint8Array }> {
    const keplr = this.keplr;
    if (!keplr?.signArbitrary || !this._account) {
      throw new Error("Keplr signArbitrary not available");
    }

    const result = await keplr.signArbitrary(chainId, this._account.address, data);

    return {
      signature: this.base64ToBytes(result.signature),
      pubKey: this.base64ToBytes(result.pub_key.value),
    };
  }

  private toKeplrChainInfo(chainInfo: ChainInfo): KeplrChainInfo {
    return {
      chainId: chainInfo.chainId,
      chainName: chainInfo.chainName,
      rpc: chainInfo.rpc,
      rest: chainInfo.rest,
      bip44: chainInfo.bip44,
      bech32Config: chainInfo.bech32Config,
      currencies: chainInfo.currencies,
      feeCurrencies: chainInfo.feeCurrencies,
      stakeCurrency: chainInfo.stakeCurrency,
    };
  }

  private createSignDoc(tx: UnsignedTx, chainId: string): KeplrSignDoc {
    return {
      bodyBytes: this.bytesToBase64(this.encodeBody(tx)),
      authInfoBytes: this.bytesToBase64(this.encodeAuthInfo(tx)),
      chainId,
      accountNumber: 0, // Would need to fetch from chain
    };
  }

  private encodeBody(_tx: UnsignedTx): Uint8Array {
    // Simplified - actual implementation would use proper proto encoding
    return new Uint8Array();
  }

  private encodeAuthInfo(_tx: UnsignedTx): Uint8Array {
    // Simplified - actual implementation would use proper proto encoding
    return new Uint8Array();
  }

  private base64ToBytes(base64: string): Uint8Array {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }

  private bytesToBase64(bytes: Uint8Array): string {
    let binary = "";
    for (let i = 0; i < bytes.length; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }
}
