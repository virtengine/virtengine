/**
 * Tests for Wallet Detection System
 * @module wallet/detector.test
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { WalletDetector, WalletPriority, type WalletDetectionResult } from '../../src/wallet/detector';
import type { WalletType } from '../../src/wallet/types';

// Mock window object for wallet detection
interface MockKeplr {
  version?: string;
  getChainInfosWithoutEndpoints?: () => Promise<{ chainId: string }[]>;
  experimentalSuggestChain?: (chainInfo: unknown) => Promise<void>;
  enable?: (chainId: string) => Promise<void>;
}

interface MockLeap {
  version?: string;
  getSupportedChains?: () => Promise<string[]>;
  enable?: (chainId: string) => Promise<void>;
}

interface MockCosmostation {
  version?: string;
  providers?: { keplr?: { version?: string } };
  cosmos?: {
    request?: (options: { method: string; params?: unknown }) => Promise<unknown>;
  };
}

interface MockWindow {
  keplr?: MockKeplr;
  leap?: MockLeap;
  cosmostation?: MockCosmostation;
  addEventListener: typeof window.addEventListener;
  removeEventListener: typeof window.removeEventListener;
}

describe('WalletDetector', () => {
  let detector: WalletDetector;
  let originalWindow: typeof globalThis.window;
  let mockWindow: MockWindow;

  beforeEach(() => {
    // Save original window
    originalWindow = globalThis.window;

    // Create mock window
    mockWindow = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    };

    // @ts-expect-error - Mocking window for tests
    globalThis.window = mockWindow;

    // Create fresh detector for each test
    detector = new WalletDetector();
  });

  afterEach(() => {
    // Restore original window
    globalThis.window = originalWindow;
    vi.restoreAllMocks();
  });

  describe('WalletPriority', () => {
    it('should define wallet priority order', () => {
      expect(WalletPriority).toEqual(['keplr', 'leap', 'cosmostation', 'walletconnect']);
    });

    it('should have keplr as highest priority', () => {
      expect(WalletPriority[0]).toBe('keplr');
    });

    it('should have walletconnect as lowest priority', () => {
      expect(WalletPriority[WalletPriority.length - 1]).toBe('walletconnect');
    });
  });

  describe('isWalletInstalled', () => {
    it('should detect keplr when window.keplr exists', () => {
      mockWindow.keplr = { version: '0.12.0' };
      
      expect(detector.isWalletInstalled('keplr')).toBe(true);
    });

    it('should return false for keplr when not installed', () => {
      delete mockWindow.keplr;
      
      expect(detector.isWalletInstalled('keplr')).toBe(false);
    });

    it('should detect leap when window.leap exists', () => {
      mockWindow.leap = { version: '0.5.0' };
      
      expect(detector.isWalletInstalled('leap')).toBe(true);
    });

    it('should return false for leap when not installed', () => {
      delete mockWindow.leap;
      
      expect(detector.isWalletInstalled('leap')).toBe(false);
    });

    it('should detect cosmostation when window.cosmostation exists', () => {
      mockWindow.cosmostation = { version: '0.3.0' };
      
      expect(detector.isWalletInstalled('cosmostation')).toBe(true);
    });

    it('should return false for cosmostation when not installed', () => {
      delete mockWindow.cosmostation;
      
      expect(detector.isWalletInstalled('cosmostation')).toBe(false);
    });

    it('should always return true for walletconnect', () => {
      expect(detector.isWalletInstalled('walletconnect')).toBe(true);
    });

    it('should return false for unknown wallet type', () => {
      expect(detector.isWalletInstalled('unknown' as WalletType)).toBe(false);
    });
  });

  describe('detectInstalledWallets', () => {
    it('should detect all installed wallets', async () => {
      mockWindow.keplr = { version: '0.12.0' };
      mockWindow.leap = { version: '0.5.0' };
      delete mockWindow.cosmostation;
      
      const results = await detector.detectInstalledWallets();
      
      expect(results).toHaveLength(4); // All priority wallets checked
      
      const keplrResult = results.find(r => r.walletType === 'keplr');
      expect(keplrResult?.isInstalled).toBe(true);
      expect(keplrResult?.version).toBe('0.12.0');
      
      const leapResult = results.find(r => r.walletType === 'leap');
      expect(leapResult?.isInstalled).toBe(true);
      expect(leapResult?.version).toBe('0.5.0');
      
      const cosmostationResult = results.find(r => r.walletType === 'cosmostation');
      expect(cosmostationResult?.isInstalled).toBe(false);
      
      const walletconnectResult = results.find(r => r.walletType === 'walletconnect');
      expect(walletconnectResult?.isInstalled).toBe(true);
    });

    it('should use cache on subsequent calls', async () => {
      mockWindow.keplr = { version: '0.12.0' };
      
      const results1 = await detector.detectInstalledWallets();
      
      // Modify window after first call
      mockWindow.keplr = { version: '0.13.0' };
      
      const results2 = await detector.detectInstalledWallets();
      
      // Should return cached result with original version
      const keplrResult = results2.find(r => r.walletType === 'keplr');
      expect(keplrResult?.version).toBe('0.12.0');
    });

    it('should clear cache when clearCache is called', async () => {
      mockWindow.keplr = { version: '0.12.0' };
      
      await detector.detectInstalledWallets();
      
      mockWindow.keplr = { version: '0.13.0' };
      detector.clearCache();
      
      const results = await detector.detectInstalledWallets();
      
      const keplrResult = results.find(r => r.walletType === 'keplr');
      expect(keplrResult?.version).toBe('0.13.0');
    });
  });

  describe('getBestAvailableWallet', () => {
    it('should return keplr when installed (highest priority)', () => {
      mockWindow.keplr = { version: '0.12.0' };
      mockWindow.leap = { version: '0.5.0' };
      
      expect(detector.getBestAvailableWallet()).toBe('keplr');
    });

    it('should return leap when keplr not installed', () => {
      delete mockWindow.keplr;
      mockWindow.leap = { version: '0.5.0' };
      
      expect(detector.getBestAvailableWallet()).toBe('leap');
    });

    it('should return cosmostation when keplr and leap not installed', () => {
      delete mockWindow.keplr;
      delete mockWindow.leap;
      mockWindow.cosmostation = { version: '0.3.0' };
      
      expect(detector.getBestAvailableWallet()).toBe('cosmostation');
    });

    it('should return walletconnect when no native wallets installed', () => {
      delete mockWindow.keplr;
      delete mockWindow.leap;
      delete mockWindow.cosmostation;
      
      expect(detector.getBestAvailableWallet()).toBe('walletconnect');
    });

    it('should skip walletconnect in priority order for native wallets', () => {
      mockWindow.cosmostation = { version: '0.3.0' };
      
      // Should return cosmostation, not walletconnect (even though walletconnect is always "installed")
      expect(detector.getBestAvailableWallet()).toBe('cosmostation');
    });
  });

  describe('getInstalledNativeWallets', () => {
    it('should return only installed native wallets', () => {
      mockWindow.keplr = { version: '0.12.0' };
      mockWindow.leap = { version: '0.5.0' };
      delete mockWindow.cosmostation;
      
      const wallets = detector.getInstalledNativeWallets();
      
      expect(wallets).toContain('keplr');
      expect(wallets).toContain('leap');
      expect(wallets).not.toContain('cosmostation');
      expect(wallets).not.toContain('walletconnect');
    });

    it('should return empty array when no native wallets installed', () => {
      delete mockWindow.keplr;
      delete mockWindow.leap;
      delete mockWindow.cosmostation;
      
      const wallets = detector.getInstalledNativeWallets();
      
      expect(wallets).toEqual([]);
    });
  });

  describe('getWalletVersion', () => {
    it('should return keplr version', async () => {
      mockWindow.keplr = { version: '0.12.0' };
      
      const version = await detector.getWalletVersion('keplr');
      
      expect(version).toBe('0.12.0');
    });

    it('should return leap version', async () => {
      mockWindow.leap = { version: '0.5.0' };
      
      const version = await detector.getWalletVersion('leap');
      
      expect(version).toBe('0.5.0');
    });

    it('should return cosmostation version', async () => {
      mockWindow.cosmostation = { version: '0.3.0' };
      
      const version = await detector.getWalletVersion('cosmostation');
      
      expect(version).toBe('0.3.0');
    });

    it('should return null for walletconnect', async () => {
      const version = await detector.getWalletVersion('walletconnect');
      
      expect(version).toBeNull();
    });

    it('should return null when wallet not installed', async () => {
      delete mockWindow.keplr;
      
      const version = await detector.getWalletVersion('keplr');
      
      expect(version).toBeNull();
    });

    it('should fallback to cosmostation provider keplr version', async () => {
      mockWindow.cosmostation = {
        providers: { keplr: { version: '0.2.0' } },
      };
      
      const version = await detector.getWalletVersion('cosmostation');
      
      expect(version).toBe('0.2.0');
    });
  });

  describe('waitForWallet', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('should return true immediately if walletconnect', async () => {
      const resultPromise = detector.waitForWallet('walletconnect');
      
      const result = await resultPromise;
      
      expect(result).toBe(true);
    });

    it('should return true immediately if wallet already installed', async () => {
      mockWindow.keplr = { version: '0.12.0' };
      
      const result = await detector.waitForWallet('keplr');
      
      expect(result).toBe(true);
    });

    it('should poll for wallet installation', async () => {
      delete mockWindow.keplr;
      
      const resultPromise = detector.waitForWallet('keplr', 500);
      
      // Advance time partially
      await vi.advanceTimersByTimeAsync(200);
      
      // Install wallet
      mockWindow.keplr = { version: '0.12.0' };
      
      // Advance time to trigger next check
      await vi.advanceTimersByTimeAsync(100);
      
      const result = await resultPromise;
      
      expect(result).toBe(true);
    });

    it('should return false on timeout', async () => {
      delete mockWindow.keplr;
      
      const resultPromise = detector.waitForWallet('keplr', 500);
      
      await vi.advanceTimersByTimeAsync(500);
      
      const result = await resultPromise;
      
      expect(result).toBe(false);
    });
  });

  describe('waitForAnyWallet', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('should return installed wallet immediately', async () => {
      mockWindow.leap = { version: '0.5.0' };
      
      const result = await detector.waitForAnyWallet();
      
      expect(result).toBe('leap');
    });

    it('should return walletconnect on timeout if no native wallets', async () => {
      delete mockWindow.keplr;
      delete mockWindow.leap;
      delete mockWindow.cosmostation;
      
      const resultPromise = detector.waitForAnyWallet(500);
      
      await vi.advanceTimersByTimeAsync(500);
      
      const result = await resultPromise;
      
      expect(result).toBe('walletconnect');
    });
  });

  describe('getDownloadUrl', () => {
    it('should return correct URL for keplr', () => {
      expect(detector.getDownloadUrl('keplr')).toBe('https://www.keplr.app/download');
    });

    it('should return correct URL for leap', () => {
      expect(detector.getDownloadUrl('leap')).toBe('https://www.leapwallet.io/download');
    });

    it('should return correct URL for cosmostation', () => {
      expect(detector.getDownloadUrl('cosmostation')).toBe('https://www.cosmostation.io/wallet');
    });

    it('should return correct URL for walletconnect', () => {
      expect(detector.getDownloadUrl('walletconnect')).toBe('https://walletconnect.com/');
    });
  });

  describe('getAllDownloadUrls', () => {
    it('should return all download URLs', () => {
      const urls = detector.getAllDownloadUrls();
      
      expect(urls).toEqual({
        keplr: 'https://www.keplr.app/download',
        leap: 'https://www.leapwallet.io/download',
        cosmostation: 'https://www.cosmostation.io/wallet',
        walletconnect: 'https://walletconnect.com/',
      });
    });
  });

  describe('getWalletDisplayName', () => {
    it('should return display name for each wallet', () => {
      expect(detector.getWalletDisplayName('keplr')).toBe('Keplr');
      expect(detector.getWalletDisplayName('leap')).toBe('Leap');
      expect(detector.getWalletDisplayName('cosmostation')).toBe('Cosmostation');
      expect(detector.getWalletDisplayName('walletconnect')).toBe('WalletConnect');
    });
  });

  describe('getWalletIconUrl', () => {
    it('should return icon URL for each wallet', () => {
      expect(detector.getWalletIconUrl('keplr')).toContain('keplr');
      expect(detector.getWalletIconUrl('leap')).toContain('leap');
      expect(detector.getWalletIconUrl('cosmostation')).toContain('cosmostation');
      expect(detector.getWalletIconUrl('walletconnect')).toContain('walletconnect');
    });
  });

  describe('isMobileBrowser', () => {
    it('should detect mobile devices', () => {
      // The mock window doesn't have navigator, so this should be false
      expect(detector.isMobileBrowser()).toBe(false);
    });
  });

  describe('getMobileDeepLink', () => {
    it('should return null for non-mobile environment', () => {
      expect(detector.getMobileDeepLink('keplr')).toBeNull();
    });

    it('should return null for walletconnect', () => {
      // @ts-expect-error - Accessing private property for test
      detector.isMobileDevice = true;
      expect(detector.getMobileDeepLink('walletconnect')).toBeNull();
    });
  });

  describe('detectInAppBrowser', () => {
    it('should return null for non-mobile environment', () => {
      expect(detector.detectInAppBrowser()).toBeNull();
    });
  });

  describe('checkChainSupport', () => {
    it('should return false when wallet not installed', async () => {
      delete mockWindow.keplr;
      
      const result = await detector.checkChainSupport('keplr', 'virtengine-1');
      
      expect(result).toBe(false);
    });

    it('should return true for walletconnect', async () => {
      const result = await detector.checkChainSupport('walletconnect', 'virtengine-1');
      
      expect(result).toBe(true);
    });

    it('should check keplr chain support via getChainInfosWithoutEndpoints', async () => {
      mockWindow.keplr = {
        version: '0.12.0',
        getChainInfosWithoutEndpoints: vi.fn().mockResolvedValue([
          { chainId: 'virtengine-1' },
          { chainId: 'cosmoshub-4' },
        ]),
      };
      
      const result = await detector.checkChainSupport('keplr', 'virtengine-1');
      
      expect(result).toBe(true);
    });

    it('should return false for unsupported chain in keplr', async () => {
      mockWindow.keplr = {
        version: '0.12.0',
        getChainInfosWithoutEndpoints: vi.fn().mockResolvedValue([
          { chainId: 'cosmoshub-4' },
        ]),
      };
      
      const result = await detector.checkChainSupport('keplr', 'virtengine-1');
      
      expect(result).toBe(false);
    });

    it('should check leap chain support', async () => {
      mockWindow.leap = {
        version: '0.5.0',
        getSupportedChains: vi.fn().mockResolvedValue(['virtengine-1', 'cosmoshub-4']),
      };
      
      const result = await detector.checkChainSupport('leap', 'virtengine-1');
      
      expect(result).toBe(true);
    });

    it('should check cosmostation chain support', async () => {
      mockWindow.cosmostation = {
        version: '0.3.0',
        cosmos: {
          request: vi.fn().mockResolvedValue({
            official: ['virtengine-1'],
            unofficial: ['testnet-1'],
          }),
        },
      };
      
      const result = await detector.checkChainSupport('cosmostation', 'virtengine-1');
      
      expect(result).toBe(true);
    });

    it('should handle errors gracefully', async () => {
      mockWindow.keplr = {
        version: '0.12.0',
        getChainInfosWithoutEndpoints: vi.fn().mockRejectedValue(new Error('Network error')),
      };
      
      const result = await detector.checkChainSupport('keplr', 'virtengine-1');
      
      expect(result).toBe(false);
    });
  });
});

describe('WalletDetector SSR safety', () => {
  let originalWindow: typeof globalThis.window;

  beforeEach(() => {
    originalWindow = globalThis.window;
  });

  afterEach(() => {
    globalThis.window = originalWindow;
  });

  it('should handle SSR environment (no window)', () => {
    // @ts-expect-error - Simulating SSR
    delete globalThis.window;
    
    const detector = new WalletDetector();
    
    expect(detector.isWalletInstalled('keplr')).toBe(false);
    expect(detector.getBestAvailableWallet()).toBe('walletconnect');
    expect(detector.isMobileBrowser()).toBe(false);
  });
});
