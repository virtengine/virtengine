import { BaseWalletAdapter } from "./BaseWalletAdapter.ts";
import type {
  ChainInfo,
  CosmostationLike,
  CosmostationRequestParams,
  CosmostationSignDoc,
  SignedTx,
  UnsignedTx,
  WalletWindow } from "./types.ts";

export class CosmostationAdapter extends BaseWalletAdapter {
  readonly name = "Cosmostation";
  readonly icon = "https://wallet.cosmostation.io/favicon.ico";

  private get cosmostation(): CosmostationLike | undefined {
    return typeof window !== "undefined" ? (window as WalletWindow).cosmostation : undefined;
  }

  isAvailable(): boolean {
    return !!this.cosmostation?.cosmos;
  }

  async connect(chainInfo: ChainInfo): Promise<void> {
    const cosmostation = this.cosmostation;
    if (!cosmostation?.cosmos) {
      throw new Error("Cosmostation wallet is not installed");
    }

    this.setState("connecting");

    try {
      // Add chain if needed
      await cosmostation.cosmos.request({
        method: "cos_addChain",
        params: this.toCosmostationChainInfo(chainInfo),
      });

      // Request account
      const account = await cosmostation.cosmos.request({
        method: "cos_requestAccount",
        params: { chainName: chainInfo.chainId },
      });

      if (!account.address || !account.publicKey) {
        throw new Error("Failed to get account from Cosmostation");
      }

      this.setAccount({
        address: account.address,
        pubKey: this.hexToBytes(account.publicKey),
        algo: "secp256k1",
      });

      this._chainId = chainInfo.chainId;
      this.setState("connected");
    } catch (error) {
      this.setState("error");
      throw error;
    }
  }

  async signTx(tx: UnsignedTx, chainId: string): Promise<SignedTx> {
    const cosmostation = this.cosmostation;
    if (!cosmostation?.cosmos || !this._account) {
      throw new Error("Cosmostation wallet is not connected");
    }

    const result = await cosmostation.cosmos.request({
      method: "cos_signDirect",
      params: {
        chainName: chainId,
        doc: this.createSignDoc(tx, chainId),
      },
    });

    if (!result.signed_doc || !result.signature) {
      throw new Error("Invalid sign response from Cosmostation");
    }

    return {
      bodyBytes: this.base64ToBytes(result.signed_doc.body_bytes),
      authInfoBytes: this.base64ToBytes(result.signed_doc.auth_info_bytes),
      signatures: [this.base64ToBytes(result.signature)],
    };
  }

  private toCosmostationChainInfo(chainInfo: ChainInfo): CosmostationRequestParams {
    return {
      chainId: chainInfo.chainId,
      chainName: chainInfo.chainName,
      addressPrefix: chainInfo.bech32Config.bech32PrefixAccAddr,
      baseDenom: chainInfo.stakeCurrency.coinMinimalDenom,
      displayDenom: chainInfo.stakeCurrency.coinDenom,
      restURL: chainInfo.rest,
      coinType: String(chainInfo.bip44.coinType),
      decimals: chainInfo.stakeCurrency.coinDecimals,
    };
  }

  private createSignDoc(_tx: UnsignedTx, chainId: string): CosmostationSignDoc {
    return {
      body_bytes: "",
      auth_info_bytes: "",
      chain_id: chainId,
      account_number: "0",
    };
  }

  private hexToBytes(hex: string): Uint8Array {
    const bytes = new Uint8Array(hex.length / 2);
    for (let i = 0; i < hex.length; i += 2) {
      bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
    }
    return bytes;
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
