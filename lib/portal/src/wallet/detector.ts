/**
 * Wallet Detection and Priority System
 * VE-700: Comprehensive wallet detection for VirtEngine Portal
 *
 * Handles:
 * - SSR-safe wallet detection
 * - Race condition handling for wallet injection
 * - Multiple wallet detection
 * - Mobile browser detection
 * - Chain support verification
 */

import type { WalletType } from './types';

/**
 * Result of detecting a single wallet
 */
export interface WalletDetectionResult {
  walletType: WalletType;
  isInstalled: boolean;
  version?: string;
  supportsChain?: boolean;
}

/**
 * Wallet priority configuration - ordered by preference
 * Higher index = lower priority
 */
export const WalletPriority: readonly WalletType[] = [
  'keplr',
  'leap',
  'cosmostation',
  'walletconnect',
] as const;

/**
 * Download URLs for each wallet type
 */
const WALLET_DOWNLOAD_URLS: Record<WalletType, string> = {
  keplr: 'https://www.keplr.app/download',
  leap: 'https://www.leapwallet.io/download',
  cosmostation: 'https://www.cosmostation.io/wallet',
  walletconnect: 'https://walletconnect.com/',
};

/**
 * Wallet global object keys on window
 */
const WALLET_GLOBAL_KEYS: Record<Exclude<WalletType, 'walletconnect'>, string> = {
  keplr: 'keplr',
  leap: 'leap',
  cosmostation: 'cosmostation',
};

/**
 * Mobile browser detection patterns
 */
const MOBILE_PATTERNS = [
  /Android/i,
  /webOS/i,
  /iPhone/i,
  /iPad/i,
  /iPod/i,
  /BlackBerry/i,
  /Windows Phone/i,
];

/**
 * Type guard for window.keplr
 */
interface KeplrWindow {
  keplr?: {
    version?: string;
    getChainInfosWithoutEndpoints?: () => Promise<{ chainId: string }[]>;
    experimentalSuggestChain?: (chainInfo: unknown) => Promise<void>;
    enable?: (chainId: string) => Promise<void>;
  };
}

/**
 * Type guard for window.leap
 */
interface LeapWindow {
  leap?: {
    version?: string;
    getSupportedChains?: () => Promise<string[]>;
    enable?: (chainId: string) => Promise<void>;
  };
}

/**
 * Type guard for window.cosmostation
 */
interface CosmostationWindow {
  cosmostation?: {
    version?: string;
    providers?: {
      keplr?: {
        version?: string;
      };
    };
    cosmos?: {
      request?: (options: { method: string; params?: unknown }) => Promise<unknown>;
    };
  };
}

/**
 * Combined window type for wallet detection
 */
type WalletWindow = Window & KeplrWindow & LeapWindow & CosmostationWindow;

/**
 * Check if code is running in browser environment
 */
function isBrowser(): boolean {
  return typeof window !== 'undefined' && typeof document !== 'undefined';
}

/**
 * Check if running on mobile browser
 */
function isMobile(): boolean {
  if (!isBrowser()) return false;
  const userAgent = navigator.userAgent || navigator.vendor || '';
  return MOBILE_PATTERNS.some((pattern) => pattern.test(userAgent));
}

/**
 * Get wallet window safely
 */
function getWalletWindow(): WalletWindow | null {
  if (!isBrowser()) return null;
  return window as WalletWindow;
}

/**
 * WalletDetector class - handles comprehensive wallet detection
 */
export class WalletDetector {
  private readonly isMobileDevice: boolean;
  private detectionCache: Map<WalletType, WalletDetectionResult> = new Map();
  private cacheTimestamp = 0;
  private readonly cacheTTL = 5000; // 5 seconds cache TTL

  constructor() {
    this.isMobileDevice = isMobile();
  }

  /**
   * Clear detection cache
   */
  clearCache(): void {
    this.detectionCache.clear();
    this.cacheTimestamp = 0;
  }

