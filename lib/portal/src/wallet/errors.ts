/**
 * Wallet Error Handling System for VirtEngine Portal
 *
 * Provides comprehensive error handling for wallet operations including:
 * - Typed error codes for all wallet failure scenarios
 * - User-friendly error messages
 * - Error classification (retryable vs non-retryable)
 * - Suggested actions for error recovery
 *
 * @module wallet/errors
 * @see VE-700: Wallet-based authentication
 */

/**
 * Enumeration of all possible wallet error codes.
 *
 * These codes cover the full lifecycle of wallet interactions:
 * - Installation and availability
 * - Connection and authentication
 * - Signing and broadcasting
 * - Network and chain issues
 * - Session management
 */
export enum WalletErrorCode {
  /** Wallet extension is not installed in the browser */
  WALLET_NOT_INSTALLED = 'WALLET_NOT_INSTALLED',

  /** Wallet is installed but locked/needs password */
  WALLET_LOCKED = 'WALLET_LOCKED',

  /** User rejected the connection request */
  WALLET_CONNECTION_REJECTED = 'WALLET_CONNECTION_REJECTED',

  /** Connection attempt timed out */
  WALLET_TIMEOUT = 'WALLET_TIMEOUT',

  /** Requested chain is not supported by the wallet */
  CHAIN_NOT_SUPPORTED = 'CHAIN_NOT_SUPPORTED',

  /** No account found for the requested chain */
  ACCOUNT_NOT_FOUND = 'ACCOUNT_NOT_FOUND',

  /** User rejected the signing request */
  SIGN_REJECTED = 'SIGN_REJECTED',

  /** Transaction broadcast failed */
  BROADCAST_FAILED = 'BROADCAST_FAILED',

  /** Account has insufficient funds for the transaction */
  INSUFFICIENT_FUNDS = 'INSUFFICIENT_FUNDS',

  /** Network communication error */
  NETWORK_ERROR = 'NETWORK_ERROR',

  /** Wallet session has expired and needs reconnection */
  SESSION_EXPIRED = 'SESSION_EXPIRED',

  /** Chain ID does not match expected value */
  INVALID_CHAIN_ID = 'INVALID_CHAIN_ID',

  /** Generic unknown error */
  UNKNOWN = 'UNKNOWN',
}

/**
 * User-friendly error messages for each wallet error code.
 *
 * These messages are designed to be shown directly to users
 * and provide clear guidance on what went wrong.
 */
export const WALLET_ERROR_MESSAGES: Record<WalletErrorCode, string> = {
  [WalletErrorCode.WALLET_NOT_INSTALLED]:
    'Wallet extension is not installed. Please install the wallet extension and refresh the page.',
  [WalletErrorCode.WALLET_LOCKED]:
    'Your wallet is locked. Please unlock your wallet and try again.',
  [WalletErrorCode.WALLET_CONNECTION_REJECTED]:
    'Connection request was rejected. Please approve the connection in your wallet to continue.',
  [WalletErrorCode.WALLET_TIMEOUT]:
    'Connection timed out. Please check your wallet and try again.',
  [WalletErrorCode.CHAIN_NOT_SUPPORTED]:
    'This blockchain network is not supported by your wallet. Please add the network to your wallet or use a different wallet.',
  [WalletErrorCode.ACCOUNT_NOT_FOUND]:
    'No account found. Please ensure you have an account set up in your wallet for this network.',
  [WalletErrorCode.SIGN_REJECTED]:
    'Transaction signing was rejected. Please approve the transaction in your wallet to proceed.',
  [WalletErrorCode.BROADCAST_FAILED]:
    'Failed to broadcast transaction. Please check your network connection and try again.',
  [WalletErrorCode.INSUFFICIENT_FUNDS]:
    'Insufficient funds in your account. Please add funds and try again.',
  [WalletErrorCode.NETWORK_ERROR]:
    'Network error occurred. Please check your internet connection and try again.',
  [WalletErrorCode.SESSION_EXPIRED]:
    'Your wallet session has expired. Please reconnect your wallet.',
  [WalletErrorCode.INVALID_CHAIN_ID]:
    'Invalid chain ID detected. Please ensure your wallet is connected to the correct network.',
  [WalletErrorCode.UNKNOWN]:
    'An unexpected error occurred. Please try again or contact support if the issue persists.',
};

/**
 * Suggested actions for each error code to help users recover.
 */
