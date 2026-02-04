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
import { fromBase64, isBrowser, toBase64, toKeplrChainInfo } from '../utils';

interface KeplrCompatProvider {
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

interface CosmostationCosmosProvider {
  request: (args: { method: string; params?: Record<string, unknown> }) => Promise<any>;
}

type CosmostationWindow = typeof window & {
  cosmostation?: {
    cosmos?: CosmostationCosmosProvider;
    providers?: {
      keplr?: KeplrCompatProvider;
    };
  };
};

function getCosmostationKeplr(): KeplrCompatProvider | null {
  if (!isBrowser()) return null;
  const cosmostation = (window as CosmostationWindow).cosmostation;
  return cosmostation?.providers?.keplr ?? null;
}

function getCosmostationCosmos(): CosmostationCosmosProvider | null {
  if (!isBrowser()) return null;
  const cosmostation = (window as CosmostationWindow).cosmostation;
  return cosmostation?.cosmos ?? null;
}

async function getAccountsFromSigner(
  provider: KeplrCompatProvider,
  chainId: string
): Promise<WalletAccount[]> {
  if (!provider.getOfflineSignerAuto) return [];
  const signer = await provider.getOfflineSignerAuto(chainId);
  return signer.getAccounts();
}

async function requestCosmosAccount(
  cosmos: CosmostationCosmosProvider,
  chainName: string
): Promise<WalletAccount> {
  const response = await cosmos.request({
    method: 'cos_requestAccount',
    params: { chainName },
  });

  return {
    address: response?.address ?? response?.bech32Address ?? '',
    algo: response?.algo,
    pubkey: response?.pubKey ? fromBase64(response.pubKey) : undefined,
    isNanoLedger: response?.isNanoLedger,
    name: response?.name,
  };
}

function buildCosmostationChainParams(chain: WalletAdapterContext['chain']) {
  const feeCurrency = chain.feeCurrencies[0];
  return {
    chainId: chain.chainId,
    chainName: chain.chainName,
    addressPrefix: chain.bech32Prefix,
    baseDenom: chain.stakeCurrency.coinMinimalDenom,
    displayDenom: chain.stakeCurrency.coinDenom,
    decimals: chain.stakeCurrency.coinDecimals,
    restURL: chain.restEndpoint,
    rpcURL: chain.rpcEndpoint,
    coinType: chain.slip44 ?? 118,
    gasRate: feeCurrency?.gasPriceStep
      ? {
          low: feeCurrency.gasPriceStep.low,
          average: feeCurrency.gasPriceStep.average,
          high: feeCurrency.gasPriceStep.high,
        }
      : undefined,
  };
}

export function createCosmostationAdapter(): WalletAdapter {
  return {
    id: 'cosmostation',
    name: 'Cosmostation',
    supportsExtension: true,
    supportsMobile: true,
    isInstalled: () => !!getCosmostationKeplr() || !!getCosmostationCosmos(),
    connect: async (context: WalletAdapterContext): Promise<WalletConnection> => {
      const keplrCompat = getCosmostationKeplr();
      if (keplrCompat) {
        if (keplrCompat.experimentalSuggestChain) {
          await keplrCompat.experimentalSuggestChain(toKeplrChainInfo(context.chain));
        }
        await keplrCompat.enable(context.chain.chainId);
        const accounts = await getAccountsFromSigner(keplrCompat, context.chain.chainId);
        if (accounts.length > 0) {
          return { accounts, activeAccount: accounts[0] };
        }
        const key = await keplrCompat.getKey(context.chain.chainId);
        const account: WalletAccount = {
          address: key.bech32Address,
          algo: key.algo,
          pubkey: key.pubKey,
          isNanoLedger: key.isNanoLedger,
          name: key.name,
        };
        return { accounts: [account], activeAccount: account };
      }

      const cosmos = getCosmostationCosmos();
      if (!cosmos) {
        throw new Error('Cosmostation extension not installed');
      }

      await cosmos.request({
        method: 'cos_addChain',
        params: buildCosmostationChainParams(context.chain),
      }).catch(() => undefined);

      const account = await requestCosmosAccount(cosmos, context.chain.chainId);
      if (!account.address) {
        throw new Error('Failed to retrieve account from Cosmostation');
      }
      return { accounts: [account], activeAccount: account };
    },
    disconnect: async () => {
      return;
    },
    getAccounts: async (chainId: string) => {
      const keplrCompat = getCosmostationKeplr();
      if (keplrCompat) {
        const accounts = await getAccountsFromSigner(keplrCompat, chainId);
        if (accounts.length > 0) return accounts;
        const key = await keplrCompat.getKey(chainId);
        return [{
          address: key.bech32Address,
          algo: key.algo,
          pubkey: key.pubKey,
          isNanoLedger: key.isNanoLedger,
          name: key.name,
        }];
      }

      const cosmos = getCosmostationCosmos();
      if (!cosmos) return [];
      const account = await requestCosmosAccount(cosmos, chainId);
      return account.address ? [account] : [];
    },
    signAmino: async (chainId: string, signer: string, signDoc: AminoSignDoc) => {
      const keplrCompat = getCosmostationKeplr();
      if (keplrCompat) {
        return keplrCompat.signAmino(chainId, signer, signDoc);
      }

      const cosmos = getCosmostationCosmos();
      if (!cosmos) {
        throw new Error('Cosmostation extension not installed');
      }

      return cosmos.request({
        method: 'cos_signAmino',
        params: {
          chainName: chainId,
          doc: signDoc,
          address: signer,
        },
      });
    },
    signDirect: async (chainId: string, signer: string, signDoc: DirectSignDoc) => {
      const keplrCompat = getCosmostationKeplr();
      if (keplrCompat) {
        return keplrCompat.signDirect(chainId, signer, signDoc);
      }

      const cosmos = getCosmostationCosmos();
      if (!cosmos) {
        throw new Error('Cosmostation extension not installed');
      }

      return cosmos.request({
        method: 'cos_signDirect',
        params: {
          chainName: chainId,
          doc: {
            bodyBytes: toBase64(signDoc.bodyBytes),
            authInfoBytes: toBase64(signDoc.authInfoBytes),
            chainId: signDoc.chainId,
            accountNumber: signDoc.accountNumber.toString(),
          },
          address: signer,
        },
      });
    },
    signArbitrary: async (chainId: string, signer: string, data: string | Uint8Array) => {
      const keplrCompat = getCosmostationKeplr();
      if (keplrCompat?.signArbitrary) {
        const payload = typeof data === 'string' ? data : new TextDecoder().decode(data);
        return keplrCompat.signArbitrary(chainId, signer, payload);
      }

      const cosmos = getCosmostationCosmos();
      if (!cosmos) {
        throw new Error('Cosmostation extension not installed');
      }

      const payload = typeof data === 'string' ? data : new TextDecoder().decode(data);
      return cosmos.request({
        method: 'cos_signArbitrary',
        params: {
          chainName: chainId,
          data: payload,
          address: signer,
        },
      });
    },
    suggestChain: async (chain) => {
      const keplrCompat = getCosmostationKeplr();
      if (keplrCompat?.experimentalSuggestChain) {
        await keplrCompat.experimentalSuggestChain(toKeplrChainInfo(chain));
        return;
      }

      const cosmos = getCosmostationCosmos();
      if (!cosmos) return;
      await cosmos.request({
        method: 'cos_addChain',
        params: buildCosmostationChainParams(chain),
      });
    },
  };
}