  /**
   * Check if cache is valid
   */
  private isCacheValid(): boolean {
    return Date.now() - this.cacheTimestamp < this.cacheTTL;
  }

  /**
   * Detect all installed wallets
   */
  async detectInstalledWallets(): Promise<WalletDetectionResult[]> {
    if (this.isCacheValid() && this.detectionCache.size > 0) {
      return Array.from(this.detectionCache.values());
    }

    const results: WalletDetectionResult[] = [];

    for (const walletType of WalletPriority) {
      const result = await this.detectWallet(walletType);
      results.push(result);
      this.detectionCache.set(walletType, result);
    }

    this.cacheTimestamp = Date.now();
    return results;
  }

  /**
   * Detect a specific wallet
   */
  private async detectWallet(type: WalletType): Promise<WalletDetectionResult> {
    const isInstalled = this.isWalletInstalled(type);
    const version = isInstalled ? await this.getWalletVersion(type) : undefined;

    return {
      walletType: type,
      isInstalled,
      version: version ?? undefined,
      supportsChain: undefined, // Chain support checked separately
    };
  }

  /**
   * Check if a specific wallet is installed
   */
  isWalletInstalled(type: WalletType): boolean {
    const win = getWalletWindow();
    if (!win) return false;

    switch (type) {
      case 'keplr':
        return this.isKeplrInstalled(win);
      case 'leap':
        return this.isLeapInstalled(win);
      case 'cosmostation':
        return this.isCosmostationInstalled(win);
      case 'walletconnect':
        // WalletConnect is always "available" as it's a protocol
        return true;
      default:
        return false;
    }
  }

  /**
   * Check if Keplr wallet is installed
   */
  private isKeplrInstalled(win: WalletWindow): boolean {
    // Check for Keplr extension
    if (win.keplr) return true;

    // On mobile, check if we're in Keplr's in-app browser
    if (this.isMobileDevice) {
      const userAgent = navigator.userAgent || '';
      if (userAgent.includes('Keplr')) return true;
    }

    return false;
  }

  /**
   * Check if Leap wallet is installed
   */
  private isLeapInstalled(win: WalletWindow): boolean {
    // Check for Leap extension
    if (win.leap) return true;

    // On mobile, check if we're in Leap's in-app browser
    if (this.isMobileDevice) {
      const userAgent = navigator.userAgent || '';
      if (userAgent.includes('LeapCosmos')) return true;
    }

    return false;
  }

  /**
   * Check if Cosmostation wallet is installed
   */
  private isCosmostationInstalled(win: WalletWindow): boolean {
    // Check for Cosmostation extension
    if (win.cosmostation) return true;

    // Cosmostation may also expose itself via Keplr-compatible API
    if (win.keplr && (win.keplr as unknown as { isCosmostationWallet?: boolean }).isCosmostationWallet) {
      return true;
    }

    // On mobile, check if we're in Cosmostation's in-app browser
    if (this.isMobileDevice) {
      const userAgent = navigator.userAgent || '';
      if (userAgent.includes('Cosmostation')) return true;
    }

    return false;
  }

  /**
   * Get wallet version string
   */
  async getWalletVersion(type: WalletType): Promise<string | null> {
    const win = getWalletWindow();
    if (!win) return null;

    try {
      switch (type) {
        case 'keplr':
          return win.keplr?.version ?? null;
        case 'leap':
          return win.leap?.version ?? null;
        case 'cosmostation':
          return win.cosmostation?.version ?? win.cosmostation?.providers?.keplr?.version ?? null;
        case 'walletconnect':
          // WalletConnect version is determined by the SDK, not browser extension
          return null;
        default:
          return null;
      }
    } catch {
      return null;
    }
  }

