/**
 * Portal Configuration Types
 * VE-700: Portal foundation configuration
 *
 * @packageDocumentation
 */

import type { ReactNode } from 'react';
import type { ChainConfig } from './chain';
import type { WalletProviderConfig } from '../src/wallet/types';

/**
 * Portal configuration
 */
export interface PortalConfig {
  /**
   * Chain RPC endpoint (WebSocket preferred for subscriptions)
   * @example 'wss://rpc.virtengine.com'
   */
  chainEndpoint: string;

  /**
   * Chain REST endpoint for queries
   * @example 'https://api.virtengine.com'
   */
  chainRestEndpoint?: string;

  /**
   * Chain ID
   * @example 'virtengine-1'
   */
  chainId?: string;

  /**
   * Network name for display
   * @example 'VirtEngine Mainnet'
   */
  networkName?: string;

  /**
   * Enable SSO login flow
   * @default false
   */
  enableSSO?: boolean;

  /**
   * SSO provider configuration
   */
  ssoConfig?: SSOConfig;

  /**
   * Session configuration
   */
  sessionConfig?: SessionConfigOptions;

  /**
   * Enable debug logging (never logs sensitive data)
   * @default false
   */
  debug?: boolean;

  /**
   * Custom logger (must not log sensitive data)
   */
  logger?: PortalLogger;
}

/**
 * SSO provider configuration
 */
export interface SSOConfig {
  /**
   * SSO provider type
   */
  provider: 'oauth2' | 'oidc';

  /**
   * Authorization endpoint
   */
  authorizationEndpoint: string;

  /**
   * Token endpoint
   */
  tokenEndpoint: string;

  /**
   * Client ID (public)
   */
  clientId: string;

  /**
   * Redirect URI
   */
  redirectUri: string;

  /**
   * Scopes to request
   */
  scopes: string[];

  /**
   * Account binding endpoint (maps SSO to blockchain account)
   */
  accountBindingEndpoint: string;

  /**
   * Session storage key for state/nonce tracking
   * @default 've_sso_request'
   */
  stateStorageKey?: string;

  /**
   * Enforce state validation against stored request
   * @default false
   */
  enforceState?: boolean;

  /**
   * Enforce PKCE verifier validation against stored request
   * @default false
   */
  enforcePKCE?: boolean;
}

/**
 * Session configuration options
 */
export interface SessionConfigOptions {
  /**
   * Session token lifetime in seconds
   * @default 3600 (1 hour)
   */
  tokenLifetimeSeconds?: number;

  /**
   * Session refresh threshold (rotate when this much time remains)
   * @default 300 (5 minutes)
   */
  refreshThresholdSeconds?: number;

  /**
   * Enable automatic session refresh
   * @default true
   */
  autoRefresh?: boolean;

  /**
   * Cookie name for session storage
   * @default 've_session'
   */
  cookieName?: string;

  /**
   * Cookie domain
   */
  cookieDomain?: string;

  /**
   * Use secure cookies only (TLS required)
   * @default true
   */
  secureCookies?: boolean;
}

/**
 * Portal logger interface
 * CRITICAL: Never log sensitive data (keys, passwords, tokens, encrypted payloads)
 */
export interface PortalLogger {
  debug(message: string, context?: Record<string, unknown>): void;
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, context?: Record<string, unknown>): void;
}

/**
 * Default portal logger (console-based, filters sensitive fields)
 */
export const defaultLogger: PortalLogger = {
  debug: (message, context) => {
    if (process.env.NODE_ENV === 'development') {
      console.debug(`[Portal] ${message}`, filterSensitive(context));
    }
  },
  info: (message, context) => {
    console.info(`[Portal] ${message}`, filterSensitive(context));
  },
  warn: (message, context) => {
    console.warn(`[Portal] ${message}`, filterSensitive(context));
  },
  error: (message, context) => {
    console.error(`[Portal] ${message}`, filterSensitive(context));
  },
};

/**
 * Sensitive field names that should never be logged
 */
const SENSITIVE_FIELDS = new Set([
  'password',
  'secret',
  'token',
  'key',
  'privateKey',
  'mnemonic',
  'seed',
  'signature',
  'encryptedPayload',
  'encryptedData',
  'credential',
  'credentials',
  'accessToken',
  'refreshToken',
  'sessionToken',
  'apiKey',
  'authToken',
]);

/**
 * Filter sensitive fields from context before logging
 */
function filterSensitive(
  context?: Record<string, unknown>
): Record<string, unknown> | undefined {
  if (!context) return undefined;

  const filtered: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(context)) {
    const lowerKey = key.toLowerCase();
    const isSensitive = SENSITIVE_FIELDS.has(key) ||
      Array.from(SENSITIVE_FIELDS).some(s => lowerKey.includes(s.toLowerCase()));

    if (isSensitive) {
      filtered[key] = '[REDACTED]';
    } else if (typeof value === 'object' && value !== null) {
      filtered[key] = filterSensitive(value as Record<string, unknown>);
    } else {
      filtered[key] = value;
    }
  }
  return filtered;
}

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
  walletConfig?: WalletProviderConfig;

  /**
   * Child components
   */
  children: ReactNode;
}

/**
 * Default portal configuration
 */
export const defaultPortalConfig: Partial<PortalConfig> = {
  chainId: 'virtengine-1',
  networkName: 'VirtEngine',
  enableSSO: false,
  debug: false,
  sessionConfig: {
    tokenLifetimeSeconds: 3600,
    refreshThresholdSeconds: 300,
    autoRefresh: true,
    cookieName: 've_session',
    secureCookies: true,
  },
};

/**
 * Validate portal configuration
 */
export function validatePortalConfig(config: PortalConfig): string[] {
  const errors: string[] = [];

  if (!config.chainEndpoint) {
    errors.push('chainEndpoint is required');
  } else if (!config.chainEndpoint.startsWith('ws://') && !config.chainEndpoint.startsWith('wss://')) {
    errors.push('chainEndpoint should be a WebSocket URL (ws:// or wss://)');
  }

  // Enforce TLS in production
  if (process.env.NODE_ENV === 'production') {
    if (config.chainEndpoint && !config.chainEndpoint.startsWith('wss://')) {
      errors.push('chainEndpoint must use wss:// in production');
    }
    if (config.chainRestEndpoint && !config.chainRestEndpoint.startsWith('https://')) {
      errors.push('chainRestEndpoint must use https:// in production');
    }
  }

  if (config.enableSSO && !config.ssoConfig) {
    errors.push('ssoConfig is required when enableSSO is true');
  }

  if (config.ssoConfig) {
    if (!config.ssoConfig.authorizationEndpoint) {
      errors.push('ssoConfig.authorizationEndpoint is required');
    }
    if (!config.ssoConfig.clientId) {
      errors.push('ssoConfig.clientId is required');
    }
    if (!config.ssoConfig.redirectUri) {
      errors.push('ssoConfig.redirectUri is required');
    }
    if (config.ssoConfig.enforcePKCE && !config.ssoConfig.tokenEndpoint) {
      errors.push('ssoConfig.tokenEndpoint is required when enforcePKCE is true');
    }
  }

  return errors;
}
