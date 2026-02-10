/**
 * SDK-specific error types for VirtEngine TypeScript SDK
 */

/**
 * Base error class for VirtEngine SDK errors with structured error information
 */
export class VirtEngineSDKError extends Error {
  constructor(
    message: string,
    public readonly code: string = "UNKNOWN_ERROR",
    public readonly details?: Record<string, unknown>,
    options?: ErrorOptions,
  ) {
    super(message, options);
    this.name = "VirtEngineSDKError";
    Object.setPrototypeOf(this, new.target.prototype);
  }

  toJSON(): Record<string, unknown> {
    return {
      name: this.name,
      code: this.code,
      message: this.message,
      details: this.details,
    };
  }
}

/**
 * Transaction-related errors
 */
export class TxError extends VirtEngineSDKError {
  constructor(
    message: string,
    public readonly txHash?: string,
    public readonly rawLog?: string,
    details?: Record<string, unknown>,
    options?: ErrorOptions,
  ) {
    super(message, "TX_ERROR", { ...details, txHash, rawLog }, options);
    this.name = "TxError";
    Object.setPrototypeOf(this, TxError.prototype);
  }
}

/**
 * Error thrown when transaction broadcast fails
 */
export class TxBroadcastError extends TxError {
  constructor(message: string, txHash?: string, rawLog?: string, options?: ErrorOptions) {
    super(message, txHash, rawLog, undefined, options);
    this.name = "TxBroadcastError";
    Object.setPrototypeOf(this, TxBroadcastError.prototype);
  }
}

/**
 * Error thrown when transaction confirmation fails
 */
export class TxConfirmationError extends TxError {
  public readonly errorCode: number;

  constructor(
    message: string,
    txHash: string,
    errorCode: number,
    rawLog?: string,
    options?: ErrorOptions,
  ) {
    super(message, txHash, rawLog, { errorCode }, options);
    this.name = "TxConfirmationError";
    this.errorCode = errorCode;
    Object.setPrototypeOf(this, TxConfirmationError.prototype);
  }
}

/**
 * Legacy alias for TxError (maintained for backwards compatibility)
 */
export class TransactionError extends VirtEngineSDKError {
  readonly errorCode: number;
  readonly txHash?: string;

  constructor(message: string, code: number, txHash?: string, options?: ErrorOptions) {
    super(message, "TX_ERROR", { errorCode: code, txHash }, options);
    this.name = "TransactionError";
    this.errorCode = code;
    this.txHash = txHash;
    Object.setPrototypeOf(this, TransactionError.prototype);
  }
}

/**
 * Query-related errors
 */
export class QueryError extends VirtEngineSDKError {
  constructor(
    message: string,
    public readonly method: string,
    public readonly grpcCode?: number,
    details?: Record<string, unknown>,
    options?: ErrorOptions,
  ) {
    super(`Query failed in ${method}: ${message}`, "QUERY_ERROR", { ...details, method, grpcCode }, options);
    this.name = "QueryError";
    Object.setPrototypeOf(this, QueryError.prototype);
  }
}

/**
 * Error thrown when a resource is not found
 */
export class NotFoundError extends QueryError {
  constructor(resource: string, id: string, options?: ErrorOptions) {
    super(`${resource} not found: ${id}`, "get", 5, { resource, id }, options);
    this.name = "NotFoundError";
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
}

/**
 * Wallet-related errors
 */
export class WalletError extends VirtEngineSDKError {
  constructor(message: string, public readonly walletType?: string, options?: ErrorOptions) {
    super(message, "WALLET_ERROR", { walletType }, options);
    this.name = "WalletError";
    Object.setPrototypeOf(this, WalletError.prototype);
  }
}

/**
 * Error thrown when wallet is not connected
 */
export class WalletNotConnectedError extends WalletError {
  constructor(options?: ErrorOptions) {
    super("Wallet is not connected", undefined, options);
    this.name = "WalletNotConnectedError";
    Object.setPrototypeOf(this, WalletNotConnectedError.prototype);
  }
}

/**
 * Error thrown when wallet is not available
 */
export class WalletNotAvailableError extends WalletError {
  constructor(walletType: string, options?: ErrorOptions) {
    super(`${walletType} wallet is not available`, walletType, options);
    this.name = "WalletNotAvailableError";
    Object.setPrototypeOf(this, WalletNotAvailableError.prototype);
  }
}

/**
 * Error thrown when validation fails
 */
export class SDKValidationError extends VirtEngineSDKError {
  readonly field?: string;
  readonly constraint?: string;

  constructor(message: string, field?: string, constraint?: string, options?: ErrorOptions) {
    super(field ? `Validation failed for ${field}: ${message}` : message, "VALIDATION_ERROR", { field, constraint }, options);
    this.name = "SDKValidationError";
    this.field = field;
    this.constraint = constraint;
    Object.setPrototypeOf(this, SDKValidationError.prototype);
  }
}

/**
 * Network-related errors
 */
export class NetworkError extends VirtEngineSDKError {
  constructor(
    message: string,
    public readonly statusCode?: number,
    public readonly url?: string,
    options?: ErrorOptions,
  ) {
    super(message, "NETWORK_ERROR", { statusCode, url }, options);
    this.name = "NetworkError";
    Object.setPrototypeOf(this, NetworkError.prototype);
  }
}

/**
 * Error thrown when an operation times out
 */
export class TimeoutError extends NetworkError {
  constructor(operation: string, timeoutMs: number, options?: ErrorOptions) {
    super(`Operation timed out after ${timeoutMs}ms: ${operation}`, undefined, undefined, options);
    this.name = "TimeoutError";
    Object.setPrototypeOf(this, TimeoutError.prototype);
  }
}

/**
 * Error thrown when a module is not yet implemented
 */
export class NotImplementedError extends VirtEngineSDKError {
  readonly module: string;

  constructor(module: string, message = "Proto generation needed", options?: ErrorOptions) {
    super(`${module} module not yet generated - ${message}`, "NOT_IMPLEMENTED", { module }, options);
    this.name = "NotImplementedError";
    this.module = module;
    Object.setPrototypeOf(this, NotImplementedError.prototype);
  }
}

/**
 * Type guard to check if an error is a VirtEngineSDKError
 */
export function isVirtEngineError(error: unknown): error is VirtEngineSDKError {
  return error instanceof VirtEngineSDKError;
}

/**
 * Type guard to check if an error is a TxError
 */
export function isTxError(error: unknown): error is TxError {
  return error instanceof TxError;
}

/**
 * Type guard to check if an error is a QueryError
 */
export function isQueryError(error: unknown): error is QueryError {
  return error instanceof QueryError;
}

/**
 * Type guard to check if an error is a WalletError
 */
export function isWalletError(error: unknown): error is WalletError {
  return error instanceof WalletError;
}

/**
 * Type guard to check if an error is a SDKValidationError
 */
export function isValidationError(error: unknown): error is SDKValidationError {
  return error instanceof SDKValidationError;
}

/**
 * Type guard to check if an error is a NetworkError
 */
export function isNetworkError(error: unknown): error is NetworkError {
  return error instanceof NetworkError;
}

/**
 * Wraps an unknown error in a VirtEngineSDKError with context
 */
export function wrapError(error: unknown, context: string): VirtEngineSDKError {
  if (error instanceof VirtEngineSDKError) {
    return error;
  }

  const message = error instanceof Error ? error.message : String(error);
  return new VirtEngineSDKError(`${context}: ${message}`, "UNKNOWN_ERROR");
}
