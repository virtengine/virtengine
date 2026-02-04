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
import { isBrowser, toKeplrChainInfo } from '../utils';

interface KeplrProvider {
  enable: (chainId: string) => Promise<void>;
  getKey: (chainId: string) => Promise<{
    name?: string;
    algo?: string;
    pubKey?: Uint8Array;
    bech32Address: string;
    isNanoLedger?: boolean;
  }>;
  signAmino: (
    chainId: string,
    signer: string,
    signDoc: AminoSignDoc
  ) => Promise<AminoSignResponse>;
  signDirect: (
    chainId: string,
    signer: string,
    signDoc: DirectSignDoc
  ) => Promise<DirectSignResponse>;
  signArbitrary?: (
    chainId: string,
    signer: string,
    data: string
  ) => Promise<ArbitrarySignResponse>;
  experimentalSuggestChain?: (chainInfo: ReturnType<typeof toKeplrChainInfo>) => Promise<void>;
  getOfflineSignerAuto?: (chainId: string) => Promise<{ getAccounts: () => Promise<WalletAccount[]> }>;
}

type KeplrWindow = typeof window & { keplr?: KeplrProvider };

function getKeplr(): KeplrProvider | null {
  if (!isBrowser()) return null;
  const keplr = (window as KeplrWindow).keplr;
  return keplr ?? null;
}

async function getAccountsFromSigner(
  keplr: KeplrProvider,
  chainId: string
): Promise<WalletAccount[]> {
  if (!keplr.getOfflineSignerAuto) return [];
  const signer = await keplr.getOfflineSignerAuto(chainId);
  return signer.getAccounts();
}

export function createKeplrAdapter(): WalletAdapter {
  return {
    id: 'keplr',
    name: 'Keplr',
    supportsExtension: true,
    supportsMobile: true,
    isInstalled: () => !!getKeplr(),
    connect: async (context: WalletAdapterContext): Promise<WalletConnection> => {
      const keplr = getKeplr();
      if (!keplr) {
        throw new Error('Keplr extension not installed');
      }

      if (keplr.experimentalSuggestChain) {
        await keplr.experimentalSuggestChain(toKeplrChainInfo(context.chain));
      }

      await keplr.enable(context.chain.chainId);

      const accounts = await getAccountsFromSigner(keplr, context.chain.chainId);
      if (accounts.length > 0) {
        return { accounts, activeAccount: accounts[0] };
      }

      const key = await keplr.getKey(context.chain.chainId);
      const account: WalletAccount = {
        address: key.bech32Address,
        algo: key.algo,
        pubkey: key.pubKey,
        isNanoLedger: key.isNanoLedger,
        name: key.name,
      };

      return { accounts: [account], activeAccount: account };
    },
    disconnect: async () => {
      return;
    },
    getAccounts: async (chainId: string) => {
      const keplr = getKeplr();
      if (!keplr) return [];
      const accounts = await getAccountsFromSigner(keplr, chainId);
      if (accounts.length > 0) return accounts;
      const key = await keplr.getKey(chainId);
      return [{
        address: key.bech32Address,
        algo: key.algo,
        pubkey: key.pubKey,
        isNanoLedger: key.isNanoLedger,
        name: key.name,
      }];
    },
    signAmino: async (chainId: string, signer: string, signDoc: AminoSignDoc) => {
      const keplr = getKeplr();
      if (!keplr) {
        throw new Error('Keplr extension not installed');
      }
      return keplr.signAmino(chainId, signer, signDoc);
    },
    signDirect: async (chainId: string, signer: string, signDoc: DirectSignDoc) => {
      const keplr = getKeplr();
      if (!keplr) {
        throw new Error('Keplr extension not installed');
      }
      return keplr.signDirect(chainId, signer, signDoc);
    },
    signArbitrary: async (chainId: string, signer: string, data: string | Uint8Array) => {
      const keplr = getKeplr();
      if (!keplr?.signArbitrary) {
        throw new Error('Keplr does not support arbitrary signing');
      }
      const payload = typeof data === 'string' ? data : new TextDecoder().decode(data);
      return keplr.signArbitrary(chainId, signer, payload);
    },
    suggestChain: async (chain) => {
      const keplr = getKeplr();
      if (!keplr?.experimentalSuggestChain) {
        return;
      }
      await keplr.experimentalSuggestChain(toKeplrChainInfo(chain));
    },
  };
}
