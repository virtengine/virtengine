/**
 * Wallet Session Management for Portal
 * VE-700: Wallet-based authentication session handling
 */

import type { WalletType } from './types';

// ============================================================================
// Interfaces
// ============================================================================

/**
 * Represents a wallet session with connection and timing information.
 */
export interface WalletSession {
  /** Type of wallet connected */
  walletType: WalletType;
  /** Wallet address (bech32 encoded) */
  address: string;
  /** Chain ID the wallet is connected to */
  chainId: string;
  /** Timestamp when the wallet was connected (ms since epoch) */
  connectedAt: number;
  /** Timestamp of last activity (ms since epoch) */
  lastActiveAt: number;
  /** Timestamp when session expires (ms since epoch), null if no expiration */
  expiresAt: number | null;
  /** Whether to attempt automatic reconnection */
  autoReconnect: boolean;
}

/**
 * Configuration for session management behavior.
 */
export interface SessionConfig {
  /** Key used for localStorage persistence */
  persistKey: string;
  /** Maximum session age in milliseconds (default: 7 days) */
  maxAge: number;
  /** Whether to attempt automatic reconnection on page load */
  autoReconnect: boolean;
  /** Whether to encrypt sensitive session data */
  encryptionEnabled: boolean;
}

// ============================================================================
// Constants
// ============================================================================

const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000;
const DEFAULT_PERSIST_KEY = 'virtengine_wallet_session';

const DEFAULT_CONFIG: SessionConfig = {
  persistKey: DEFAULT_PERSIST_KEY,
  maxAge: SEVEN_DAYS_MS,
  autoReconnect: true,
  encryptionEnabled: false,
};

// ============================================================================
// Storage Abstraction
// ============================================================================

/**
 * Checks if we're in a browser environment with localStorage available.
 */
function isStorageAvailable(): boolean {
  if (typeof window === 'undefined') {
    return false;
  }
  
  try {
    const testKey = '__virtengine_storage_test__';
    window.localStorage.setItem(testKey, 'test');
    window.localStorage.removeItem(testKey);
    return true;
  } catch {
    return false;
  }
}

/**
 * In-memory fallback storage for SSR or when localStorage is unavailable.
 */
const memoryStorage: Map<string, string> = new Map();

/**
 * Storage abstraction that handles SSR and localStorage unavailability.
 */
const storage = {
  getItem(key: string): string | null {
    if (isStorageAvailable()) {
      return window.localStorage.getItem(key);
    }
    return memoryStorage.get(key) ?? null;
  },

  setItem(key: string, value: string): void {
    if (isStorageAvailable()) {
      window.localStorage.setItem(key, value);
    } else {
      memoryStorage.set(key, value);
    }
  },

  removeItem(key: string): void {
    if (isStorageAvailable()) {
      window.localStorage.removeItem(key);
    } else {
      memoryStorage.delete(key);
    }
  },
};

// ============================================================================
// Encryption Utilities
// ============================================================================

/**
 * Simple encoding for session data.
 * TODO: Implement proper encryption using Web Crypto API with AES-GCM
 * This is a placeholder that provides basic obfuscation only.
 */
function encodeData(data: string): string {
  try {
    // Base64 encode with a simple transformation
    // NOTE: This is NOT secure encryption - just obfuscation
    const base64 = btoa(encodeURIComponent(data));
    return `v1:${base64}`;
  } catch {
    return data;
  }
}

/**
 * Simple decoding for session data.
 * TODO: Implement proper decryption using Web Crypto API with AES-GCM
 */
function decodeData(encoded: string): string {
  try {
    if (encoded.startsWith('v1:')) {
      const base64 = encoded.slice(3);
      return decodeURIComponent(atob(base64));
    }
    // Handle unencoded legacy data
    return encoded;
  } catch {
    // Return original if decoding fails
    return encoded;
  }
}

// ============================================================================
// Session Validation
// ============================================================================

/**
 * Type guard to validate WalletSession structure.
 */
function isValidSessionShape(obj: unknown): obj is WalletSession {
  if (typeof obj !== 'object' || obj === null) {
    return false;
  }

  const session = obj as Record<string, unknown>;

  return (
    typeof session.walletType === 'string' &&
    ['keplr', 'leap', 'cosmostation', 'walletconnect'].includes(session.walletType) &&
    typeof session.address === 'string' &&
    session.address.length > 0 &&
    typeof session.chainId === 'string' &&
    session.chainId.length > 0 &&
    typeof session.connectedAt === 'number' &&
    session.connectedAt > 0 &&
    typeof session.lastActiveAt === 'number' &&
    session.lastActiveAt > 0 &&
    (session.expiresAt === null || typeof session.expiresAt === 'number') &&
    typeof session.autoReconnect === 'boolean'
  );
}

// ============================================================================
// WalletSessionManager Class
// ============================================================================

/**
 * Manages wallet session persistence, validation, and lifecycle.
 * Handles SSR gracefully and provides storage abstraction.
 */
export class WalletSessionManager {
  private config: SessionConfig;
  private cachedSession: WalletSession | null = null;
  private expectedChainId: string | null = null;

