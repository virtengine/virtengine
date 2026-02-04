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

interface LeapProvider {
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

type LeapWindow = typeof window & { leap?: LeapProvider };

function getLeap(): LeapProvider | null {
  if (!isBrowser()) return null;
  const leap = (window as LeapWindow).leap;
  return leap ?? null;
}

async function getAccountsFromSigner(
  leap: LeapProvider,
  chainId: string
): Promise<WalletAccount[]> {
  if (!leap.getOfflineSignerAuto) return [];
  const signer = await leap.getOfflineSignerAuto(chainId);
  return signer.getAccounts();
}

export function createLeapAdapter(): WalletAdapter {
  return {
    id: 'leap',
    name: 'Leap',
    supportsExtension: true,
    supportsMobile: true,
    isInstalled: () => !!getLeap(),
    connect: async (context: WalletAdapterContext): Promise<WalletConnection> => {
      const leap = getLeap();
      if (!leap) {
        throw new Error('Leap extension not installed');
      }

      if (leap.experimentalSuggestChain) {
        await leap.experimentalSuggestChain(toKeplrChainInfo(context.chain));
      }

      await leap.enable(context.chain.chainId);

      const accounts = await getAccountsFromSigner(leap, context.chain.chainId);
      if (accounts.length > 0) {
        return { accounts, activeAccount: accounts[0] };
      }

      const key = await leap.getKey(context.chain.chainId);
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
      const leap = getLeap();
      if (!leap) return [];
      const accounts = await getAccountsFromSigner(leap, chainId);
      if (accounts.length > 0) return accounts;
      const key = await leap.getKey(chainId);
      return [{
        address: key.bech32Address,
        algo: key.algo,
        pubkey: key.pubKey,
        isNanoLedger: key.isNanoLedger,
        name: key.name,
      }];
    },
    signAmino: async (chainId: string, signer: string, signDoc: AminoSignDoc) => {
      const leap = getLeap();
      if (!leap) {
        throw new Error('Leap extension not installed');
      }
      return leap.signAmino(chainId, signer, signDoc);
    },
    signDirect: async (chainId: string, signer: string, signDoc: DirectSignDoc) => {
      const leap = getLeap();
      if (!leap) {
        throw new Error('Leap extension not installed');
      }
      return leap.signDirect(chainId, signer, signDoc);
    },
    signArbitrary: async (chainId: string, signer: string, data: string | Uint8Array) => {
      const leap = getLeap();
      if (!leap?.signArbitrary) {
        throw new Error('Leap does not support arbitrary signing');
      }
      const payload = typeof data === 'string' ? data : new TextDecoder().decode(data);
      return leap.signArbitrary(chainId, signer, payload);
    },
    suggestChain: async (chain) => {
      const leap = getLeap();
      if (!leap?.experimentalSuggestChain) {
        return;
      }
      await leap.experimentalSuggestChain(toKeplrChainInfo(chain));
    },
  };
}