const WALLET_ERROR_SUGGESTED_ACTIONS: Record<WalletErrorCode, string> = {
  [WalletErrorCode.WALLET_NOT_INSTALLED]: 'Install the wallet extension from the official website.',
  [WalletErrorCode.WALLET_LOCKED]: 'Enter your wallet password to unlock.',
  [WalletErrorCode.WALLET_CONNECTION_REJECTED]: 'Click Connect and approve in your wallet popup.',
  [WalletErrorCode.WALLET_TIMEOUT]: 'Refresh the page and try connecting again.',
  [WalletErrorCode.CHAIN_NOT_SUPPORTED]: 'Add VirtEngine network to your wallet manually.',
  [WalletErrorCode.ACCOUNT_NOT_FOUND]: 'Create or import an account in your wallet.',
  [WalletErrorCode.SIGN_REJECTED]: 'Review the transaction details and click Approve.',
  [WalletErrorCode.BROADCAST_FAILED]: 'Wait a moment and retry the transaction.',
  [WalletErrorCode.INSUFFICIENT_FUNDS]: 'Transfer funds to your account before proceeding.',
  [WalletErrorCode.NETWORK_ERROR]: 'Check your connection and try again.',
  [WalletErrorCode.SESSION_EXPIRED]: 'Click Connect Wallet to reconnect.',
  [WalletErrorCode.INVALID_CHAIN_ID]: 'Switch to VirtEngine network in your wallet.',
  [WalletErrorCode.UNKNOWN]: 'Try refreshing the page or contact support.',
};

/**
 * Set of error codes that can potentially be resolved by retrying.
 */
const RETRYABLE_ERROR_CODES = new Set<WalletErrorCode>([
  WalletErrorCode.WALLET_TIMEOUT,
  WalletErrorCode.BROADCAST_FAILED,
  WalletErrorCode.NETWORK_ERROR,
  WalletErrorCode.SESSION_EXPIRED,
]);

/**
 * Custom error class for wallet-related errors.
 *
 * Extends the standard Error class with additional metadata
 * useful for error handling and user feedback.
 *
 * @example
 * ```typescript
 * try {
 *   await wallet.connect();
 * } catch (error) {
 *   if (error instanceof WalletError) {
 *     console.log(error.code); // WalletErrorCode.WALLET_NOT_INSTALLED
 *     console.log(error.isRetryable); // false
 *     console.log(error.suggestedAction); // 'Install the wallet extension...'
 *   }
 * }
 * ```
 */
export class WalletError extends Error {
  /**
   * The specific error code identifying the type of error.
   */
  public readonly code: WalletErrorCode;

  /**
   * The original error that caused this wallet error, if any.
   */
  public readonly cause?: unknown;

  /**
   * Whether this error might be resolved by retrying the operation.
   */
  public readonly isRetryable: boolean;

  /**
   * A suggested action the user can take to resolve the error.
   */
  public readonly suggestedAction?: string;

  /**
   * Creates a new WalletError instance.
   *
   * @param code - The wallet error code
   * @param message - User-friendly error message (defaults to standard message for code)
   * @param options - Additional error options
   * @param options.cause - The original error that caused this error
   * @param options.isRetryable - Override default retryable status
   * @param options.suggestedAction - Override default suggested action
   */
  constructor(
    code: WalletErrorCode,
    message?: string,
    options?: {
      cause?: unknown;
      isRetryable?: boolean;
      suggestedAction?: string;
    }
  ) {
    super(message ?? WALLET_ERROR_MESSAGES[code]);

    this.name = 'WalletError';
    this.code = code;
    this.cause = options?.cause;
    this.isRetryable = options?.isRetryable ?? RETRYABLE_ERROR_CODES.has(code);
    this.suggestedAction = options?.suggestedAction ?? WALLET_ERROR_SUGGESTED_ACTIONS[code];

    // Maintains proper stack trace for where error was thrown (V8 only)
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, WalletError);
    }
  }

  /**
   * Returns a JSON representation of the error for logging/serialization.
   */
  toJSON(): {
    name: string;
    code: WalletErrorCode;
    message: string;
    isRetryable: boolean;
    suggestedAction?: string;
    cause?: string;
  } {
    return {
      name: this.name,
      code: this.code,
      message: this.message,
      isRetryable: this.isRetryable,
      suggestedAction: this.suggestedAction,
      cause: this.cause instanceof Error ? this.cause.message : undefined,
    };
  }
}

/**
 * Factory function to create a WalletError with appropriate defaults.
 *
 * @param code - The wallet error code
 * @param cause - Optional original error that caused this error
 * @returns A new WalletError instance
 *
 * @example
 * ```typescript
 * throw createWalletError(WalletErrorCode.WALLET_LOCKED);
 * // or with cause
 * throw createWalletError(WalletErrorCode.NETWORK_ERROR, originalError);
 * ```
 */
export function createWalletError(code: WalletErrorCode, cause?: unknown): WalletError {
  return new WalletError(code, undefined, { cause });
}

