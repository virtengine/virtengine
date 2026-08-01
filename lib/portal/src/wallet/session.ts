/**
 * Wallet reconnect metadata and in-memory authorization management.
 * Persistent storage is untrusted and never contains authorization material.
 */

import type { WalletType } from './types';

export interface WalletSession {
  walletType: WalletType;
  address: string;
  chainId: string;
  connectedAt: number;
  lastActiveAt: number;
  expiresAt: number | null;
  autoReconnect: boolean;
}

export interface SessionConfig {
  persistKey: string;
  maxAge: number;
  autoReconnect: boolean;
}

export interface MfaAuthorization {
  scopes: readonly string[];
  expiresAt: number;
}

export interface WalletAuthorizationBinding {
  chainId: string;
  account: string;
  publicKey: string;
  walletType: WalletType;
  deviceId: string;
  sessionId: string;
  issuedAt: number;
  expiresAt: number;
  mfa: MfaAuthorization;
}

export type WalletSessionInvalidationReason = 'removed' | 'tampered' | 'disconnect';

export type WalletAuthorizationContext = Pick<
  WalletAuthorizationBinding,
  'chainId' | 'account' | 'publicKey' | 'walletType' | 'deviceId' | 'sessionId'
>;

interface PersistedReconnectMetadata {
  version: 2;
  walletType: WalletType;
  address: string;
  chainId: string;
  connectedAt: number;
  lastActiveAt: number;
  expiresAt: number;
  autoReconnect: boolean;
}

const DEFAULT_CONFIG: SessionConfig = {
  persistKey: 'virtengine_wallet_session',
  maxAge: 7 * 24 * 60 * 60 * 1000,
  autoReconnect: true,
};
const WALLET_TYPES: readonly WalletType[] = [
  'keplr',
  'leap',
  'cosmostation',
  'walletconnect',
];
const RECONNECT_KEYS = [
  'version',
  'walletType',
  'address',
  'chainId',
  'connectedAt',
  'lastActiveAt',
  'expiresAt',
  'autoReconnect',
] as const;

const memoryStorage = new Map<string, string>();

function isStorageAvailable(): boolean {
  if (typeof window === 'undefined') return false;
  try {
    const key = '__virtengine_storage_test__';
    window.localStorage.setItem(key, 'test');
    window.localStorage.removeItem(key);
    return true;
  } catch {
    return false;
  }
}

const storage = {
  getItem(key: string): string | null {
    return isStorageAvailable()
      ? window.localStorage.getItem(key)
      : memoryStorage.get(key) ?? null;
  },
  setItem(key: string, value: string): void {
    if (isStorageAvailable()) window.localStorage.setItem(key, value);
    else memoryStorage.set(key, value);
  },
  removeItem(key: string): void {
    if (isStorageAvailable()) window.localStorage.removeItem(key);
    else memoryStorage.delete(key);
  },
};

function isNonEmptyString(value: unknown): value is string {
  return typeof value === 'string' && value.length > 0;
}

function isTimestamp(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0;
}

function isWalletType(value: unknown): value is WalletType {
  return typeof value === 'string' && WALLET_TYPES.includes(value as WalletType);
}

function hasExactKeys(value: Record<string, unknown>): boolean {
  const actual = Object.keys(value).sort();
  const expected = [...RECONNECT_KEYS].sort();
  return actual.length === expected.length && actual.every((key, i) => key === expected[i]);
}

function isReconnectMetadata(value: unknown): value is PersistedReconnectMetadata {
  if (typeof value !== 'object' || value === null) return false;
  const metadata = value as Record<string, unknown>;
  return (
    hasExactKeys(metadata) &&
    metadata.version === 2 &&
    isWalletType(metadata.walletType) &&
    isNonEmptyString(metadata.address) &&
    isNonEmptyString(metadata.chainId) &&
    isTimestamp(metadata.connectedAt) &&
    isTimestamp(metadata.lastActiveAt) &&
    isTimestamp(metadata.expiresAt) &&
    metadata.connectedAt <= metadata.lastActiveAt &&
    typeof metadata.autoReconnect === 'boolean'
  );
}

function isAuthorizationBinding(value: WalletAuthorizationBinding): boolean {
  const now = Date.now();
  return (
    isNonEmptyString(value.chainId) &&
    isNonEmptyString(value.account) &&
    isNonEmptyString(value.publicKey) &&
    isWalletType(value.walletType) &&
    isNonEmptyString(value.deviceId) &&
    isNonEmptyString(value.sessionId) &&
    isTimestamp(value.issuedAt) &&
    isTimestamp(value.expiresAt) &&
    value.issuedAt <= now &&
    value.issuedAt < value.expiresAt &&
    value.expiresAt > now &&
    Array.isArray(value.mfa?.scopes) &&
    value.mfa.scopes.length > 0 &&
    value.mfa.scopes.every(isNonEmptyString) &&
    new Set(value.mfa.scopes).size === value.mfa.scopes.length &&
    isTimestamp(value.mfa.expiresAt) &&
    value.mfa.expiresAt > now &&
    value.mfa.expiresAt <= value.expiresAt
  );
}

