import type { AminoSignDoc, AminoSignResponse, DirectSignDoc, DirectSignResponse, WalletAccount, WalletChainInfo } from '../types';
import { BaseWalletAdapter } from './base';

interface CosmostationLike {
  cosmos?: {
    request(params: { method: string; params?: Record<string, unknown> }): Promise<any>;
  };
}

declare global {
  interface Window {
    cosmostation?: CosmostationLike;
  }
}

export class CosmostationAdapter extends BaseWalletAdapter {
  readonly type = 'cosmostation' as const;
  readonly name = 'Cosmostation';
  readonly icon = 'https://wallet.cosmostation.io/favicon.ico';

  private get cosmostation(): CosmostationLike | undefined {
    if (typeof window === 'undefined') return undefined;
    return window.cosmostation;
  }

  isAvailable(): boolean {
    return !!this.cosmostation?.cosmos;
  }

  async connect(chainInfo: WalletChainInfo): Promise<WalletAccount[]> {
    const cosmostation = this.cosmostation;
    if (!cosmostation?.cosmos) {
      throw new Error('Cosmostation wallet is not installed');
    }

    await cosmostation.cosmos.request({
      method: 'cos_addChain',
      params: this.toCosmostationChainInfo(chainInfo),
    });

    const account = await cosmostation.cosmos.request({
      method: 'cos_requestAccount',
      params: { chainName: chainInfo.chainId },
    });

    const accounts = this.toAccounts(account);
    if (accounts.length === 0) {
      throw new Error('No Cosmostation accounts available');
    }

    this.setAccounts(accounts);
    return accounts;
  }

  async getAccounts(chainInfo: WalletChainInfo): Promise<WalletAccount[]> {
    const cosmostation = this.cosmostation;
    if (!cosmostation?.cosmos) {
      throw new Error('Cosmostation wallet is not installed');
    }

    const account = await cosmostation.cosmos.request({
      method: 'cos_requestAccount',
      params: { chainName: chainInfo.chainId },
    });

    return this.toAccounts(account);
  }

  async signAmino(
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc
  ): Promise<AminoSignResponse> {
    const cosmostation = this.cosmostation;
    if (!cosmostation?.cosmos) {
      throw new Error('Cosmostation wallet is not connected');
    }

    const response = await cosmostation.cosmos.request({
      method: 'cos_signAmino',
      params: {
        chainName: chainId,
        doc: signDoc,
        signer: signerAddress,
      },
    });

    if (!response?.signed_doc || !response?.signature) {
      throw new Error('Invalid Cosmostation signAmino response');
    }

    return {
      signed: response.signed_doc,
      signature: response.signature,
    } as AminoSignResponse;
  }

  async signDirect(
    chainId: string,
    signerAddress: string,
    signDoc: DirectSignDoc
  ): Promise<DirectSignResponse> {
    const cosmostation = this.cosmostation;
    if (!cosmostation?.cosmos) {
      throw new Error('Cosmostation wallet is not connected');
    }

    const response = await cosmostation.cosmos.request({
      method: 'cos_signDirect',
      params: {
        chainName: chainId,
        signer: signerAddress,
        doc: {
          body_bytes: this.bytesToBase64(signDoc.bodyBytes),
          auth_info_bytes: this.bytesToBase64(signDoc.authInfoBytes),
          chain_id: signDoc.chainId,
          account_number: String(signDoc.accountNumber),
        },
      },
    });

    if (!response?.signed_doc || !response?.signature) {
      throw new Error('Invalid Cosmostation signDirect response');
    }

    return {
      signed: {
        bodyBytes: this.base64ToBytes(response.signed_doc.body_bytes),
        authInfoBytes: this.base64ToBytes(response.signed_doc.auth_info_bytes),
        chainId: signDoc.chainId,
        accountNumber: signDoc.accountNumber,
      },
      signature: {
        signature: response.signature,
      },
    };
  }

  private toAccounts(account: any): WalletAccount[] {
    if (!account?.address || !account?.publicKey) {
      return [];
    }

    return [
      {
        address: account.address as string,
        pubKey: this.hexToBytes(account.publicKey as string),
        algo: 'secp256k1',
      },
    ];
  }

  private toCosmostationChainInfo(chainInfo: WalletChainInfo): Record<string, unknown> {
    return {
      chainId: chainInfo.chainId,
      chainName: chainInfo.chainName,
      addressPrefix: chainInfo.bech32Config.bech32PrefixAccAddr,
      baseDenom: chainInfo.stakeCurrency.coinMinimalDenom,
      displayDenom: chainInfo.stakeCurrency.coinDenom,
      restURL: chainInfo.restEndpoint,
      coinType: String(chainInfo.bip44?.coinType ?? 118),
      decimals: chainInfo.stakeCurrency.coinDecimals,
    };
  }

  private hexToBytes(hex: string): Uint8Array {
    const normalized = hex.startsWith('0x') ? hex.slice(2) : hex;
    const bytes = new Uint8Array(normalized.length / 2);
    for (let i = 0; i < normalized.length; i += 2) {
      bytes[i / 2] = parseInt(normalized.substring(i, i + 2), 16);
    }
    return bytes;
  }
}
