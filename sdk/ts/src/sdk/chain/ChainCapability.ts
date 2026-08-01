import { TransportError } from "../transport/TransportError.ts";
import type { TxClient } from "../transport/tx/TxClient.ts";

export const ChainCapability = {
  Disconnected: "disconnected",
  QueryOnly: "query-only",
  SigningReady: "signing-ready",
  MfaAuthorized: "MFA-authorized",
} as const;

export type ChainCapabilityState = typeof ChainCapability[keyof typeof ChainCapability];

export interface MFAAuthorizationMetadata {
  authorizationId: string;
  expiresAt: number;
}

export const ChainCapabilityErrorReason = {
  Disconnected: "disconnected",
  SigningRequired: "signing-required",
  MfaRequired: "MFA-required",
  InvalidMfaAuthorization: "invalid-MFA-authorization",
  InvalidTransition: "invalid-transition",
} as const;

export type ChainCapabilityErrorReason = typeof ChainCapabilityErrorReason[keyof typeof ChainCapabilityErrorReason];

export class ChainCapabilityError extends TransportError {
  constructor(
    public readonly reason: ChainCapabilityErrorReason,
    public readonly currentCapability: ChainCapabilityState,
    public readonly requiredCapability: ChainCapabilityState,
    public readonly operation: string,
  ) {
    super(
      `Chain SDK capability '${currentCapability}' cannot perform '${operation}'; '${requiredCapability}' is required`,
      TransportError.Code.FailedPrecondition,
    );
    this.name = "ChainCapabilityError";
    Object.setPrototypeOf(this, new.target.prototype);
  }
}

export class ChainCapabilityController {
  #state: ChainCapabilityState;
  #signer?: TxClient;
  #mfaAuthorization?: MFAAuthorizationMetadata;

  constructor(signer?: TxClient) {
    this.#signer = signer;
    this.#state = signer ? ChainCapability.SigningReady : ChainCapability.QueryOnly;
  }

  get state(): ChainCapabilityState {
    return this.#state;
  }

  connect(): void {
    if (this.#state === ChainCapability.Disconnected) {
      this.#state = ChainCapability.QueryOnly;
    }
  }

  disconnect(): void {
    this.#signer = undefined;
    this.#mfaAuthorization = undefined;
    this.#state = ChainCapability.Disconnected;
  }

  setSigner(signer: TxClient): void {
    if (this.#state === ChainCapability.Disconnected) {
      throw this.#invalidTransition(ChainCapability.SigningReady, "setSigner");
    }
    this.#signer = signer;
    this.#mfaAuthorization = undefined;
    this.#state = ChainCapability.SigningReady;
  }

  clearSigner(): void {
    this.#signer = undefined;
    this.#mfaAuthorization = undefined;
    if (this.#state !== ChainCapability.Disconnected) {
      this.#state = ChainCapability.QueryOnly;
    }
  }

  authorizeMFA(metadata: MFAAuthorizationMetadata): void {
    if (this.#state !== ChainCapability.SigningReady) {
      throw this.#invalidTransition(ChainCapability.SigningReady, "authorizeMFA");
    }
    if (!isValidMFAAuthorization(metadata)) {
      throw new ChainCapabilityError(
        ChainCapabilityErrorReason.InvalidMfaAuthorization,
        this.#state,
        ChainCapability.MfaAuthorized,
        "authorizeMFA",
      );
    }
    this.#mfaAuthorization = { ...metadata };
    this.#state = ChainCapability.MfaAuthorized;
  }

  revokeMFA(): void {
    this.#mfaAuthorization = undefined;
    if (this.#state === ChainCapability.MfaAuthorized) {
      this.#state = ChainCapability.SigningReady;
    }
  }

  assertCanQuery(operation: string): void {
    if (this.#state === ChainCapability.Disconnected) {
      throw new ChainCapabilityError(
        ChainCapabilityErrorReason.Disconnected,
        this.#state,
        ChainCapability.QueryOnly,
        operation,
      );
    }
  }

  requireSigner(operation: string): TxClient {
    if (this.#state === ChainCapability.Disconnected) {
      throw new ChainCapabilityError(
        ChainCapabilityErrorReason.Disconnected,
        this.#state,
        ChainCapability.SigningReady,
        operation,
      );
    }
    if (!this.#signer) {
      throw new ChainCapabilityError(
        ChainCapabilityErrorReason.SigningRequired,
        this.#state,
        ChainCapability.SigningReady,
        operation,
      );
    }
    return this.#signer;
  }

  assertMFAAuthorized(operation: string): void {
    if (this.#state === ChainCapability.MfaAuthorized && !isValidMFAAuthorization(this.#mfaAuthorization)) {
      this.#mfaAuthorization = undefined;
      this.#state = this.#signer ? ChainCapability.SigningReady : ChainCapability.QueryOnly;
    }
    if (this.#state !== ChainCapability.MfaAuthorized) {
      throw new ChainCapabilityError(
        ChainCapabilityErrorReason.MfaRequired,
        this.#state,
        ChainCapability.MfaAuthorized,
        operation,
      );
    }
  }

  #invalidTransition(requiredCapability: ChainCapabilityState, operation: string): ChainCapabilityError {
    return new ChainCapabilityError(
      ChainCapabilityErrorReason.InvalidTransition,
      this.#state,
      requiredCapability,
      operation,
    );
  }
}

function isValidMFAAuthorization(metadata: unknown): metadata is MFAAuthorizationMetadata {
  if (!metadata || typeof metadata !== "object") return false;
  const authorization = metadata as Partial<MFAAuthorizationMetadata>;
  return typeof authorization.authorizationId === "string"
    && authorization.authorizationId.trim().length > 0
    && typeof authorization.expiresAt === "number"
    && Number.isFinite(authorization.expiresAt)
    && authorization.expiresAt > Date.now();
}