/**
 * Type guard to check if an error is a WalletError.
 *
 * @param error - The error to check
 * @returns True if the error is a WalletError instance
 *
 * @example
 * ```typescript
 * try {
 *   await wallet.sign(tx);
 * } catch (error) {
 *   if (isWalletError(error)) {
 *     // TypeScript knows error is WalletError here
 *     handleWalletError(error.code);
 *   }
 * }
 * ```
 */
export function isWalletError(error: unknown): error is WalletError {
  return error instanceof WalletError;
}

/**
 * Gets the user-friendly error message for a given error code.
 *
 * @param code - The wallet error code
 * @returns The user-friendly message for the error code
 *
 * @example
 * ```typescript
 * const message = getErrorMessage(WalletErrorCode.WALLET_LOCKED);
 * // Returns: 'Your wallet is locked. Please unlock your wallet and try again.'
 * ```
 */
export function getErrorMessage(code: WalletErrorCode): string {
  return WALLET_ERROR_MESSAGES[code];
}

/**
 * Gets the suggested action for a given error code.
 *
 * @param code - The wallet error code
 * @returns The suggested action for the error code
 *
 * @example
 * ```typescript
 * const action = getSuggestedAction(WalletErrorCode.INSUFFICIENT_FUNDS);
 * // Returns: 'Transfer funds to your account before proceeding.'
 * ```
 */
export function getSuggestedAction(code: WalletErrorCode): string {
  return WALLET_ERROR_SUGGESTED_ACTIONS[code];
}

/**
 * Known error message patterns from wallet extensions.
 * Used to identify specific error types from generic error messages.
 */
const ERROR_PATTERNS: Array<{ pattern: RegExp; code: WalletErrorCode }> = [
  // Keplr patterns
  { pattern: /not installed/i, code: WalletErrorCode.WALLET_NOT_INSTALLED },
  { pattern: /window\.keplr is undefined/i, code: WalletErrorCode.WALLET_NOT_INSTALLED },
  { pattern: /window\.leap is undefined/i, code: WalletErrorCode.WALLET_NOT_INSTALLED },
  { pattern: /rejected/i, code: WalletErrorCode.WALLET_CONNECTION_REJECTED },
  { pattern: /request rejected/i, code: WalletErrorCode.WALLET_CONNECTION_REJECTED },
  { pattern: /user denied/i, code: WalletErrorCode.WALLET_CONNECTION_REJECTED },
  { pattern: /user rejected/i, code: WalletErrorCode.SIGN_REJECTED },
  { pattern: /timeout/i, code: WalletErrorCode.WALLET_TIMEOUT },
  { pattern: /timed out/i, code: WalletErrorCode.WALLET_TIMEOUT },

  // Chain/Network patterns
  { pattern: /chain.*not.*supported/i, code: WalletErrorCode.CHAIN_NOT_SUPPORTED },
  { pattern: /unknown chain/i, code: WalletErrorCode.CHAIN_NOT_SUPPORTED },
  { pattern: /chain id.*mismatch/i, code: WalletErrorCode.INVALID_CHAIN_ID },
  { pattern: /invalid chain/i, code: WalletErrorCode.INVALID_CHAIN_ID },

  // Account patterns
  { pattern: /no account/i, code: WalletErrorCode.ACCOUNT_NOT_FOUND },
  { pattern: /account not found/i, code: WalletErrorCode.ACCOUNT_NOT_FOUND },
  { pattern: /locked/i, code: WalletErrorCode.WALLET_LOCKED },

  // Transaction patterns
  { pattern: /insufficient funds/i, code: WalletErrorCode.INSUFFICIENT_FUNDS },
  { pattern: /insufficient balance/i, code: WalletErrorCode.INSUFFICIENT_FUNDS },
  { pattern: /not enough/i, code: WalletErrorCode.INSUFFICIENT_FUNDS },
  { pattern: /broadcast.*fail/i, code: WalletErrorCode.BROADCAST_FAILED },
  { pattern: /tx.*fail/i, code: WalletErrorCode.BROADCAST_FAILED },
  { pattern: /transaction.*fail/i, code: WalletErrorCode.BROADCAST_FAILED },

  // Network patterns
  { pattern: /network error/i, code: WalletErrorCode.NETWORK_ERROR },
  { pattern: /fetch.*fail/i, code: WalletErrorCode.NETWORK_ERROR },
  { pattern: /failed to fetch/i, code: WalletErrorCode.NETWORK_ERROR },
  { pattern: /connection.*fail/i, code: WalletErrorCode.NETWORK_ERROR },
  { pattern: /ECONNREFUSED/i, code: WalletErrorCode.NETWORK_ERROR },
  { pattern: /ETIMEDOUT/i, code: WalletErrorCode.NETWORK_ERROR },

  // Session patterns
  { pattern: /session.*expired/i, code: WalletErrorCode.SESSION_EXPIRED },
  { pattern: /disconnected/i, code: WalletErrorCode.SESSION_EXPIRED },
];