  /**
   * Check if wallet supports a specific chain
   */
  async checkChainSupport(type: WalletType, chainId: string): Promise<boolean> {
    const win = getWalletWindow();
    if (!win) return false;

    if (!this.isWalletInstalled(type)) return false;

    try {
      switch (type) {
        case 'keplr':
          return this.checkKeplrChainSupport(win, chainId);
        case 'leap':
          return this.checkLeapChainSupport(win, chainId);
        case 'cosmostation':
          return this.checkCosmostationChainSupport(win, chainId);
        case 'walletconnect':
          // WalletConnect supports any chain the connected wallet supports
          return true;
        default:
          return false;
      }
    } catch {
      return false;
    }
  }

  /**
   * Check Keplr chain support
   */
  private async checkKeplrChainSupport(win: WalletWindow, chainId: string): Promise<boolean> {
    if (!win.keplr) return false;

    try {
      // Try to get chain infos if available
      if (win.keplr.getChainInfosWithoutEndpoints) {
        const chains = await win.keplr.getChainInfosWithoutEndpoints();
        return chains.some((chain) => chain.chainId === chainId);
      }

      // Fallback: try to enable the chain - this may prompt user
      // so we only do this if explicitly checking support
      if (win.keplr.experimentalSuggestChain) {
        // Chain can be suggested, so technically any chain is "supported"
        return true;
      }

      return false;
    } catch {
      return false;
    }
  }

