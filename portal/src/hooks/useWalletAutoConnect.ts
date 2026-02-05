/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useEffect, useRef, useCallback } from 'react';

import { useWallet } from '@/lib/portal-adapter';
import type { PortalWalletType } from '@/lib/portal-adapter';

/**
 * Configuration for auto-connect behavior.
 */
export interface WalletAutoConnectConfig {
  /** Whether auto-connect is enabled (default: true) */
  enabled?: boolean;
  /** Maximum number of reconnection attempts (default: 3) */
  maxRetries?: number;
  /** Delay between retry attempts in ms (default: 1000) */
  retryDelayMs?: number;
  /** Storage key for session persistence */
  storageKey?: string;
}

/**
 * Return type for the useWalletAutoConnect hook.
 */
export interface UseWalletAutoConnectResult {
  /** Whether auto-connect is currently in progress */
  isAutoConnecting: boolean;
  /** Whether auto-connect has been attempted */
  hasAttemptedAutoConnect: boolean;
}

/**
 * Stored session data shape.
 */
interface StoredSession {
  walletType: PortalWalletType;
  chainId?: string;
  lastConnectedAt?: number;
  autoConnect?: boolean;
}

/**
 * Default configuration values.
 */
const DEFAULT_CONFIG: Required<WalletAutoConnectConfig> = {
  enabled: true,
  maxRetries: 3,
  retryDelayMs: 1000,
  storageKey: 've_wallet_session',
};

/**
 * Session age threshold in ms (7 days).
 */
const MAX_SESSION_AGE_MS = 7 * 24 * 60 * 60 * 1000;

/**
 * Validates that a wallet type is valid.
 */
function isValidWalletType(type: unknown): type is PortalWalletType {
  return (
    typeof type === 'string' && ['keplr', 'leap', 'cosmostation', 'walletconnect'].includes(type)
  );
}

/**
 * Validates stored session data.
 */
function isValidSession(data: unknown): data is StoredSession {
  if (!data || typeof data !== 'object') {
    return false;
  }

  const session = data as Record<string, unknown>;

  if (!isValidWalletType(session.walletType)) {
    return false;
  }

  // Check if autoConnect is explicitly disabled
  if (session.autoConnect === false) {
    return false;
  }

  // Check session age
  if (typeof session.lastConnectedAt === 'number') {
    const age = Date.now() - session.lastConnectedAt;
    if (age > MAX_SESSION_AGE_MS) {
      return false;
    }
  }

  return true;
}

/**
 * Loads session from storage.
 */
function loadSession(storageKey: string): StoredSession | null {
  if (typeof window === 'undefined') {
    return null;
  }

  try {
    const stored = window.localStorage.getItem(storageKey);
    if (!stored) {
      return null;
    }

    const parsed: unknown = JSON.parse(stored);
    if (!isValidSession(parsed)) {
      window.localStorage.removeItem(storageKey);
      return null;
    }

    return parsed;
  } catch {
    return null;
  }
}

/**
 * Delays execution for a specified duration.
 */
function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Hook for automatic wallet reconnection on page load.
 *
 * Checks for a valid session on mount and attempts to reconnect
 * if one exists. Handles retry logic with exponential backoff
 * and graceful error handling.
 *
 * @param config - Optional configuration for auto-connect behavior
 * @returns Object with connection status
 *
 * @example
 * ```tsx
 * function App() {
 *   const { isAutoConnecting, hasAttemptedAutoConnect } = useWalletAutoConnect({
 *     enabled: true,
 *     maxRetries: 3,
 *   });
 *
 *   if (isAutoConnecting) {
 *     return <LoadingSpinner />;
 *   }
 *
 *   return <WalletButton />;
 * }
 * ```
 *
 * @example
 * ```tsx
 * // Disable auto-connect
 * function ManualConnectOnly() {
 *   useWalletAutoConnect({ enabled: false });
 *   return <WalletButton />;
 * }
 * ```
 */
export function useWalletAutoConnect(config?: WalletAutoConnectConfig): UseWalletAutoConnectResult {
  const wallet = useWallet();
  const [isAutoConnecting, setIsAutoConnecting] = useState(false);
  const [hasAttemptedAutoConnect, setHasAttemptedAutoConnect] = useState(false);

  const configRef = useRef<Required<WalletAutoConnectConfig>>({
    ...DEFAULT_CONFIG,
    ...config,
  });

  // Update config ref when props change
  useEffect(() => {
    configRef.current = { ...DEFAULT_CONFIG, ...config };
  }, [config]);

  const attemptReconnect = useCallback(async () => {
    const { enabled, maxRetries, retryDelayMs, storageKey } = configRef.current;

    if (!enabled) {
      setHasAttemptedAutoConnect(true);
      return;
    }

    // Skip if already connected or connecting
    if (wallet.status === 'connected' || wallet.status === 'connecting') {
      setHasAttemptedAutoConnect(true);
      return;
    }

    const session = loadSession(storageKey);
    if (!session) {
      setHasAttemptedAutoConnect(true);
      return;
    }

    setIsAutoConnecting(true);

    let lastError: unknown = null;
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        await wallet.connect(session.walletType);
        setIsAutoConnecting(false);
        setHasAttemptedAutoConnect(true);
        return;
      } catch (err) {
        lastError = err;

        // Don't retry on user rejection
        if (err instanceof Error) {
          const message = err.message.toLowerCase();
          if (
            message.includes('rejected') ||
            message.includes('denied') ||
            message.includes('not installed')
          ) {
            break;
          }
        }

        // Wait before retrying (with exponential backoff)
        if (attempt < maxRetries - 1) {
          await delay(retryDelayMs * Math.pow(2, attempt));
        }
      }
    }

    // Log error but don't throw - auto-connect failures should be silent
    if (lastError && process.env.NODE_ENV === 'development') {
      // eslint-disable-next-line no-console
      console.debug('[useWalletAutoConnect] Auto-connect failed:', lastError);
    }

    setIsAutoConnecting(false);
    setHasAttemptedAutoConnect(true);
  }, [wallet]);

  // Attempt reconnection on mount
  useEffect(() => {
    // Only run once
    if (hasAttemptedAutoConnect) {
      return;
    }

    // Delay to ensure wallet extensions are loaded
    const timer = setTimeout(() => {
      void attemptReconnect();
    }, 100);

    return () => clearTimeout(timer);
  }, [attemptReconnect, hasAttemptedAutoConnect]);

  return {
    isAutoConnecting,
    hasAttemptedAutoConnect,
  };
}