  constructor(config: Partial<SessionConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  /**
   * Sets the expected chain ID for session validation.
   * Sessions with mismatched chain IDs will be considered invalid.
   */
  setExpectedChainId(chainId: string): void {
    this.expectedChainId = chainId;
  }

  /**
   * Saves a wallet session to persistent storage.
   */
  saveSession(session: WalletSession): void {
    try {
      const sessionWithExpiry: WalletSession = {
        ...session,
        expiresAt: session.expiresAt ?? Date.now() + this.config.maxAge,
        lastActiveAt: Date.now(),
      };

      const serialized = JSON.stringify(sessionWithExpiry);
      const encoded = this.config.encryptionEnabled
        ? encodeData(serialized)
        : serialized;

      storage.setItem(this.config.persistKey, encoded);
      this.cachedSession = sessionWithExpiry;
    } catch (error) {
      console.warn('[WalletSessionManager] Failed to save session:', error);
    }
  }

  /**
   * Loads a wallet session from persistent storage.
   * Returns null if no session exists, is invalid, or is corrupted.
   */
  loadSession(): WalletSession | null {
    try {
      const stored = storage.getItem(this.config.persistKey);
      if (!stored) {
        this.cachedSession = null;
        return null;
      }

      const decoded = this.config.encryptionEnabled
        ? decodeData(stored)
        : stored;

      const parsed: unknown = JSON.parse(decoded);

      if (!isValidSessionShape(parsed)) {
        console.warn('[WalletSessionManager] Invalid session shape, clearing');
        this.clearSession();
        return null;
      }

      this.cachedSession = parsed;
      return parsed;
    } catch (error) {
      console.warn('[WalletSessionManager] Failed to load session:', error);
      this.clearSession();
      return null;
    }
  }

  /**
   * Clears the current session from storage.
   */
  clearSession(): void {
    storage.removeItem(this.config.persistKey);
    this.cachedSession = null;
  }

  /**
   * Checks if the current session is valid.
   * Validates expiration and chain ID match.
   */
  isSessionValid(): boolean {
    const session = this.cachedSession ?? this.loadSession();
    if (!session) {
      return false;
    }

    // Check expiration
    if (session.expiresAt !== null && Date.now() > session.expiresAt) {
      console.debug('[WalletSessionManager] Session expired');
      this.clearSession();
      return false;
    }

    // Check chain ID match if expected chain is set
    if (this.expectedChainId && session.chainId !== this.expectedChainId) {
      console.debug('[WalletSessionManager] Chain ID mismatch');
      return false;
    }

    return true;
  }

  /**
   * Refreshes the session by extending its expiration.
   * Does nothing if no valid session exists.
   */
  refreshSession(): void {
    const session = this.cachedSession ?? this.loadSession();
    if (!session) {
      return;
    }

    if (!this.isSessionValid()) {
      return;
    }

    this.saveSession({
      ...session,
      expiresAt: Date.now() + this.config.maxAge,
      lastActiveAt: Date.now(),
    });
  }

  /**
   * Returns the session age in milliseconds since connection.
   * Returns -1 if no session exists.
   */
  getSessionAge(): number {
    const session = this.cachedSession ?? this.loadSession();
    if (!session) {
      return -1;
    }
    return Date.now() - session.connectedAt;
  }

  /**
   * Determines if auto-reconnect should be attempted.
   * Considers both config and session-level settings.
   */
  shouldAutoReconnect(): boolean {
    if (!this.config.autoReconnect) {
      return false;
    }

    const session = this.cachedSession ?? this.loadSession();
    if (!session) {
      return false;
    }

    if (!session.autoReconnect) {
      return false;
    }

    return this.isSessionValid();
  }

  /**
   * Updates the last active timestamp for the current session.
   * Does nothing if no valid session exists.
   */
  updateLastActive(): void {
    const session = this.cachedSession ?? this.loadSession();
    if (!session || !this.isSessionValid()) {
      return;
    }

    try {
      const updated: WalletSession = {
        ...session,
        lastActiveAt: Date.now(),
      };

      const serialized = JSON.stringify(updated);
      const encoded = this.config.encryptionEnabled
        ? encodeData(serialized)
        : serialized;

      storage.setItem(this.config.persistKey, encoded);
      this.cachedSession = updated;
    } catch (error) {
      console.warn('[WalletSessionManager] Failed to update last active:', error);
    }
  }

  /**
   * Returns the currently cached session without loading from storage.
   */
  getCachedSession(): WalletSession | null {
    return this.cachedSession;
  }

  /**
   * Returns the time until session expires in milliseconds.
   * Returns -1 if no session or session has no expiration.
   * Returns 0 if session is already expired.
   */
  getTimeUntilExpiry(): number {
    const session = this.cachedSession ?? this.loadSession();
    if (!session || session.expiresAt === null) {
      return -1;
    }

    const remaining = session.expiresAt - Date.now();
    return remaining > 0 ? remaining : 0;
  }

  /**
   * Creates a new session from connection parameters.
   */
  createSession(params: {
    walletType: WalletType;
    address: string;
    chainId: string;
    autoReconnect?: boolean;
  }): WalletSession {
    const now = Date.now();
    return {
      walletType: params.walletType,
      address: params.address,
      chainId: params.chainId,
      connectedAt: now,
      lastActiveAt: now,
      expiresAt: now + this.config.maxAge,
      autoReconnect: params.autoReconnect ?? this.config.autoReconnect,
    };
  }
}

// ============================================================================
// Default Instance Export
// ============================================================================

/**
 * Default session manager instance for convenience.
 * Can be used directly or create custom instances with different configs.
 */
export const walletSessionManager = new WalletSessionManager();

/**
 * Creates a new session manager with custom configuration.
 */
export function createSessionManager(config: Partial<SessionConfig> = {}): WalletSessionManager {
  return new WalletSessionManager(config);
}
