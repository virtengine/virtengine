import type { WalletAccount, WalletAdapter, WalletChainInfo, AminoSignDoc, DirectSignDoc, WalletSignOptions, WalletType, AminoSignResponse, DirectSignResponse } from '../types';

export abstract class BaseWalletAdapter implements WalletAdapter {
  abstract readonly type: WalletType;
  abstract readonly name: string;
  abstract readonly icon?: string;

  protected accounts: WalletAccount[] = [];

  abstract isAvailable(): boolean;
  abstract connect(chainInfo: WalletChainInfo): Promise<WalletAccount[]>;
  abstract getAccounts(chainInfo: WalletChainInfo): Promise<WalletAccount[]>;
  abstract signAmino(
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc,
    signOptions?: WalletSignOptions
  ): Promise<AminoSignResponse>;
  abstract signDirect(
    chainId: string,
    signerAddress: string,
    signDoc: DirectSignDoc
  ): Promise<DirectSignResponse>;

  async disconnect(): Promise<void> {
    this.accounts = [];
  }

  protected setAccounts(accounts: WalletAccount[]): void {
    this.accounts = accounts;
  }

  protected base64ToBytes(base64: string): Uint8Array {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }

  protected bytesToBase64(bytes: Uint8Array): string {
    let binary = '';
    for (let i = 0; i < bytes.length; i += 1) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }
}
