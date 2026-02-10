import SignClient from '@walletconnect/sign-client';
import { WalletConnectModal } from '@walletconnect/modal';
import type { AminoSignDoc, AminoSignResponse, DirectSignDoc, DirectSignResponse, WalletAccount, WalletChainInfo } from '../types';
import { BaseWalletAdapter } from './base';

const COSMOS_METHODS = ['cosmos_getAccounts', 'cosmos_signAmino', 'cosmos_signDirect'];
const COSMOS_EVENTS = ['accountsChanged', 'chainChanged'];

export class WalletConnectAdapter extends BaseWalletAdapter {
  readonly type = 'walletconnect' as const;
  readonly name = 'WalletConnect';
  readonly icon = 'https://walletconnect.com/favicon.ico';

  private client: SignClient | null = null;
  private modal: WalletConnectModal | null = null;
  private sessionTopic: string | null = null;
  private projectId: string;
  private metadata: { name: string; description: string; url: string; icons: string[] };

  constructor(projectId: string, metadata: { name: string; description: string; url: string; icons: string[] }) {
    super();
    this.projectId = projectId;
    this.metadata = metadata;
  }

  isAvailable(): boolean {
    return !!this.projectId;
  }

  async connect(chainInfo: WalletChainInfo): Promise<WalletAccount[]> {
    if (!this.projectId) {
      throw new Error('WalletConnect project ID is not configured');
    }

    await this.ensureClient(chainInfo);
    if (!this.client || !this.modal) {
      throw new Error('WalletConnect client initialization failed');
    }

    const caipChainId = this.toCaipChainId(chainInfo.chainId);

    const existingSession = this.client.session.getAll().find((session) =>
      session.namespaces?.cosmos?.chains?.includes(caipChainId)
    );

    if (existingSession) {
      this.sessionTopic = existingSession.topic;
      const accounts = await this.getAccounts(chainInfo);
      this.setAccounts(accounts);
      return accounts;
    }

    const { uri, approval } = await this.client.connect({
      requiredNamespaces: {
        cosmos: {
          chains: [caipChainId],
          methods: COSMOS_METHODS,
          events: COSMOS_EVENTS,
        },
      },
    });

    if (uri) {
      await this.modal.openModal({ uri, chains: [caipChainId] });
    }

    const session = await approval();
    this.sessionTopic = session.topic;

    this.modal.closeModal();

    const accounts = await this.getAccounts(chainInfo);
    this.setAccounts(accounts);
    return accounts;
  }

  async disconnect(): Promise<void> {
    if (this.client && this.sessionTopic) {
      await this.client.disconnect({
        topic: this.sessionTopic,
        reason: {
          code: 6000,
          message: 'User disconnected',
        },
      });
    }

    this.sessionTopic = null;
    this.accounts = [];
  }

  async getAccounts(chainInfo: WalletChainInfo): Promise<WalletAccount[]> {
    if (!this.client || !this.sessionTopic) {
      throw new Error('WalletConnect is not connected');
    }

    const caipChainId = this.toCaipChainId(chainInfo.chainId);

    const response = await this.client.request({
      topic: this.sessionTopic,
      chainId: caipChainId,
      request: {
        method: 'cosmos_getAccounts',
        params: {},
      },
    });

    if (!Array.isArray(response)) {
      throw new Error('Invalid WalletConnect account response');
    }

    return response.map((account: any) => ({
      address: account.address,
      algo: account.algo ?? 'secp256k1',
      pubKey: this.base64ToBytes(account.pubkey ?? account.pubKey ?? ''),
    }));
  }

  async signAmino(
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc
  ): Promise<AminoSignResponse> {
    if (!this.client || !this.sessionTopic) {
      throw new Error('WalletConnect is not connected');
    }

    const response = (await this.client.request({
      topic: this.sessionTopic,
      chainId: this.toCaipChainId(chainId),
      request: {
        method: 'cosmos_signAmino',
        params: {
          signerAddress,
          signDoc,
        },
      },
    })) as any;

    if (!response?.signature) {
      throw new Error('Invalid WalletConnect signAmino response');
    }

    return response as AminoSignResponse;
  }

  async signDirect(
    chainId: string,
    signerAddress: string,
    signDoc: DirectSignDoc
  ): Promise<DirectSignResponse> {
    if (!this.client || !this.sessionTopic) {
      throw new Error('WalletConnect is not connected');
    }

    const response = (await this.client.request({
      topic: this.sessionTopic,
      chainId: this.toCaipChainId(chainId),
      request: {
        method: 'cosmos_signDirect',
        params: {
          signerAddress,
          signDoc: {
            bodyBytes: this.bytesToBase64(signDoc.bodyBytes),
            authInfoBytes: this.bytesToBase64(signDoc.authInfoBytes),
            chainId: signDoc.chainId,
            accountNumber: signDoc.accountNumber,
          },
        },
      },
    })) as any;

    if (!response?.signature) {
      throw new Error('Invalid WalletConnect signDirect response');
    }

    return {
      signed: signDoc,
      signature: response.signature,
    } as DirectSignResponse;
  }

  private async ensureClient(chainInfo: WalletChainInfo): Promise<void> {
    if (this.client) return;

    this.client = await SignClient.init({
      projectId: this.projectId,
      metadata: this.metadata,
    });

    this.modal = new WalletConnectModal({
      projectId: this.projectId,
      chains: [this.toCaipChainId(chainInfo.chainId)],
    });
  }

  private toCaipChainId(chainId: string): string {
    return `cosmos:${chainId}`;
  }
}
