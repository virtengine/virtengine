import type { LivenessCheckResult, LivenessProvider } from "../types/capture";

export interface LivenessValidationContext {
  sessionId: string;
  challengeId: string;
  timeoutMs: number;
}

export function validateLivenessEvidence(
  evidence: LivenessCheckResult | null | undefined,
  context: LivenessValidationContext,
): LivenessCheckResult {
  if (!evidence) {
    throw new Error("Liveness evidence is missing");
  }
  if (!evidence.passed) {
    throw new Error("Liveness verification failed");
  }
  if (
    evidence.sessionId !== context.sessionId ||
    evidence.challengeId !== context.challengeId
  ) {
    throw new Error("Liveness evidence is not bound to this capture");
  }
  if (!evidence.providerId.trim() || !evidence.providerVersion.trim()) {
    throw new Error("Liveness provider identity is missing");
  }
  if (!evidence.evidenceDigest.trim()) {
    throw new Error("Liveness evidence digest is missing");
  }
  if (
    !Number.isFinite(evidence.score) ||
    evidence.score < 0 ||
    evidence.score > 1
  ) {
    throw new Error("Liveness score is invalid");
  }
  if (
    !Number.isFinite(evidence.challengeDurationMs) ||
    evidence.challengeDurationMs <= 0 ||
    evidence.challengeDurationMs > context.timeoutMs
  ) {
    throw new Error("Liveness evidence duration is invalid");
  }
  return evidence;
}

export async function requestLivenessEvidence(
  provider: LivenessProvider | undefined,
  context: LivenessValidationContext,
): Promise<LivenessCheckResult> {
  if (!provider) {
    throw new Error("Liveness provider is unavailable");
  }

  let timeout: ReturnType<typeof setTimeout> | undefined;
  try {
    const evidence = await Promise.race([
      provider.verify(context),
      new Promise<never>((_, reject) => {
        timeout = setTimeout(
          () => reject(new Error("Liveness verification timed out")),
          context.timeoutMs,
        );
      }),
    ]);
    return validateLivenessEvidence(evidence, context);
  } finally {
    if (timeout) clearTimeout(timeout);
  }
}
