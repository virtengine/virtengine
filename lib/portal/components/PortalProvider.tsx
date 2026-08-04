// @ts-nocheck
/**
 * Portal Provider
 * VE-700: Main portal context provider that combines all sub-providers
 */
import * as React from "react";
import { AuthProvider } from "../hooks/useAuth";
import { IdentityProvider } from "../hooks/useIdentity";
import { MFAProvider } from "../hooks/useMFA";
import { MarketplaceProvider } from "../hooks/useMarketplace";
import { useChain } from "../hooks/useChain";
import { ProviderProvider } from "../hooks/useProvider";
import { HPCProvider } from "../hooks/useHPC";
import { ChainProvider } from "../hooks/useChain";
import { WalletProvider, useWallet } from "../src/wallet/context";
import type {
  CheckoutMutationAdapter,
  CheckoutMutationProjector,
} from "./marketplace/checkout-mutation";
import type { HPCSignerAdapter } from "./hpc/hpc-mutation";
import type { HPCOutputAdapter } from "./hpc/hpc-output";
import type { HPCQueryAdapter } from "./hpc/hpc-query";
import type { ProviderDomainVerifier } from "./provider/domain-verification";
import type { ProviderOfferingMutationAdapter } from "./provider/offering-mutation";
import type {
  WalletProviderConfig,
  WalletChainInfo,
} from "../src/wallet/types";
import type { PortalConfig, PortalProviderProps } from "../types/config";

function ProductProviders({
  children,
  mutationAdapter,
  resultProjector,
  mutationTimeoutMs,
  queryChainId,
  hpcMutationAdapter,
  hpcOutputAdapter,
  hpcQueryAdapter,
  providerDomainVerifier,
  providerOfferingMutationAdapter,
}: {
  children: React.ReactNode;
  mutationAdapter?: CheckoutMutationAdapter;
  resultProjector?: CheckoutMutationProjector;
  mutationTimeoutMs?: number;
  queryChainId: string;
  hpcMutationAdapter?: HPCSignerAdapter;
  hpcOutputAdapter?: HPCOutputAdapter;
  hpcQueryAdapter?: HPCQueryAdapter;
  providerDomainVerifier?: ProviderDomainVerifier;
  providerOfferingMutationAdapter?: ProviderOfferingMutationAdapter;
}) {
  const chain = useChain();
  const wallet = useWallet();
  const walletAddress =
    wallet.accounts[wallet.activeAccountIndex]?.address ?? null;
  const accountAddress = wallet.chainId === queryChainId ? walletAddress : null;
  const mutationContext = React.useMemo(
    () =>
      accountAddress && wallet.chainId === queryChainId
        ? { chainId: wallet.chainId, customerAddress: accountAddress }
        : undefined,
    [accountAddress, queryChainId, wallet.chainId],
  );

  return (
    <MarketplaceProvider
      queryClient={chain.queryClient}
      accountAddress={accountAddress}
      mutationAdapter={mutationAdapter}
      mutationContext={mutationContext}
      resultProjector={resultProjector}
      mutationTimeoutMs={mutationTimeoutMs}
    >
      <ProviderProvider
        queryClient={chain.queryClient}
        chainId={queryChainId}
        accountAddress={accountAddress}
        domainVerifier={
          accountAddress &&
          providerDomainVerifier?.chainId === queryChainId &&
          providerDomainVerifier.accountAddress === accountAddress
            ? providerDomainVerifier
            : undefined
        }
        offeringMutationAdapter={
          accountAddress &&
          providerOfferingMutationAdapter?.chainId === queryChainId &&
          providerOfferingMutationAdapter.accountAddress === accountAddress
            ? providerOfferingMutationAdapter
            : undefined
        }
      >
        <HPCProvider
          queryClient={chain.queryClient}
          chainId={queryChainId}
          accountAddress={accountAddress}
          mutationAdapter={
            accountAddress &&
            hpcMutationAdapter?.chainId === queryChainId &&
            hpcMutationAdapter.accountAddress === accountAddress
              ? hpcMutationAdapter
              : undefined
          }
          outputAdapter={
            accountAddress &&
            hpcOutputAdapter?.chainId === queryChainId &&
            hpcOutputAdapter.accountAddress === accountAddress
              ? hpcOutputAdapter
              : undefined
          }
          queryAdapter={
            accountAddress &&
            hpcQueryAdapter?.chainId === queryChainId &&
            hpcQueryAdapter.accountAddress === accountAddress
              ? hpcQueryAdapter
              : undefined
          }
        >
          {children}
        </HPCProvider>
      </ProviderProvider>
    </MarketplaceProvider>
  );
}