/**
 * Parses an unknown error and converts it to a WalletError.
 *
 * This function analyzes error messages to determine the most appropriate
 * wallet error code. If the error is already a WalletError, it is returned
 * as-is. Unknown errors are wrapped with the UNKNOWN code.
 *
 * @param error - Any error to parse
 * @returns A WalletError with appropriate code and message
 *
 * @example
 * ```typescript
 * try {
 *   await someWalletOperation();
 * } catch (error) {
 *   const walletError = parseWalletError(error);
 *   // walletError is now a properly typed WalletError
 *   showErrorToUser(walletError.message);
 * }
 * ```
 */
export function parseWalletError(error: unknown): WalletError {
  // Already a WalletError, return as-is
  if (isWalletError(error)) {
    return error;
  }

  // Extract error message
  let message = '';
  if (error instanceof Error) {
    message = error.message;
  } else if (typeof error === 'string') {
    message = error;
  } else if (error && typeof error === 'object' && 'message' in error) {
    message = String((error as { message: unknown }).message);
  }

  // Try to match against known patterns
  for (const { pattern, code } of ERROR_PATTERNS) {
    if (pattern.test(message)) {
      return new WalletError(code, undefined, { cause: error });
    }
  }

  // Return unknown error with original message preserved in cause
  return new WalletError(WalletErrorCode.UNKNOWN, undefined, { cause: error });
}

/**
 * Checks if an error is retryable.
 *
 * Retryable errors are those that might succeed on a subsequent attempt,
 * such as network timeouts or temporary connectivity issues.
 *
 * @param error - The error to check
 * @returns True if the error can potentially be resolved by retrying
 *
 * @example
 * ```typescript
 * try {
 *   await broadcastTransaction(tx);
 * } catch (error) {
 *   if (isRetryableError(error)) {
 *     await delay(1000);
 *     await broadcastTransaction(tx); // Retry
 *   } else {
 *     throw error; // Don't retry
 *   }
 * }
 * ```
 */
export function isRetryableError(error: unknown): boolean {
  if (isWalletError(error)) {
    return error.isRetryable;
  }

  // Parse unknown errors and check retryability
  const walletError = parseWalletError(error);
  return walletError.isRetryable;
}

/**
 * Creates a timeout-wrapped promise for wallet operations.
 *
 * Useful for preventing indefinite hangs when wallet extensions
 * become unresponsive.
 *
 * @param promise - The promise to wrap with a timeout
 * @param timeoutMs - Timeout in milliseconds (default: 30000)
 * @param operationName - Name of the operation for error message
 * @returns The result of the promise or throws WalletError on timeout
 *
 * @example
 * ```typescript
 * const accounts = await withWalletTimeout(
 *   keplr.enable('virtengine-1'),
 *   10000,
 *   'connect'
 * );
 * ```
 */
export async function withWalletTimeout<T>(
  promise: Promise<T>,
  timeoutMs: number = 30000,
  operationName: string = 'operation'
): Promise<T> {
  let timeoutId: ReturnType<typeof setTimeout>;

  const timeoutPromise = new Promise<never>((_, reject) => {
    timeoutId = setTimeout(() => {
      reject(
        new WalletError(
          WalletErrorCode.WALLET_TIMEOUT,
          `Wallet ${operationName} timed out after ${timeoutMs}ms`
        )
      );
    }, timeoutMs);
  });

  try {
    const result = await Promise.race([promise, timeoutPromise]);
    clearTimeout(timeoutId!);
    return result;
  } catch (error) {
    clearTimeout(timeoutId!);
    throw error;
  }
}

/**
 * Wraps an async function to automatically convert errors to WalletErrors.
 *
 * @param fn - The async function to wrap
 * @returns A wrapped function that throws WalletError on failure
 *
 * @example
 * ```typescript
 * const safeConnect = wrapWithWalletError(async () => {
 *   return await keplr.enable('virtengine-1');
 * });
 *
 * try {
 *   await safeConnect();
 * } catch (error) {
 *   // error is always a WalletError
 * }
 * ```
 */
export function wrapWithWalletError<TArgs extends unknown[], TResult>(
  fn: (...args: TArgs) => Promise<TResult>
): (...args: TArgs) => Promise<TResult> {
  return async (...args: TArgs): Promise<TResult> => {
    try {
      return await fn(...args);
    } catch (error) {
      throw parseWalletError(error);
    }
  };
}
