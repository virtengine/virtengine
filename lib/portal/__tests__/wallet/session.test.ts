/**
 * Tests for Wallet Session Management
 * @module wallet/session.test
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  WalletSessionManager,
  createSessionManager,
  walletSessionManager,
  type WalletSession,
  type SessionConfig,
} from '../../src/wallet/session';

// Mock localStorage
const createMockLocalStorage = () => {
  const store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      Object.keys(store).forEach(key => delete store[key]);
    }),
    get length() {
      return Object.keys(store).length;
    },
    key: vi.fn((index: number) => Object.keys(store)[index] ?? null),
    _store: store,
  };
};

describe('WalletSessionManager', () => {
  let manager: WalletSessionManager;
  let mockLocalStorage: ReturnType<typeof createMockLocalStorage>;
  let originalWindow: typeof globalThis.window;

  const createTestSession = (overrides: Partial<WalletSession> = {}): WalletSession => ({
    walletType: 'keplr',
    address: 'virtengine1abc123xyz',
    chainId: 'virtengine-1',
    connectedAt: Date.now(),
    lastActiveAt: Date.now(),
    expiresAt: Date.now() + 7 * 24 * 60 * 60 * 1000,
    autoReconnect: true,
    ...overrides,
  });

  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2024-01-15T12:00:00Z'));
    
    // Save original window
    originalWindow = globalThis.window;
    
    // Create mock localStorage
    mockLocalStorage = createMockLocalStorage();
    
    // Mock window with localStorage
    const mockWindow = {
      localStorage: mockLocalStorage,
    };
    
    // @ts-expect-error - Mocking window
    globalThis.window = mockWindow;
    
    // Create fresh manager
    manager = new WalletSessionManager();
  });

  afterEach(() => {
    globalThis.window = originalWindow;
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  describe('saveSession', () => {
    it('should save session to localStorage', () => {
      const session = createTestSession();
      
      manager.saveSession(session);
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'virtengine_wallet_session',
        expect.any(String)
      );
    });

    it('should serialize session as JSON', () => {
      const session = createTestSession();
      
      manager.saveSession(session);
      
      // Find the call that saves the session (not the storage test)
      const sessionCall = mockLocalStorage.setItem.mock.calls.find(
        call => call[0] === 'virtengine_wallet_session'
      );
      expect(sessionCall).toBeDefined();
      const parsed = JSON.parse(sessionCall![1]);
      
      expect(parsed.walletType).toBe('keplr');
      expect(parsed.address).toBe('virtengine1abc123xyz');
      expect(parsed.chainId).toBe('virtengine-1');
    });

    it('should update lastActiveAt on save', () => {
      const session = createTestSession({ lastActiveAt: 1000 });
      
      manager.saveSession(session);
      
      // Find the call that saves the session (not the storage test)
      const sessionCall = mockLocalStorage.setItem.mock.calls.find(
        call => call[0] === 'virtengine_wallet_session'
      );
      expect(sessionCall).toBeDefined();
      const parsed = JSON.parse(sessionCall![1]);
      
      expect(parsed.lastActiveAt).toBeGreaterThan(1000);
    });

    it('should set expiration if not provided', () => {
      const session = createTestSession({ expiresAt: null });
      
      manager.saveSession(session);
      
      // Find the call that saves the session (not the storage test)
      const sessionCall = mockLocalStorage.setItem.mock.calls.find(
        call => call[0] === 'virtengine_wallet_session'
      );
      expect(sessionCall).toBeDefined();
      const parsed = JSON.parse(sessionCall![1]);
      
      expect(parsed.expiresAt).toBeDefined();
      expect(parsed.expiresAt).toBeGreaterThan(Date.now());
    });

    it('should use custom persist key from config', () => {
      const customManager = new WalletSessionManager({ persistKey: 'custom_key' });
      const session = createTestSession();
      
      customManager.saveSession(session);
      
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'custom_key',
        expect.any(String)
      );
    });

    it('should encode session when encryption is enabled', () => {
      const encryptedManager = new WalletSessionManager({ encryptionEnabled: true });
      const session = createTestSession();
      
      encryptedManager.saveSession(session);
      
      // Find the call that saves the session (not the storage test)
      const sessionCall = mockLocalStorage.setItem.mock.calls.find(
        call => call[0] === 'virtengine_wallet_session'
      );
      expect(sessionCall).toBeDefined();
      expect(sessionCall![1]).toMatch(/^v1:/);
    });
  });

  describe('loadSession', () => {
    it('should return null when no session exists', () => {
      const session = manager.loadSession();
      
      expect(session).toBeNull();
    });

    it('should load and parse saved session', () => {
      const originalSession = createTestSession();
      manager.saveSession(originalSession);
      
      // Create new manager to clear cache
      const newManager = new WalletSessionManager();
      const loadedSession = newManager.loadSession();
      
      expect(loadedSession).not.toBeNull();
      expect(loadedSession?.walletType).toBe('keplr');
      expect(loadedSession?.address).toBe('virtengine1abc123xyz');
    });

    it('should return null for corrupted JSON', () => {
      mockLocalStorage._store['virtengine_wallet_session'] = 'not valid json';
      
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const session = manager.loadSession();
      
      expect(session).toBeNull();
      expect(consoleSpy).toHaveBeenCalled();
    });

    it('should return null for invalid session shape', () => {
      mockLocalStorage._store['virtengine_wallet_session'] = JSON.stringify({
        walletType: 'keplr',
        // Missing required fields
      });
      
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      const session = manager.loadSession();
      
      expect(session).toBeNull();
      expect(consoleSpy).toHaveBeenCalled();
    });

    it('should clear corrupted session from storage', () => {
      mockLocalStorage._store['virtengine_wallet_session'] = 'invalid';
      
      vi.spyOn(console, 'warn').mockImplementation(() => {});
      manager.loadSession();
      
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('virtengine_wallet_session');
    });

    it('should decode encrypted session', () => {
      const encryptedManager = new WalletSessionManager({ encryptionEnabled: true });
      const session = createTestSession();
      
      encryptedManager.saveSession(session);
      
      // Load with new encrypted manager
      const newEncryptedManager = new WalletSessionManager({ encryptionEnabled: true });
      const loadedSession = newEncryptedManager.loadSession();
      
      expect(loadedSession).not.toBeNull();
      expect(loadedSession?.walletType).toBe('keplr');
    });

    it('should validate walletType is valid', () => {
      mockLocalStorage._store['virtengine_wallet_session'] = JSON.stringify({
        walletType: 'invalid_wallet',
        address: 'virtengine1abc',
        chainId: 'virtengine-1',
        connectedAt: Date.now(),
        lastActiveAt: Date.now(),
        expiresAt: Date.now() + 10000,
        autoReconnect: true,
      });
      
      vi.spyOn(console, 'warn').mockImplementation(() => {});
      const session = manager.loadSession();
      
      expect(session).toBeNull();
    });

    it('should validate address is non-empty', () => {
      mockLocalStorage._store['virtengine_wallet_session'] = JSON.stringify({
        walletType: 'keplr',
        address: '',
        chainId: 'virtengine-1',
        connectedAt: Date.now(),
        lastActiveAt: Date.now(),
        expiresAt: Date.now() + 10000,
        autoReconnect: true,
      });
      
      vi.spyOn(console, 'warn').mockImplementation(() => {});
      const session = manager.loadSession();
      
      expect(session).toBeNull();
    });
  });

  describe('session expiration', () => {
    it('should detect expired session', () => {
      const session = createTestSession({
        expiresAt: Date.now() - 1000, // Already expired
      });
      manager.saveSession(session);
      
      const isValid = manager.isSessionValid();
      
      expect(isValid).toBe(false);
    });

    it('should detect valid non-expired session', () => {
      const session = createTestSession({
        expiresAt: Date.now() + 60000, // Expires in 1 minute
      });
      manager.saveSession(session);
      
      const isValid = manager.isSessionValid();
      
      expect(isValid).toBe(true);
    });

    it('should handle session with null expiration', () => {
      const session = createTestSession();
      manager.saveSession(session);
      
      // Force null expiration in storage
      const stored = JSON.parse(mockLocalStorage._store['virtengine_wallet_session']);
      stored.expiresAt = null;
      mockLocalStorage._store['virtengine_wallet_session'] = JSON.stringify(stored);
      
      // Clear cache and check
      const newManager = new WalletSessionManager();
      newManager.loadSession();
      
      expect(newManager.isSessionValid()).toBe(true);
    });

    it('should clear expired session from storage', () => {
      const session = createTestSession({
        expiresAt: Date.now() - 1000,
      });
      manager.saveSession(session);
      
      vi.spyOn(console, 'debug').mockImplementation(() => {});
      manager.isSessionValid();
      
      expect(mockLocalStorage.removeItem).toHaveBeenCalled();
    });

    it('should return time until expiry', () => {
      const expiresIn = 60000;
      const session = createTestSession({
        expiresAt: Date.now() + expiresIn,
      });
      manager.saveSession(session);
      
      const timeUntil = manager.getTimeUntilExpiry();
      
      // Allow small margin for execution time
      expect(timeUntil).toBeLessThanOrEqual(expiresIn);
      expect(timeUntil).toBeGreaterThan(expiresIn - 100);
    });

    it('should return 0 for expired session', () => {
      const session = createTestSession({
        expiresAt: Date.now() - 1000,
      });
      manager.saveSession(session);
      
      const timeUntil = manager.getTimeUntilExpiry();
      
      expect(timeUntil).toBe(0);
    });

    it('should return -1 for no session', () => {
      const timeUntil = manager.getTimeUntilExpiry();
      
      expect(timeUntil).toBe(-1);
    });
  });

  describe('chain ID validation', () => {
    it('should validate session with matching chain ID', () => {
      const session = createTestSession({ chainId: 'virtengine-1' });
      manager.saveSession(session);
      manager.setExpectedChainId('virtengine-1');
      
      expect(manager.isSessionValid()).toBe(true);
    });

    it('should invalidate session with mismatched chain ID', () => {
      const session = createTestSession({ chainId: 'virtengine-1' });
      manager.saveSession(session);
      manager.setExpectedChainId('different-chain');
      
      vi.spyOn(console, 'debug').mockImplementation(() => {});
      expect(manager.isSessionValid()).toBe(false);
    });

    it('should allow any chain ID when expected not set', () => {
      const session = createTestSession({ chainId: 'any-chain' });
      manager.saveSession(session);
      
      expect(manager.isSessionValid()).toBe(true);
    });
  });

  describe('clearSession', () => {
    it('should remove session from storage', () => {
      const session = createTestSession();
      manager.saveSession(session);
      
      manager.clearSession();
      
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('virtengine_wallet_session');
    });

    it('should clear cached session', () => {
      const session = createTestSession();
      manager.saveSession(session);
      
      manager.clearSession();
      
      expect(manager.getCachedSession()).toBeNull();
    });
  });

  describe('refreshSession', () => {
    it('should extend session expiration', () => {
      const session = createTestSession({
        expiresAt: Date.now() + 1000,
      });
      manager.saveSession(session);
      
      // Advance time
      vi.advanceTimersByTime(500);
      
      manager.refreshSession();
      
      const timeUntil = manager.getTimeUntilExpiry();
      
      // Should be reset to full maxAge
      expect(timeUntil).toBeGreaterThan(1000);
    });

    it('should update lastActiveAt', () => {
      const session = createTestSession({
        lastActiveAt: Date.now() - 10000,
      });
      manager.saveSession(session);
      
      manager.refreshSession();
      
      const cached = manager.getCachedSession();
      expect(cached?.lastActiveAt).toBeGreaterThan(Date.now() - 100);
    });

    it('should do nothing if session is invalid', () => {
      const initialCallCount = mockLocalStorage.setItem.mock.calls.length;
      
      const session = createTestSession({
        expiresAt: Date.now() - 1000,
      });
      manager.saveSession(session);
      
      const afterSaveCount = mockLocalStorage.setItem.mock.calls.filter(
        call => call[0] === 'virtengine_wallet_session'
      ).length;
      
      vi.spyOn(console, 'debug').mockImplementation(() => {});
      manager.refreshSession();
      
      // No additional session saves after initial
      const finalCount = mockLocalStorage.setItem.mock.calls.filter(
        call => call[0] === 'virtengine_wallet_session'
      ).length;
      expect(finalCount).toBe(afterSaveCount);
    });

    it('should do nothing if no session exists', () => {
      const initialSessionSaves = mockLocalStorage.setItem.mock.calls.filter(
        call => call[0] === 'virtengine_wallet_session'
      ).length;
      
      manager.refreshSession();
      
      const finalSessionSaves = mockLocalStorage.setItem.mock.calls.filter(
        call => call[0] === 'virtengine_wallet_session'
      ).length;
      expect(finalSessionSaves).toBe(initialSessionSaves);
    });
  });

  describe('getSessionAge', () => {
    it('should return session age in milliseconds', () => {
      const session = createTestSession({
        connectedAt: Date.now() - 5000,
      });
      manager.saveSession(session);
      
      // The save updates lastActiveAt, so we need to reload
      const age = manager.getSessionAge();
      
      // Age should be around 5000ms (with some margin for execution)
      expect(age).toBeGreaterThanOrEqual(4900);
      expect(age).toBeLessThan(5500);
    });

    it('should return -1 for no session', () => {
      const age = manager.getSessionAge();
      
      expect(age).toBe(-1);
    });
  });

  describe('shouldAutoReconnect', () => {
    it('should return true when all conditions met', () => {
      const session = createTestSession({ autoReconnect: true });
      manager.saveSession(session);
      
      expect(manager.shouldAutoReconnect()).toBe(true);
    });

    it('should return false when config disables auto-reconnect', () => {
      const noAutoManager = new WalletSessionManager({ autoReconnect: false });
      const session = createTestSession({ autoReconnect: true });
      noAutoManager.saveSession(session);
      
      expect(noAutoManager.shouldAutoReconnect()).toBe(false);
    });

    it('should return false when session disables auto-reconnect', () => {
      const session = createTestSession({ autoReconnect: false });
      manager.saveSession(session);
      
      expect(manager.shouldAutoReconnect()).toBe(false);
    });

    it('should return false when no session exists', () => {
      expect(manager.shouldAutoReconnect()).toBe(false);
    });

    it('should return false when session is expired', () => {
      const session = createTestSession({
        autoReconnect: true,
        expiresAt: Date.now() - 1000,
      });
      manager.saveSession(session);
      
      vi.spyOn(console, 'debug').mockImplementation(() => {});
      expect(manager.shouldAutoReconnect()).toBe(false);
    });
  });

  describe('updateLastActive', () => {
    it('should update lastActiveAt timestamp', () => {
      const session = createTestSession();
      manager.saveSession(session);
      
      // Advance time
      vi.advanceTimersByTime(5000);
      
      manager.updateLastActive();
      
      const cached = manager.getCachedSession();
      expect(cached?.lastActiveAt).toBe(Date.now());
    });

    it('should do nothing for invalid session', () => {
      const session = createTestSession({
        expiresAt: Date.now() - 1000,
      });
      manager.saveSession(session);
      
      const afterSaveCount = mockLocalStorage.setItem.mock.calls.filter(
        call => call[0] === 'virtengine_wallet_session'
      ).length;
      
      vi.spyOn(console, 'debug').mockImplementation(() => {});
      manager.updateLastActive();
      
      // No additional session saves after initial
      const finalCount = mockLocalStorage.setItem.mock.calls.filter(
        call => call[0] === 'virtengine_wallet_session'
      ).length;
      expect(finalCount).toBe(afterSaveCount);
    });
  });

  describe('createSession', () => {
    it('should create session with required params', () => {
      const session = manager.createSession({
        walletType: 'leap',
        address: 'virtengine1xyz',
        chainId: 'virtengine-1',
      });
      
      expect(session.walletType).toBe('leap');
      expect(session.address).toBe('virtengine1xyz');
      expect(session.chainId).toBe('virtengine-1');
      expect(session.connectedAt).toBe(Date.now());
      expect(session.lastActiveAt).toBe(Date.now());
      expect(session.expiresAt).toBeDefined();
    });

    it('should use default autoReconnect from config', () => {
      const session = manager.createSession({
        walletType: 'keplr',
        address: 'virtengine1abc',
        chainId: 'virtengine-1',
      });
      
      expect(session.autoReconnect).toBe(true);
    });

    it('should allow overriding autoReconnect', () => {
      const session = manager.createSession({
        walletType: 'keplr',
        address: 'virtengine1abc',
        chainId: 'virtengine-1',
        autoReconnect: false,
      });
      
      expect(session.autoReconnect).toBe(false);
    });
  });

  describe('createSessionManager factory', () => {
    it('should create manager with custom config', () => {
      const customManager = createSessionManager({
        persistKey: 'custom_session',
        maxAge: 1000,
        autoReconnect: false,
      });
      
      expect(customManager).toBeInstanceOf(WalletSessionManager);
    });
  });

  describe('default instance', () => {
    it('should export singleton instance', () => {
      expect(walletSessionManager).toBeInstanceOf(WalletSessionManager);
    });
  });
});

describe('WalletSessionManager SSR safety', () => {
  let originalWindow: typeof globalThis.window;

  beforeEach(() => {
    originalWindow = globalThis.window;
  });

  afterEach(() => {
    globalThis.window = originalWindow;
  });

  it('should handle missing window', () => {
    // @ts-expect-error - Simulating SSR
    delete globalThis.window;
    
    const manager = new WalletSessionManager();
    
    // Should not throw
    expect(() => manager.loadSession()).not.toThrow();
    expect(manager.loadSession()).toBeNull();
  });

  it('should handle localStorage not available', () => {
    // @ts-expect-error - Mock window without localStorage
    globalThis.window = {};
    
    const manager = new WalletSessionManager();
    
    // Should use memory storage fallback
    const session: WalletSession = {
      walletType: 'keplr',
      address: 'virtengine1abc',
      chainId: 'virtengine-1',
      connectedAt: Date.now(),
      lastActiveAt: Date.now(),
      expiresAt: Date.now() + 10000,
      autoReconnect: true,
    };
    
    expect(() => manager.saveSession(session)).not.toThrow();
    expect(manager.loadSession()).not.toBeNull();
  });
});