  /**
   * Check Leap chain support
   */
  private async checkLeapChainSupport(win: WalletWindow, chainId: string): Promise<boolean> {
    if (!win.leap) return false;

    try {
      if (win.leap.getSupportedChains) {
        const chains = await win.leap.getSupportedChains();
        return chains.includes(chainId);
      }

      // Leap generally supports most Cosmos chains
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Check Cosmostation chain support
   */
  private async checkCosmostationChainSupport(win: WalletWindow, chainId: string): Promise<boolean> {
    if (!win.cosmostation) return false;

    try {
      if (win.cosmostation.cosmos?.request) {
        const result = await win.cosmostation.cosmos.request({
          method: 'cos_supportedChainIds',
        }) as { official?: string[]; unofficial?: string[] };

        const allChains = [...(result.official || []), ...(result.unofficial || [])];
        return allChains.includes(chainId);
      }

      return false;
    } catch {
      return false;
    }
  }

  /**
   * Get the best available wallet based on priority
   */
  getBestAvailableWallet(): WalletType | null {
    for (const walletType of WalletPriority) {
      if (this.isWalletInstalled(walletType)) {
        // Skip WalletConnect if native wallets are available
        if (walletType === 'walletconnect') continue;
        return walletType;
      }
    }

    // Fall back to WalletConnect if no native wallet is installed
    return 'walletconnect';
  }

  /**
   * Get list of all installed wallets (excluding WalletConnect)
   */
  getInstalledNativeWallets(): WalletType[] {
    return WalletPriority.filter((type) => {
      if (type === 'walletconnect') return false;
      return this.isWalletInstalled(type);
    });
  }

  /**
   * Wait for a wallet to be available (handles race conditions)
   */
  async waitForWallet(type: WalletType, timeout = 3000): Promise<boolean> {
    // WalletConnect is always available
    if (type === 'walletconnect') return true;

    // If not in browser, wallet will never be available
    if (!isBrowser()) return false;

    // Check if already installed
    if (this.isWalletInstalled(type)) return true;

    const globalKey = WALLET_GLOBAL_KEYS[type];
    if (!globalKey) return false;

    return new Promise((resolve) => {
      const startTime = Date.now();
      const checkInterval = 100; // Check every 100ms

      const check = (): void => {
        // Check if wallet is now available
        if (this.isWalletInstalled(type)) {
          this.clearCache();
          resolve(true);
          return;
        }

        // Check timeout
        if (Date.now() - startTime >= timeout) {
          resolve(false);
          return;
        }

        // Schedule next check
        setTimeout(check, checkInterval);
      };

      // Also listen for wallet injection events
      const handleWalletEvent = (): void => {
        if (this.isWalletInstalled(type)) {
          this.clearCache();
          window.removeEventListener(`${globalKey}_keystorechange`, handleWalletEvent);
          resolve(true);
        }
      };

      // Some wallets dispatch events when ready
      window.addEventListener(`${globalKey}_keystorechange`, handleWalletEvent);

      // Start polling
      check();
    });
  }

  /**
   * Wait for any wallet to become available
   */
  async waitForAnyWallet(timeout = 3000): Promise<WalletType | null> {
    if (!isBrowser()) return null;

    // Check if any wallet is already installed
    const existing = this.getBestAvailableWallet();
    if (existing && existing !== 'walletconnect') return existing;

    return new Promise((resolve) => {
      const startTime = Date.now();
      const checkInterval = 100;

      const check = (): void => {
        const wallet = this.getBestAvailableWallet();
        if (wallet && wallet !== 'walletconnect') {
          this.clearCache();
          resolve(wallet);
          return;
        }

        if (Date.now() - startTime >= timeout) {
          // Return WalletConnect as fallback
          resolve('walletconnect');
          return;
        }

        setTimeout(check, checkInterval);
      };

      check();
    });
  }

  /**
   * Get download URL for a wallet
   */
  getDownloadUrl(type: WalletType): string {
    return WALLET_DOWNLOAD_URLS[type];
  }

  /**
   * Get all wallet download URLs
   */
  getAllDownloadUrls(): Record<WalletType, string> {
    return { ...WALLET_DOWNLOAD_URLS };
  }

  /**
   * Check if running on mobile device
   */
  isMobileBrowser(): boolean {
    return this.isMobileDevice;
  }

  /**
   * Get mobile deep link for wallet
   */
  getMobileDeepLink(type: WalletType, uri?: string): string | null {
    if (!this.isMobileDevice) return null;

    switch (type) {
      case 'keplr':
        return uri
          ? `keplr://wcV2?${encodeURIComponent(uri)}`
          : 'keplr://';
      case 'leap':
        return uri
          ? `leapcosmos://wcV2?uri=${encodeURIComponent(uri)}`
          : 'leapcosmos://';
      case 'cosmostation':
        return uri
          ? `cosmostation://wcV2?uri=${encodeURIComponent(uri)}`
          : 'cosmostation://';
      case 'walletconnect':
        return null;
      default:
        return null;
    }
  }

  /**
   * Detect if we're inside a wallet's in-app browser
   */
  detectInAppBrowser(): WalletType | null {
    if (!isBrowser() || !this.isMobileDevice) return null;

    const userAgent = navigator.userAgent || '';

    if (userAgent.includes('Keplr')) return 'keplr';
    if (userAgent.includes('LeapCosmos')) return 'leap';
    if (userAgent.includes('Cosmostation')) return 'cosmostation';

    return null;
  }

  /**
   * Get wallet display name
   */
  getWalletDisplayName(type: WalletType): string {
    const names: Record<WalletType, string> = {
      keplr: 'Keplr',
      leap: 'Leap',
      cosmostation: 'Cosmostation',
      walletconnect: 'WalletConnect',
    };
    return names[type];
  }

  /**
   * Get wallet icon URL
   */
  getWalletIconUrl(type: WalletType): string {
    const icons: Record<WalletType, string> = {
      keplr: 'https://raw.githubusercontent.com/cosmos/chain-registry/master/wallets/keplr-extension/images/keplr-icon.png',
      leap: 'https://raw.githubusercontent.com/cosmos/chain-registry/master/wallets/leap-extension/images/leap-logo.png',
      cosmostation: 'https://raw.githubusercontent.com/cosmos/chain-registry/master/wallets/cosmostation-extension/images/cosmostation-logo.png',
      walletconnect: 'https://raw.githubusercontent.com/WalletConnect/walletconnect-assets/master/Icon/Blue%20(Default)/Icon.png',
    };
    return icons[type];
  }
}

/**
 * Singleton instance of WalletDetector
 */
export const walletDetector = new WalletDetector();

/**
 * Default export for convenience
 */
export default walletDetector;
