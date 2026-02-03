import { BaseWalletAdapter } from "./BaseWalletAdapter.ts";
import type { ChainInfo, KeplrChainInfo, KeplrSignDoc, LeapLike, SignedTx, UnsignedTx, WalletWindow } from "./types.ts";

export class LeapAdapter extends BaseWalletAdapter {
  readonly name = "Leap";
  readonly icon = "https://assets.leapwallet.io/logos/leap-cosmos-logo.svg";

  private get leap(): LeapLike | undefined {
    return typeof window !== "undefined" ? (window as WalletWindow).leap : undefined;
  }

  isAvailable(): boolean {
    return !!this.leap;
  }

  async connect(chainInfo: ChainInfo): Promise<void> {
    const leap = this.leap;
    if (!leap) {
      throw new Error("Leap wallet is not installed");
    }

    this.setState("connecting");

    try {
      if (leap.experimentalSuggestChain) {
        await leap.experimentalSuggestChain(this.toLeapChainInfo(chainInfo));
      }

      await leap.enable(chainInfo.chainId);

      const key = await leap.getKey(chainInfo.chainId);

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
    const leap = this.leap;
    if (!leap || !this._account) {
      throw new Error("Leap wallet is not connected");
    }

    const signDoc = this.createSignDoc(tx, chainId);
    const result = await leap.signDirect(chainId, this._account.address, signDoc);

    return {
      bodyBytes: this.base64ToBytes(result.signed.bodyBytes),
      authInfoBytes: this.base64ToBytes(result.signed.authInfoBytes),
      signatures: [this.base64ToBytes(result.signature.signature)],
    };
  }

  async signArbitrary(chainId: string, data: string | Uint8Array): Promise<{ signature: Uint8Array; pubKey: Uint8Array }> {
    const leap = this.leap;
    if (!leap?.signArbitrary || !this._account) {
      throw new Error("Leap signArbitrary not available");
    }

    const result = await leap.signArbitrary(chainId, this._account.address, data);

    return {
      signature: this.base64ToBytes(result.signature),
      pubKey: this.base64ToBytes(result.pub_key.value),
    };
  }

  private toLeapChainInfo(chainInfo: ChainInfo): KeplrChainInfo {
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

  private createSignDoc(_tx: UnsignedTx, chainId: string): KeplrSignDoc {
    return {
      bodyBytes: "",
      authInfoBytes: "",
      chainId,
      accountNumber: 0,
    };
  }

  private base64ToBytes(base64: string): Uint8Array {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }
}