export class WalletSessionManager {
  private readonly config: SessionConfig;
  private cachedSession: WalletSession | null = null;
  private liveAuthorization: WalletAuthorizationBinding | null = null;
  private expectedContext: Partial<WalletAuthorizationContext> = {};
  private readonly invalidationListeners = new Set<
    (reason: WalletSessionInvalidationReason) => void
  >();
  private readonly storageListener: ((event: StorageEvent) => void) | null;

  constructor(config: Partial<SessionConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
    this.storageListener =
      typeof window !== 'undefined' && typeof window.addEventListener === 'function'
        ? event => this.handleStorageEvent(event)
        : null;
    if (this.storageListener) window.addEventListener('storage', this.storageListener);
  }

  dispose(): void {
    if (this.storageListener && typeof window?.removeEventListener === 'function') {
      window.removeEventListener('storage', this.storageListener);
    }
    this.liveAuthorization = null;
    this.invalidationListeners.clear();
  }

  onInvalidated(listener: (reason: WalletSessionInvalidationReason) => void): () => void {
    this.invalidationListeners.add(listener);
    return () => this.invalidationListeners.delete(listener);
  }

  setExpectedChainId(chainId: string): void {
    this.setExpectedContext({ ...this.expectedContext, chainId });
  }

  setExpectedContext(context: Partial<WalletAuthorizationContext>): void {
    this.expectedContext = { ...context };
    if (this.liveAuthorization && !this.matchesExpectedContext(this.liveAuthorization)) {
      this.liveAuthorization = null;
    }
    if (this.cachedSession && !this.matchesReconnectContext(this.cachedSession)) {
      this.clearInvalidStorage();
    }
  }

  saveSession(session: WalletSession): void {
    try {
      const now = Date.now();
      const metadata: PersistedReconnectMetadata = {
        version: 2,
        walletType: session.walletType,
        address: session.address,
        chainId: session.chainId,
        connectedAt: session.connectedAt,
        lastActiveAt: now,
        expiresAt: session.expiresAt ?? now + this.config.maxAge,
        autoReconnect: session.autoReconnect,
      };
      if (!isReconnectMetadata(metadata)) throw new Error('Invalid reconnect metadata');
      if (
        this.liveAuthorization &&
        (this.liveAuthorization.chainId !== metadata.chainId ||
          this.liveAuthorization.account !== metadata.address ||
          this.liveAuthorization.walletType !== metadata.walletType)
      ) {
        this.liveAuthorization = null;
      }
      storage.setItem(this.config.persistKey, JSON.stringify(metadata));
      this.cachedSession = this.toSession(metadata);
    } catch (error) {
      this.clearInvalidStorage();
      console.warn('[WalletSessionManager] Failed to save reconnect metadata:', error);
    }
  }

  loadSession(): WalletSession | null {
    try {
      const stored = storage.getItem(this.config.persistKey);
      if (!stored) {
        this.cachedSession = null;
        return null;
      }
      const parsed: unknown = JSON.parse(stored);
      if (!isReconnectMetadata(parsed) || parsed.expiresAt <= Date.now()) {
        console.warn('[WalletSessionManager] Invalid reconnect metadata, clearing');
        this.clearInvalidStorage();
        return null;
      }
      const session = this.toSession(parsed);
      if (!this.matchesReconnectContext(session)) {
        this.clearInvalidStorage();
        return null;
      }
      this.cachedSession = session;
      return session;
    } catch (error) {
      console.warn('[WalletSessionManager] Invalid reconnect metadata, clearing:', error);
      this.clearInvalidStorage();
      return null;
    }
  }

  clearSession(): void {
    this.clearInvalidStorage();
    try {
      storage.setItem(this.disconnectKey, String(Date.now()));
      storage.removeItem(this.disconnectKey);
    } catch {
      // Local state remains cleared if a cross-tab broadcast cannot be written.
    }
  }

  setLiveAuthorization(binding: WalletAuthorizationBinding): boolean {
    this.liveAuthorization = null;
    if (!isAuthorizationBinding(binding) || !this.matchesExpectedContext(binding)) return false;
    const reconnect = this.cachedSession ?? this.loadSession();
    if (
      reconnect &&
      (binding.chainId !== reconnect.chainId ||
        binding.account !== reconnect.address ||
        binding.walletType !== reconnect.walletType)
    ) {
      return false;
    }
    this.liveAuthorization = this.cloneAuthorization(binding);
    return true;
  }

  getLiveAuthorization(
    context: WalletAuthorizationContext,
    requiredScopes: readonly string[] = []
  ): WalletAuthorizationBinding | null {
    const authorization = this.liveAuthorization;
    if (
      !authorization ||
      !isAuthorizationBinding(authorization) ||
      !this.matchesContext(authorization, context) ||
      !requiredScopes.every(scope => authorization.mfa.scopes.includes(scope))
    ) {
      this.liveAuthorization = null;
      return null;
    }
    return this.cloneAuthorization(authorization);
  }

