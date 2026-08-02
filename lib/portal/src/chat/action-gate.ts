export interface GatedAction {
  actionId: string;
  toolName: string;
  requiredCapability: string;
  payload: unknown;
}

export interface PolicyDecision {
  actionId: string;
  toolName: string;
  allowed: boolean;
  expiresAt: number;
}

export interface CapabilityDecision {
  actionId: string;
  toolName: string;
  capability: string;
  allowed: boolean;
  expiresAt: number;
}

export interface AuthoritativeSimulation {
  actionId: string;
  toolName: string;
  stateDigest: string;
  impact: unknown;
}

export interface ActionPreview {
  actionId: string;
  toolName: string;
  requiredCapability: string;
  canonicalAction: string;
  policyDecision: PolicyDecision;
  capabilityDecision: CapabilityDecision;
  stateDigest: string;
  impact: unknown;
  nonce: string;
  issuedAt: number;
  expiresAt: number;
  previewDigest: string;
}

export interface PreviewConfirmation {
  confirmed: true;
  previewDigest: string;
  nonce: string;
  confirmedAt: number;
  expiresAt: number;
}

export interface SignerEvidence {
  actionId: string;
  toolName: string;
  previewDigest: string;
  stateDigest: string;
  walletAddress: string;
  accountId: string;
  chainId: string;
  signature: string;
  expiresAt: number;
  mfa?: {
    scope: string;
    expiresAt: number;
  };
}

export interface ExecutionContext {
  walletAddress: string;
  accountId: string;
  chainId: string;
  mfaScope: string;
}

export interface ExecuteActionRequest {
  preview: ActionPreview;
  confirmation?: PreviewConfirmation;
  signer?: SignerEvidence;
  context: ExecutionContext;
}

export interface FinalActionAuthorization {
  action: GatedAction;
  previewDigest: string;
  stateDigest: string;
  nonce: string;
  signer: SignerEvidence;
}

export type ActionExecutor<Result> = (
  authorization: FinalActionAuthorization,
) => Promise<Result>;

export type GateDenialReason =
  | "policy_missing"
  | "policy_denied"
  | "policy_mismatch"
  | "policy_expired"
  | "capability_missing"
  | "capability_denied"
  | "capability_mismatch"
  | "capability_expired"
  | "simulation_failed"
  | "simulation_mismatch"
  | "simulation_invalid"
  | "preview_unknown"
  | "preview_digest_mismatch"
  | "preview_expired"
  | "confirmation_missing"
  | "confirmation_mismatch"
  | "confirmation_expired"
  | "signer_missing"
  | "signer_mismatch"
  | "signer_expired"
  | "signature_missing"
  | "mfa_missing"
  | "mfa_scope_mismatch"
  | "mfa_expired"
  | "nonce_in_use"
  | "nonce_replayed"
  | "authoritative_state_unavailable"
  | "state_drift"
  | "execution_failed";

export interface GateDenied {
  status: "denied";
  reason: GateDenialReason;
}

export interface GatePreviewReady {
  status: "preview";
  preview: ActionPreview;
}

export interface GateExecuted<Result> {
  status: "executed";
  result: Result;
}

export type PrepareActionResult = GateDenied | GatePreviewReady;
export type ExecuteActionResult<Result> = GateDenied | GateExecuted<Result>;

export interface ActionGateAuthorities {
  decidePolicy: (action: GatedAction) => Promise<PolicyDecision | undefined>;
  decideCapability: (
    action: GatedAction,
  ) => Promise<CapabilityDecision | undefined>;
  simulate: (action: GatedAction) => Promise<AuthoritativeSimulation>;
  readCurrentStateDigest: (action: GatedAction) => Promise<string>;
}

export interface ActionGateOptions {
  now?: () => number;
  previewLifetimeMs?: number;
  maxConfirmationLifetimeMs?: number;
}

interface PreviewRecord {
  action: GatedAction;
  previewDigest: string;
  state: "available" | "active" | "consumed";
}

const DEFAULT_PREVIEW_LIFETIME_MS = 2 * 60 * 1000;
const DEFAULT_CONFIRMATION_LIFETIME_MS = 60 * 1000;
const textEncoder = new TextEncoder();

const denied = (reason: GateDenialReason): GateDenied => ({
  status: "denied",
  reason,
});