/**
 * Portal context value
 */
export interface PortalContextValue {
  /**
   * Portal configuration
   */
  config: PortalConfig;

  /**
   * Whether the portal is ready
   */
  isReady: boolean;
}

const PortalContext = React.createContext<PortalContextValue | null>(null);

/**
 * Portal provider component
 * Combines all portal providers in the correct order
 */
export function PortalProvider({
  config,
  chainConfig,
  walletConfig,
  marketplaceMutationAdapter,
  marketplaceResultProjector,
  marketplaceMutationTimeoutMs,
  hpcMutationAdapter,
  hpcOutputAdapter,
  hpcQueryAdapter,
  providerDomainVerifier,
  providerOfferingMutationAdapter,
  children,
}: PortalProviderProps): JSX.Element {
  const [isReady, setIsReady] = React.useState(false);

  // Initialize portal
  React.useEffect(() => {
    const init = async () => {
      // Perform any initialization
      setIsReady(true);
    };

    init();
  }, []);

  const value: PortalContextValue = {
    config,
    isReady,
  };

  const defaultChainInfo: WalletChainInfo = {
    chainId: chainConfig.chainId ?? config.chainId ?? "virtengine-1",
    chainName: config.networkName ?? "VirtEngine",
    rpcEndpoint: config.chainEndpoint,
    restEndpoint: config.chainRestEndpoint ?? chainConfig.restEndpoint,
    bech32Config: {
      bech32PrefixAccAddr: "virtengine",
      bech32PrefixAccPub: "virtenginepub",
      bech32PrefixValAddr: "virtenginevaloper",
      bech32PrefixValPub: "virtenginevaloperpub",
      bech32PrefixConsAddr: "virtenginevalcons",
      bech32PrefixConsPub: "virtenginevalconspub",
    },
    bip44: { coinType: 118 },
    stakeCurrency: {
      coinDenom: "VE",
      coinMinimalDenom: "uve",
      coinDecimals: 6,
    },
    currencies: [
      {
        coinDenom: "VE",
        coinMinimalDenom: "uve",
        coinDecimals: 6,
      },
    ],
    feeCurrencies: [
      {
        coinDenom: "VE",
        coinMinimalDenom: "uve",
        coinDecimals: 6,
        gasPriceStep: { low: 0.01, average: 0.025, high: 0.04 },
      },
    ],
    features: ["cosmwasm", "ibc-transfer", "ibc-go"],
  };

  const resolvedWalletConfig: WalletProviderConfig = walletConfig ?? {
    chainInfo: defaultChainInfo,
    autoConnect: true,
  };

  // Providers are nested in dependency order:
  // Chain (base) -> Auth -> Identity -> MFA -> Marketplace, Provider, HPC
  return (
    <PortalContext.Provider value={value}>
      <ChainProvider config={chainConfig}>
        <WalletProvider config={resolvedWalletConfig}>
          <AuthProvider config={config}>
            <IdentityProvider>
              <MFAProvider>
                <ProductProviders
                  mutationAdapter={marketplaceMutationAdapter}
                  resultProjector={marketplaceResultProjector}
                  mutationTimeoutMs={marketplaceMutationTimeoutMs}
                  queryChainId={chainConfig.chainId}
                  hpcMutationAdapter={hpcMutationAdapter}
                  hpcOutputAdapter={hpcOutputAdapter}
                  hpcQueryAdapter={hpcQueryAdapter}
                  providerDomainVerifier={providerDomainVerifier}
                  providerOfferingMutationAdapter={
                    providerOfferingMutationAdapter
                  }
                >
                  {children}
                </ProductProviders>
              </MFAProvider>
            </IdentityProvider>
          </AuthProvider>
        </WalletProvider>
      </ChainProvider>
    </PortalContext.Provider>
  );
}

/**
 * Hook to access portal context
 */
export function usePortal(): PortalContextValue {
  const context = React.useContext(PortalContext);
  if (!context) {
    throw new Error("usePortal must be used within a PortalProvider");
  }
  return context;
}
