/**
 * Portal Provider
 * VE-700: Main portal context provider that combines all sub-providers
 */
import * as React from 'react';
import { AuthProvider } from '../hooks/useAuth';
import { IdentityProvider } from '../hooks/useIdentity';
import { MFAProvider } from '../hooks/useMFA';
import { MarketplaceProvider } from '../hooks/useMarketplace';
import { ProviderProvider } from '../hooks/useProvider';
import { HPCProvider } from '../hooks/useHPC';
import { ChainProvider } from '../hooks/useChain';
import type { PortalConfig } from '../types/config';
import type { ChainConfig } from '../types/chain';
import type { WalletConfig } from '../types/wallet';

/**
 * Portal provider props
 */
export interface PortalProviderProps {
  /**
   * Portal configuration
   */
  config: PortalConfig;

  /**
   * Chain configuration
   */
  chainConfig: ChainConfig;

  /**
   * Wallet configuration
   */
  walletConfig?: Partial<WalletConfig>;

  /**
   * Children
   */
  children: React.ReactNode;
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

  // Providers are nested in dependency order:
  // Chain (base) -> Auth -> Identity -> MFA -> Marketplace, Provider, HPC
  return (
    <PortalContext.Provider value={value}>
      <ChainProvider config={chainConfig}>
        <AuthProvider config={config}>
          <IdentityProvider>
            <MFAProvider>
              <MarketplaceProvider>
                <ProviderProvider>
                  <HPCProvider>
                    {children}
                  </HPCProvider>
                </ProviderProvider>
              </MarketplaceProvider>
            </MFAProvider>
          </IdentityProvider>
        </AuthProvider>
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
    throw new Error('usePortal must be used within a PortalProvider');
  }
  return context;
}
