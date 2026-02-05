/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useCallback } from 'react';

import { toast } from './use-toast';

import type { WalletError as WalletErrorType } from '@/lib/portal-adapter';

/**
 * Error with wallet-specific properties.
 */
export interface WalletError {
  code: string;
  message: string;
  cause?: unknown;
  isRetryable?: boolean;
  suggestedAction?: string;
}

/**
 * Return type for the useWalletErrors hook.
 */
export interface UseWalletErrorsResult {
  /** Current wallet error, if any */
  error: WalletError | null;
  /** Set or clear the current error */
  setError: (error: WalletError | null) => void;
  /** Clear the current error */
  clearError: () => void;
  /** Convert an unknown error to WalletError and set it */
  handleError: (error: unknown) => void;
  /** Show a toast notification for a wallet error */
  showToast: (error: WalletError) => void;
}

/**
 * Known error message patterns to identify error types.
 */
const ERROR_PATTERNS: Array<{ pattern: RegExp; code: string; isRetryable: boolean }> = [
  { pattern: /not installed/i, code: 'WALLET_NOT_INSTALLED', isRetryable: false },
  { pattern: /window\.keplr is undefined/i, code: 'WALLET_NOT_INSTALLED', isRetryable: false },
  { pattern: /window\.leap is undefined/i, code: 'WALLET_NOT_INSTALLED', isRetryable: false },
  { pattern: /rejected/i, code: 'WALLET_CONNECTION_REJECTED', isRetryable: false },
  { pattern: /user denied/i, code: 'WALLET_CONNECTION_REJECTED', isRetryable: false },
  { pattern: /user rejected/i, code: 'SIGN_REJECTED', isRetryable: false },
  { pattern: /timeout/i, code: 'WALLET_TIMEOUT', isRetryable: true },
  { pattern: /timed out/i, code: 'WALLET_TIMEOUT', isRetryable: true },
  { pattern: /locked/i, code: 'WALLET_LOCKED', isRetryable: false },
  { pattern: /chain.*not.*supported/i, code: 'CHAIN_NOT_SUPPORTED', isRetryable: false },
  { pattern: /insufficient funds/i, code: 'INSUFFICIENT_FUNDS', isRetryable: false },
  { pattern: /insufficient balance/i, code: 'INSUFFICIENT_FUNDS', isRetryable: false },
  { pattern: /broadcast.*fail/i, code: 'BROADCAST_FAILED', isRetryable: true },
  { pattern: /network error/i, code: 'NETWORK_ERROR', isRetryable: true },
  { pattern: /failed to fetch/i, code: 'NETWORK_ERROR', isRetryable: true },
  { pattern: /session.*expired/i, code: 'SESSION_EXPIRED', isRetryable: true },
];

/**
 * User-friendly messages for error codes.
 */
const ERROR_MESSAGES: Record<string, string> = {
  WALLET_NOT_INSTALLED: 'Wallet extension is not installed. Please install and refresh.',
  WALLET_LOCKED: 'Your wallet is locked. Please unlock and try again.',
  WALLET_CONNECTION_REJECTED: 'Connection request was rejected.',
  WALLET_TIMEOUT: 'Connection timed out. Please try again.',
  CHAIN_NOT_SUPPORTED: 'This network is not supported by your wallet.',
  ACCOUNT_NOT_FOUND: 'No account found for this network.',
  SIGN_REJECTED: 'Transaction signing was rejected.',
  BROADCAST_FAILED: 'Failed to broadcast transaction.',
  INSUFFICIENT_FUNDS: 'Insufficient funds in your account.',
  NETWORK_ERROR: 'Network error occurred. Check your connection.',
  SESSION_EXPIRED: 'Your wallet session has expired. Please reconnect.',
  UNKNOWN: 'An unexpected error occurred.',
};

/**
 * Suggested actions for error codes.
 */
const SUGGESTED_ACTIONS: Record<string, string> = {
  WALLET_NOT_INSTALLED: 'Install the wallet extension from the official website.',
  WALLET_LOCKED: 'Enter your wallet password to unlock.',
  WALLET_CONNECTION_REJECTED: 'Click Connect and approve in your wallet popup.',
  WALLET_TIMEOUT: 'Refresh the page and try connecting again.',
  CHAIN_NOT_SUPPORTED: 'Add VirtEngine network to your wallet.',
  ACCOUNT_NOT_FOUND: 'Create or import an account in your wallet.',
  SIGN_REJECTED: 'Review the transaction and click Approve.',
  BROADCAST_FAILED: 'Wait a moment and retry the transaction.',
  INSUFFICIENT_FUNDS: 'Transfer funds to your account before proceeding.',
  NETWORK_ERROR: 'Check your connection and try again.',
  SESSION_EXPIRED: 'Click Connect Wallet to reconnect.',
  UNKNOWN: 'Try refreshing the page or contact support.',
};

/**
 * Parses an unknown error and converts it to a WalletError.
 */
function parseError(error: unknown): WalletError {
  // Already a WalletError-like object
  if (error && typeof error === 'object' && 'code' in error && 'message' in error) {
    const walletErr = error as WalletErrorType;
    return {
      code: walletErr.code,
      message: walletErr.message,
      cause: walletErr.cause,
      isRetryable: false,
    };
  }

  // Extract message from various error types
  let message = '';
  if (error instanceof Error) {
    message = error.message;
  } else if (typeof error === 'string') {
    message = error;
  } else if (error && typeof error === 'object' && 'message' in error) {
    message = String((error as { message: unknown }).message);
  }

  // Match against known patterns
  for (const { pattern, code, isRetryable } of ERROR_PATTERNS) {
    if (pattern.test(message)) {
      return {
        code,
        message: ERROR_MESSAGES[code] || message,
        cause: error,
        isRetryable,
        suggestedAction: SUGGESTED_ACTIONS[code],
      };
    }
  }

  // Unknown error
  return {
    code: 'UNKNOWN',
    message: message || ERROR_MESSAGES.UNKNOWN,
    cause: error,
    isRetryable: false,
    suggestedAction: SUGGESTED_ACTIONS.UNKNOWN,
  };
}

/**
 * Hook for managing wallet errors with toast notifications.
 *
 * Provides error state management, parsing of unknown errors into
 * structured WalletError objects, and toast notifications.
 *
 * @example
 * ```tsx
 * function WalletConnection() {
 *   const { error, handleError, clearError, showToast } = useWalletErrors();
 *
 *   const connect = async () => {
 *     try {
 *       await wallet.connect();
 *     } catch (err) {
 *       handleError(err);
 *       if (error) showToast(error);
 *     }
 *   };
 *
 *   return (
 *     <div>
 *       {error && <p>{error.message}</p>}
 *       <button onClick={connect}>Connect</button>
 *     </div>
 *   );
 * }
 * ```
 */
export function useWalletErrors(): UseWalletErrorsResult {
  const [error, setError] = useState<WalletError | null>(null);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const handleError = useCallback((err: unknown) => {
    const parsedError = parseError(err);
    setError(parsedError);
  }, []);

  const showToast = useCallback((walletError: WalletError) => {
    toast({
      title: 'Wallet Error',
      description: walletError.message,
      variant: 'destructive',
    });
  }, []);

  return {
    error,
    setError,
    clearError,
    handleError,
    showToast,
  };
}
