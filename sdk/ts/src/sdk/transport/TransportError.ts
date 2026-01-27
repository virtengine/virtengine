import { Code, ConnectError } from "@connectrpc/connect";

export enum TransportErrorCode {
  /**
   * Canceled, usually be the user
   */
  Canceled = 1,
  /**
   * Unknown error
   */
  Unknown = 2,
  /**
   * Argument invalid regardless of system state
   */
  InvalidArgument = 3,
  /**
   * Operation timed out.
   */
  DeadlineExceeded = 4,
  /**
   * Entity not found.
   */
  NotFound = 5,
  /**
   * Entity already exists.
   */
  AlreadyExists = 6,
  /**
   * Operation not authorized.
   */
  PermissionDenied = 7,
  /**
   * Quota exhausted.
   */
  ResourceExhausted = 8,
  /**
   * Argument invalid in current system state.
   */
  FailedPrecondition = 9,
  /**
   * Operation aborted.
   */
  Aborted = 10,
  /**
   * Out of bounds, use instead of FailedPrecondition.
   */
  OutOfRange = 11,
  /**
   * Operation not implemented or disabled.
   */
  Unimplemented = 12,
  /**
   * Internal error, reserved for "serious errors".
   */
  Internal = 13,
  /**
   * Unavailable, client should back off and retry.
   */
  Unavailable = 14,
  /**
   * Unrecoverable data loss or corruption.
   */
  DataLoss = 15,
  /**
   * Request isn't authenticated.
   */
  Unauthenticated = 16,
}

export class TransportError extends Error {
  static Code = TransportErrorCode;

  public readonly code: typeof TransportErrorCode[keyof typeof TransportErrorCode];
  public readonly metadata: Headers;
  public readonly cause?: unknown;

  /**
   * Convert any value - typically a caught error into a TransportError,
   * following these rules:
   * - If the value is already a TransportError, return it as is.
   * - If the value is an AbortError from the fetch API, return the message
   *   of the AbortError with code Canceled.
   * - For other Errors, return the error message with code Unknown by default.
   * - For other values, return the values String representation as a message,
   *   with the code Unknown by default.
   * The original value will be used for the "cause" property for the new
   * TransportError.
   */
  static from(cause: unknown, code = TransportError.Code.Unknown) {
    if (cause instanceof this) return cause;
    if (cause instanceof ConnectError) {
      const key = Code[cause.code];
      const code = Object.hasOwn(TransportErrorCode, key) ? TransportErrorCode[key as keyof typeof TransportErrorCode] : TransportErrorCode.Unknown;
      return new TransportError(cause.message, code, { cause, metadata: cause.metadata });
    }
    if (cause instanceof Error) {
      if (cause.name == "AbortError") {
        return new TransportError(cause.message, TransportErrorCode.Canceled);
      }
      return new TransportError(cause.message, code, { cause });
    }
    return new TransportError(String(cause), code, { cause });
  }

  /**
   * Create a new TransportError.
   * Outgoing details are only relevant for the server side - a service may
   * raise an error with details, and it is up to the protocol implementation
   * to encode and send the details along with error.
   */
  constructor(message: string, code = TransportError.Code.Unknown, options?: {
    metadata?: HeadersInit;
    cause?: unknown;
  }) {
    super(`[${stringifyCode(code)}] ${message}`, { cause: options?.cause });
    this.name = "TransportError";
    Object.setPrototypeOf(this, new.target.prototype);
    this.code = code;
    this.metadata = new Headers(options?.metadata ?? {});
    this.cause ??= options?.cause;
  }
}

function stringifyCode(value: TransportErrorCode) {
  const name = TransportErrorCode[value];
  if (typeof name !== "string") return value.toString();
  return (name[0].toLowerCase() + name.slice(1).replace(/[A-Z]/g, (char) => `_${char.toLowerCase()}`));
}