const canonicalize = (value: unknown, seen = new Set<object>()): string => {
  if (
    value === null ||
    typeof value === "string" ||
    typeof value === "boolean"
  ) {
    return JSON.stringify(value);
  }
  if (typeof value === "number") {
    if (!Number.isFinite(value)) {
      throw new TypeError("Canonical values must contain finite numbers");
    }
    return JSON.stringify(value);
  }
  if (typeof value !== "object") {
    throw new TypeError("Canonical values must be JSON-compatible");
  }
  if (seen.has(value)) {
    throw new TypeError("Canonical values must not be cyclic");
  }

  seen.add(value);
  try {
    if (Array.isArray(value)) {
      return `[${value.map((item) => canonicalize(item, seen)).join(",")}]`;
    }

    const prototype = Object.getPrototypeOf(value);
    if (prototype !== Object.prototype && prototype !== null) {
      throw new TypeError("Canonical values must use plain objects");
    }
    const record = value as Record<string, unknown>;
    return `{${Object.keys(record)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${canonicalize(record[key], seen)}`)
      .join(",")}}`;
  } finally {
    seen.delete(value);
  }
};

const toHex = (bytes: Uint8Array): string =>
  Array.from(bytes)
    .map((byte) => byte.toString(16).padStart(2, "0"))
    .join("");

const sha256Hex = async (payload: string): Promise<string> => {
  const bytes = textEncoder.encode(payload);
  if (globalThis.crypto?.subtle) {
    const digest = await globalThis.crypto.subtle.digest("SHA-256", bytes);
    return toHex(new Uint8Array(digest));
  }

  const nodeCrypto = await import("crypto");
  return nodeCrypto.createHash("sha256").update(bytes).digest("hex");
};

const randomNonce = async (): Promise<string> => {
  const bytes = new Uint8Array(16);
  if (globalThis.crypto?.getRandomValues) {
    globalThis.crypto.getRandomValues(bytes);
    return toHex(bytes);
  }

  const nodeCrypto = await import("crypto");
  return nodeCrypto.randomBytes(16).toString("hex");
};

const cloneCanonical = <Value>(value: Value): Value =>
  JSON.parse(canonicalize(value)) as Value;

export class ActionGate {
  private readonly previews = new Map<string, PreviewRecord>();
  private readonly now: () => number;
  private readonly previewLifetimeMs: number;
  private readonly maxConfirmationLifetimeMs: number;

  constructor(
    private readonly authorities: ActionGateAuthorities,
    options: ActionGateOptions = {},
  ) {
    this.now = options.now ?? Date.now;
    this.previewLifetimeMs = Math.max(
      1,
      Math.min(
        options.previewLifetimeMs ?? DEFAULT_PREVIEW_LIFETIME_MS,
        DEFAULT_PREVIEW_LIFETIME_MS,
      ),
    );
    this.maxConfirmationLifetimeMs = Math.max(
      1,
      Math.min(
        options.maxConfirmationLifetimeMs ?? DEFAULT_CONFIRMATION_LIFETIME_MS,
        DEFAULT_CONFIRMATION_LIFETIME_MS,
      ),
    );
  }

  async prepare(action: GatedAction): Promise<PrepareActionResult> {
    const now = this.now();
    const policy = await this.authorities.decidePolicy(action);
    if (!policy) return denied("policy_missing");
    if (
      policy.actionId !== action.actionId ||
      policy.toolName !== action.toolName
    ) {
      return denied("policy_mismatch");
    }
    if (!policy.allowed) return denied("policy_denied");
    if (policy.expiresAt <= now) return denied("policy_expired");

    const capability = await this.authorities.decideCapability(action);
    if (!capability) return denied("capability_missing");
    if (
      capability.actionId !== action.actionId ||
      capability.toolName !== action.toolName ||
      capability.capability !== action.requiredCapability
    ) {
      return denied("capability_mismatch");
    }
    if (!capability.allowed) return denied("capability_denied");
    if (capability.expiresAt <= now) return denied("capability_expired");

    let simulation: AuthoritativeSimulation;
    try {
      simulation = await this.authorities.simulate(action);
    } catch {
      return denied("simulation_failed");
    }
    if (
      simulation.actionId !== action.actionId ||
      simulation.toolName !== action.toolName
    ) {
      return denied("simulation_mismatch");
    }
    if (!simulation.stateDigest) return denied("simulation_invalid");

    let canonicalAction: string;
    let impact: unknown;
    try {
      canonicalAction = canonicalize(action);
      impact = cloneCanonical(simulation.impact);
    } catch {
      return denied("simulation_invalid");
    }

    let nonce = await randomNonce();
    while (this.previews.has(nonce)) nonce = await randomNonce();

    const previewWithoutDigest = {
      actionId: action.actionId,
      toolName: action.toolName,
      requiredCapability: action.requiredCapability,
      canonicalAction,
      policyDecision: cloneCanonical(policy),
      capabilityDecision: cloneCanonical(capability),
      stateDigest: simulation.stateDigest,
      impact,
      nonce,
      issuedAt: now,
      expiresAt: Math.min(
        now + this.previewLifetimeMs,
        policy.expiresAt,
        capability.expiresAt,
      ),
    };
    const preview: ActionPreview = {
      ...previewWithoutDigest,
      previewDigest: await sha256Hex(canonicalize(previewWithoutDigest)),
    };
    this.previews.set(nonce, {
      action: cloneCanonical(action),
      previewDigest: preview.previewDigest,
      state: "available",
    });
    return { status: "preview", preview };
  }

  async execute<Result>(
    request: ExecuteActionRequest,
    executor: ActionExecutor<Result>,
  ): Promise<ExecuteActionResult<Result>> {
    const { preview, confirmation, signer, context } = request;
    const record = this.previews.get(preview.nonce);
    if (!record) return denied("preview_unknown");

    const { previewDigest, ...previewWithoutDigest } = preview;
    let computedDigest: string;
    try {
      computedDigest = await sha256Hex(canonicalize(previewWithoutDigest));
    } catch {
      return denied("preview_digest_mismatch");
    }
    if (
      computedDigest !== previewDigest ||
      record.previewDigest !== previewDigest
    ) {
      return denied("preview_digest_mismatch");
    }

    const now = this.now();
    if (preview.expiresAt <= now) return denied("preview_expired");
    if (!confirmation) return denied("confirmation_missing");
    if (
      confirmation.confirmed !== true ||
      confirmation.previewDigest !== previewDigest ||
      confirmation.nonce !== preview.nonce ||
      confirmation.confirmedAt > now
    ) {
      return denied("confirmation_mismatch");
    }
    if (
      confirmation.expiresAt <= now ||
      confirmation.expiresAt > preview.expiresAt ||
      confirmation.expiresAt - confirmation.confirmedAt >
        this.maxConfirmationLifetimeMs
    ) {
      return denied("confirmation_expired");
    }

    if (!signer) return denied("signer_missing");
    if (
      signer.actionId !== preview.actionId ||
      signer.toolName !== preview.toolName ||
      signer.previewDigest !== previewDigest ||
      signer.stateDigest !== preview.stateDigest ||
      signer.walletAddress !== context.walletAddress ||
      signer.accountId !== context.accountId ||
      signer.chainId !== context.chainId
    ) {
      return denied("signer_mismatch");
    }
    if (signer.expiresAt <= now || signer.expiresAt > preview.expiresAt) {
      return denied("signer_expired");
    }
    if (!signer.signature.trim()) return denied("signature_missing");
    if (!signer.mfa) return denied("mfa_missing");
    if (
      context.mfaScope !== record.action.requiredCapability ||
      signer.mfa.scope !== record.action.requiredCapability
    ) {
      return denied("mfa_scope_mismatch");
    }
    if (
      signer.mfa.expiresAt <= now ||
      signer.mfa.expiresAt > preview.expiresAt
    ) {
      return denied("mfa_expired");
    }

    if (record.state === "active") return denied("nonce_in_use");
    if (record.state === "consumed") return denied("nonce_replayed");
    record.state = "active";

    let currentStateDigest: string;
    try {
      currentStateDigest = await this.authorities.readCurrentStateDigest(
        record.action,
      );
    } catch {
      record.state = "consumed";
      return denied("authoritative_state_unavailable");
    }
    if (currentStateDigest !== preview.stateDigest) {
      record.state = "consumed";
      return denied("state_drift");
    }

    try {
      const result = await executor({
        action: record.action,
        previewDigest,
        stateDigest: currentStateDigest,
        nonce: preview.nonce,
        signer,
      });
      record.state = "consumed";
      return { status: "executed", result };
    } catch {
      record.state = "consumed";
      return denied("execution_failed");
    }
  }
}
