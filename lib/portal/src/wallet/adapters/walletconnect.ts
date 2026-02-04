import type {
  WalletAdapter,
  WalletAdapterContext,
  WalletConnection,
  WalletAccount,
  AminoSignDoc,
  AminoSignResponse,
  DirectSignDoc,
  DirectSignResponse,
  ArbitrarySignResponse,
} from '../types';
import { toBase64 } from '../utils';

type WalletConnectClient = {
  connect: (args: any) => Promise<{ uri?: string; approval: () => Promise<any> }>;
  request: (args: any) => Promise<any>;
  disconnect: (args: any) => Promise<void>;
  session?: any;
};

type WalletConnectModal = {
  openModal: (args: { uri: string; standaloneChains?: string[] }) => void;
  closeModal: () => void;
};

let wcClient: WalletConnectClient | null = null;
let wcModal: WalletConnectModal | null = null;
let wcSession: any | null = null;

async function initClient(config: WalletAdapterContext['walletConnect']): Promise<WalletConnectClient> {
  if (!config) {
    throw new Error('WalletConnect configuration is missing');
  }
  if (wcClient) return wcClient;

  const [{ default: SignClient }, { WalletConnectModal }] = await Promise.all([
    import('@walletconnect/sign-client'),
    import('@walletconnect/modal'),
  ]);

  wcClient = await SignClient.init({
    projectId: config.projectId,
    relayUrl: config.relayUrl,
    metadata: config.metadata,
  });

  wcModal = new WalletConnectModal({
    projectId: config.projectId,
  });

  return wcClient;
}

function parseAccounts(accounts: string[]): WalletAccount[] {
  return accounts.map((account) => {
    const parts = account.split(':');
    return {
      address: parts[2] ?? account,
    };
  });
}

export function createWalletConnectAdapter(): WalletAdapter {
  return {
    id: 'walletconnect',
    name: 'WalletConnect',
    supportsExtension: false,
    supportsMobile: true,
    isInstalled: () => true,
    connect: async (context: WalletAdapterContext): Promise<WalletConnection> => {
      if (!context.walletConnect?.projectId) {
        throw new Error('WalletConnect projectId is not configured');
      }

      const client = await initClient(context.walletConnect);
      const chainId = `cosmos:${context.chain.chainId}`;

      const { uri, approval } = await client.connect({
        requiredNamespaces: {
          cosmos: {
            methods: ['cosmos_signAmino', 'cosmos_signDirect', 'cosmos_getAccounts'],
            chains: [chainId],
            events: ['accountsChanged', 'chainChanged'],
          },
        },
      });

      if (uri && wcModal) {
        wcModal.openModal({ uri, standaloneChains: [chainId] });
      }

      const session = await approval();
      wcSession = session;

      if (wcModal) {
        wcModal.closeModal();
      }

      const namespace = session.namespaces?.cosmos;
      const accounts = namespace?.accounts ? parseAccounts(namespace.accounts) : [];

      return { accounts, activeAccount: accounts[0] };
    },
    disconnect: async () => {
      if (!wcClient || !wcSession) return;
      await wcClient.disconnect({
        topic: wcSession.topic,
        reason: {
          code: 6000,
          message: 'User disconnected',
        },
      });
      wcSession = null;
    },
    getAccounts: async () => {
      if (!wcSession) return [];
      const namespace = wcSession.namespaces?.cosmos;
      if (!namespace?.accounts) return [];
      return parseAccounts(namespace.accounts);
    },
    signAmino: async (chainId: string, signer: string, signDoc: AminoSignDoc) => {
      if (!wcClient || !wcSession) {
        throw new Error('WalletConnect session not established');
      }
      const wcChainId = `cosmos:${chainId}`;
      const response = await wcClient.request({
        topic: wcSession.topic,
        chainId: wcChainId,
        request: {
          method: 'cosmos_signAmino',
          params: {
            signerAddress: signer,
            signDoc,
          },
        },
      });
      return response as AminoSignResponse;
    },
    signDirect: async (chainId: string, signer: string, signDoc: DirectSignDoc) => {
      if (!wcClient || !wcSession) {
        throw new Error('WalletConnect session not established');
      }
      const wcChainId = `cosmos:${chainId}`;
      const response = await wcClient.request({
        topic: wcSession.topic,
        chainId: wcChainId,
        request: {
          method: 'cosmos_signDirect',
          params: {
            signerAddress: signer,
            signDoc: {
              bodyBytes: toBase64(signDoc.bodyBytes),
              authInfoBytes: toBase64(signDoc.authInfoBytes),
              chainId: signDoc.chainId,
              accountNumber: signDoc.accountNumber.toString(),
            },
          },
        },
      });
      return response as DirectSignResponse;
    },
    signArbitrary: async (chainId: string, signer: string, data: string | Uint8Array) => {
      if (!wcClient || !wcSession) {
        throw new Error('WalletConnect session not established');
      }
      const wcChainId = `cosmos:${chainId}`;
      const payload = typeof data === 'string' ? data : new TextDecoder().decode(data);
      const response = await wcClient.request({
        topic: wcSession.topic,
        chainId: wcChainId,
        request: {
          method: 'cosmos_signArbitrary',
          params: {
            signerAddress: signer,
            data: payload,
          },
        },
      });
      return response as ArbitrarySignResponse;
    },
  };
}