  clearLiveAuthorization(): void {
    this.liveAuthorization = null;
  }

  isSessionValid(): boolean {
    const session = this.cachedSession ?? this.loadSession();
    if (
      !session ||
      session.expiresAt === null ||
      session.expiresAt <= Date.now() ||
      !this.matchesReconnectContext(session)
    ) {
      if (session) this.clearInvalidStorage();
      return false;
    }
    return true;
  }

  refreshSession(): void {
    const session = this.cachedSession ?? this.loadSession();
    if (session && this.isSessionValid()) {
      this.saveSession({ ...session, expiresAt: Date.now() + this.config.maxAge });
    }
  }

  getSessionAge(): number {
    const session = this.cachedSession ?? this.loadSession();
    return session ? Date.now() - session.connectedAt : -1;
  }

  shouldAutoReconnect(): boolean {
    if (!this.config.autoReconnect) return false;
    const session = this.cachedSession ?? this.loadSession();
    return Boolean(session?.autoReconnect && this.isSessionValid());
  }

  updateLastActive(): void {
    const session = this.cachedSession ?? this.loadSession();
    if (session && this.isSessionValid()) this.saveSession(session);
  }

  getCachedSession(): WalletSession | null {
    return this.cachedSession;
  }

  getTimeUntilExpiry(): number {
    const session = this.cachedSession ?? this.loadSession();
    if (!session || session.expiresAt === null) return -1;
    return Math.max(0, session.expiresAt - Date.now());
  }

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

  private get disconnectKey(): string {
    return `${this.config.persistKey}:disconnect`;
  }

  private clearInvalidStorage(): void {
    storage.removeItem(this.config.persistKey);
    this.cachedSession = null;
    this.liveAuthorization = null;
  }

  private handleStorageEvent(event: StorageEvent): void {
    if (event.key === this.disconnectKey) {
      this.cachedSession = null;
      this.liveAuthorization = null;
      this.notifyInvalidated('disconnect');
    } else if (event.key === this.config.persistKey) {
      this.cachedSession = null;
      this.liveAuthorization = null;
      if (event.newValue === null) {
        this.notifyInvalidated('removed');
        return;
      }
      try {
        const parsed: unknown = JSON.parse(event.newValue);
        if (!isReconnectMetadata(parsed) || parsed.expiresAt <= Date.now()) {
          this.clearInvalidStorage();
          this.notifyInvalidated('tampered');
          return;
        }
        const session = this.toSession(parsed);
        if (!this.matchesReconnectContext(session)) {
          this.clearInvalidStorage();
          this.notifyInvalidated('tampered');
          return;
        }
        this.cachedSession = session;
      } catch {
        this.clearInvalidStorage();
        this.notifyInvalidated('tampered');
      }
    }
  }

  private notifyInvalidated(reason: WalletSessionInvalidationReason): void {
    this.invalidationListeners.forEach(listener => listener(reason));
  }

  private matchesReconnectContext(session: WalletSession): boolean {
    return (
      (!this.expectedContext.chainId || this.expectedContext.chainId === session.chainId) &&
      (!this.expectedContext.account || this.expectedContext.account === session.address) &&
      (!this.expectedContext.walletType || this.expectedContext.walletType === session.walletType)
    );
  }

  private matchesExpectedContext(binding: WalletAuthorizationBinding): boolean {
    return Object.entries(this.expectedContext).every(
      ([key, value]) => !value || binding[key as keyof WalletAuthorizationContext] === value
    );
  }

  private matchesContext(
    binding: WalletAuthorizationBinding,
    context: WalletAuthorizationContext
  ): boolean {
    return (
      binding.chainId === context.chainId &&
      binding.account === context.account &&
      binding.publicKey === context.publicKey &&
      binding.walletType === context.walletType &&
      binding.deviceId === context.deviceId &&
      binding.sessionId === context.sessionId
    );
  }

  private cloneAuthorization(binding: WalletAuthorizationBinding): WalletAuthorizationBinding {
    return { ...binding, mfa: { ...binding.mfa, scopes: [...binding.mfa.scopes] } };
  }

  private toSession(metadata: PersistedReconnectMetadata): WalletSession {
    return {
      walletType: metadata.walletType,
      address: metadata.address,
      chainId: metadata.chainId,
      connectedAt: metadata.connectedAt,
      lastActiveAt: metadata.lastActiveAt,
      expiresAt: metadata.expiresAt,
      autoReconnect: metadata.autoReconnect,
    };
  }
}

export const walletSessionManager = new WalletSessionManager();

export function createSessionManager(
  config: Partial<SessionConfig> = {}
): WalletSessionManager {
  return new WalletSessionManager(config);
}
