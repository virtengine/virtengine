/**
 * VEID on-chain submission pipeline.
 */

import type { CaptureResult } from '../../types/capture';
import { verifySalt } from '../../utils/salt-generator';
import { computeHash } from '../../utils/signature';
import {
  createUploadMetadata,
  createUploadNonce,
  createUploadSignatures,
} from './envelope';
import { encryptPayloadForRecipients } from './encryption';
import { fetchValidatorEncryptionKeys } from './validator-keys';
import {
  buildUploadScopeMessage,
  createScopeId,
  normalizeScopeType,
} from './transaction';
import type {
  ApprovedClientCheckOptions,
  EncryptedPayloadEnvelope,
  SubmissionRequest,
  SubmissionResult,
  SubmissionStatus,
  SubmissionUpdate,
  ValidatorRecipientKey,
} from './types';
import {
  BroadcastError,
  CaptureValidationError,
  EncryptionError,
  SigningError,
  SubmissionError,
  SubmissionTimeoutError,
} from './types';

const DEFAULT_SCORE_POLL_TIMEOUT = 5 * 60 * 1000;
const DEFAULT_SCORE_POLL_INTERVAL = 5000;

function updateStatus(
  onStatus: SubmissionRequest['onStatus'],
  status: SubmissionStatus,
  update: Omit<SubmissionUpdate, 'status'> = {}
) {
  if (!onStatus) return;
  onStatus({ status, ...update });
}

function resolveCaptureTimestamp(capture: CaptureResult): number {
  const capturedAt = capture.metadata?.capturedAt;
  if (!capturedAt) return Math.floor(Date.now() / 1000);
  const parsed = Date.parse(capturedAt);
  if (Number.isNaN(parsed)) return Math.floor(Date.now() / 1000);
  return Math.floor(parsed / 1000);
}

async function buildCapturePayload(capture: CaptureResult): Promise<Uint8Array> {
  const buffer = await capture.imageBlob.arrayBuffer();
  const bytes = new Uint8Array(buffer);
  if (!bytes.length) {
    throw new CaptureValidationError('Capture payload is empty');
  }
  return bytes;
}

function validateCapture(capture: CaptureResult) {
  if (!capture) {
    throw new CaptureValidationError('Capture result is required');
  }
  if (!capture.imageBlob) {
    throw new CaptureValidationError('Capture image blob is missing');
  }
  if (!capture.salt || !verifySalt(capture.salt)) {
    throw new CaptureValidationError('Capture salt is invalid');
  }
  if (!capture.metadata?.deviceFingerprint) {
    throw new CaptureValidationError('Device fingerprint is required');
  }
  if (!capture.metadata?.capturedAt) {
    throw new CaptureValidationError('Capture timestamp is required');
  }
}

async function getValidatorKeys(request: SubmissionRequest): Promise<ValidatorRecipientKey[]> {
  if (request.validatorKeys && request.validatorKeys.length > 0) {
    return request.validatorKeys;
  }
  return fetchValidatorEncryptionKeys({
    restEndpoint: request.restEndpoint,
    fetchJson: request.fetchJson,
  });
}

function buildEnvelopeMetadata(capture: CaptureResult, scopeType: number): Record<string, string> {
  const metadata: Record<string, string> = {
    capture_type: String(scopeType),
    mime_type: capture.mimeType ?? '',
    document_type: capture.metadata?.documentType ?? '',
    document_side: capture.metadata?.documentSide ?? '',
    capture_timestamp: capture.metadata?.capturedAt ?? '',
    device_fingerprint: capture.metadata?.deviceFingerprint ?? '',
    session_id: capture.metadata?.sessionId ?? '',
    client_version: capture.metadata?.clientVersion ?? '',
    quality_score: capture.metadata?.qualityScore?.toString() ?? '',
    width: capture.dimensions?.width?.toString() ?? '',
    height: capture.dimensions?.height?.toString() ?? '',
  };

  Object.keys(metadata).forEach((key) => {
    if (!metadata[key]) delete metadata[key];
  });

  return metadata;
}

function normalizeApprovedClientPayload(payload: unknown): Array<Record<string, unknown>> {
  if (!payload || typeof payload !== 'object') return [];
  const record = payload as Record<string, unknown>;
  const candidates = [record.clients, record.approved_clients, record.approvedClients, record.items];
  for (const candidate of candidates) {
    if (Array.isArray(candidate)) return candidate as Array<Record<string, unknown>>;
  }
  if (record.result && typeof record.result === 'object') {
    const result = record.result as Record<string, unknown>;
    if (Array.isArray(result.clients)) return result.clients as Array<Record<string, unknown>>;
  }
  if (record.data && typeof record.data === 'object') {
    const data = record.data as Record<string, unknown>;
    if (Array.isArray(data.clients)) return data.clients as Array<Record<string, unknown>>;
  }
  return [];
}

async function ensureApprovedClient(
  clientId: string,
  restEndpoint: string,
  fetchJson: SubmissionRequest['fetchJson'],
  options?: ApprovedClientCheckOptions
) {
  if (!options?.enabled) return;
  if (options.allowedClientIds?.length) {
    if (!options.allowedClientIds.includes(clientId)) {
      throw new SigningError('Client is not in the approved allowlist', { clientId });
    }
    return;
  }

  const request = fetchJson
    ? (path: string, params?: Record<string, string | number | boolean | undefined>) =>
        fetchJson(path, params)
    : async (path: string, params?: Record<string, string | number | boolean | undefined>) => {
        const url = new URL(path, restEndpoint);
        if (params) {
          Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
              url.searchParams.set(key, String(value));
            }
          });
        }
        const response = await fetch(url.toString(), { headers: { Accept: 'application/json' } });
        if (!response.ok) {
          throw new Error(`Approved client query failed: ${response.status}`);
        }
        return response.json();
      };

  const payload = await request('/virtengine/veid/v1/approved_clients');
  const clients = normalizeApprovedClientPayload(payload);
  const matched = clients.find((client) => {
    const record = client ?? {};
    const id =
      (record.client_id as string | undefined) ??
      (record.clientId as string | undefined) ??
      (record.id as string | undefined);
    return id === clientId;
  });
  if (!matched) {
    throw new SigningError('Client is not in the approved client list', { clientId });
  }
  const active =
    matched.active === undefined ? true : Boolean(matched.active ?? matched.is_active ?? matched.isActive);
  if (!active) {
    throw new SigningError('Approved client is inactive', { clientId });
  }
}

async function pollForScore(options: {
  restEndpoint: string;
  address: string;
  timeoutMs: number;
  intervalMs: number;
  fetchJson?: SubmissionRequest['fetchJson'];
}): Promise<number | null> {
  const deadline = Date.now() + options.timeoutMs;
  const request = options.fetchJson
    ? (path: string, params?: Record<string, string | number | boolean | undefined>) =>
        options.fetchJson!(path, params)
    : async (path: string, params?: Record<string, string | number | boolean | undefined>) => {
        const url = new URL(path, options.restEndpoint);
        if (params) {
          Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
              url.searchParams.set(key, String(value));
            }
          });
        }
        const response = await fetch(url.toString(), { headers: { Accept: 'application/json' } });
        if (!response.ok) {
          throw new Error(`Score query failed: ${response.status}`);
        }
        return response.json();
      };

  while (Date.now() < deadline) {
    try {
      const payload = await request(`/virtengine/veid/v1/identity_score/${options.address}`);
      const record = payload as Record<string, unknown>;
      const scoreRecord =
        (record.score as Record<string, unknown> | undefined) ??
        (record.identity_score as Record<string, unknown> | undefined) ??
        (record.identityScore as Record<string, unknown> | undefined) ??
        record;
      const scoreValue =
        scoreRecord && typeof scoreRecord === 'object'
          ? Number((scoreRecord as Record<string, unknown>).score ?? scoreRecord.value ?? 0)
          : Number(record.score ?? 0);
      if (!Number.isNaN(scoreValue) && scoreValue > 0) {
        return scoreValue;
      }
    } catch {
      // ignore transient errors during polling
    }
    await new Promise((resolve) => setTimeout(resolve, options.intervalMs));
  }
  return null;
}

export async function submitCaptureScope(request: SubmissionRequest): Promise<SubmissionResult> {
  updateStatus(request.onStatus, 'pending');

  try {
    validateCapture(request.capture);

    const senderAddress = request.senderAddress || (await request.userKeyProvider.getAccountAddress());
    const userAddress = await request.userKeyProvider.getAccountAddress();
    if (senderAddress !== userAddress) {
      throw new SigningError('Sender address does not match wallet address', {
        senderAddress,
        userAddress,
      });
    }

    updateStatus(request.onStatus, 'encrypting');

    const scopeType = normalizeScopeType(request.scopeType);
    const payload = await buildCapturePayload(request.capture);
    const validators = await getValidatorKeys(request);

    let envelope: EncryptedPayloadEnvelope;
    try {
      envelope = await encryptPayloadForRecipients(payload, validators);
    } catch (error) {
      throw new EncryptionError('Failed to encrypt capture payload', {
        cause: error instanceof Error ? error.message : error,
      });
    }

    envelope.metadata = {
      ...envelope.metadata,
      ...buildEnvelopeMetadata(request.capture, scopeType),
    };

    const payloadHash = await computeHash(envelope.ciphertext, 'SHA-256');
    const uploadNonce = request.uploadNonce ?? createUploadNonce();
    const clientId = await request.clientKeyProvider.getClientId();

    await ensureApprovedClient(
      clientId,
      request.restEndpoint,
      request.fetchJson,
      request.approvedClientCheck
    );

    updateStatus(request.onStatus, 'signing');

    const { clientSignature, userSignature } = await createUploadSignatures({
      salt: request.capture.salt,
      deviceFingerprint: request.capture.metadata.deviceFingerprint,
      clientId,
      payloadHash,
      uploadNonce,
      clientKeyProvider: request.clientKeyProvider,
      userKeyProvider: request.userKeyProvider,
    });

    const uploadMetadata = await createUploadMetadata({
      salt: request.capture.salt,
      deviceFingerprint: request.capture.metadata.deviceFingerprint,
      clientId,
      clientSignature,
      userSignature,
      payloadHash,
      uploadNonce,
      captureTimestamp: resolveCaptureTimestamp(request.capture),
      geoHint: request.geoHint ?? '',
    });

    const scopeId = request.scopeId ?? createScopeId('scope');
    const message = buildUploadScopeMessage({
      senderAddress,
      scopeId,
      scopeType: request.scopeType,
      envelope,
      metadata: uploadMetadata,
    });

    updateStatus(request.onStatus, 'broadcasting');

    const broadcastResult = await request.broadcaster.broadcast(message, request.memo);
    if (broadcastResult.code !== 0) {
      throw new BroadcastError('Transaction broadcast failed', {
        code: broadcastResult.code,
        rawLog: broadcastResult.rawLog,
      });
    }

    updateStatus(request.onStatus, 'confirmed', {
      scopeId,
      txHash: broadcastResult.txHash,
    });

    let score: number | undefined;
    if (request.scorePoll?.enabled) {
      const pollTimeout = request.scorePoll.timeoutMs ?? DEFAULT_SCORE_POLL_TIMEOUT;
      const pollInterval = request.scorePoll.intervalMs ?? DEFAULT_SCORE_POLL_INTERVAL;
      const pollResult = await pollForScore({
        restEndpoint: request.restEndpoint,
        address: senderAddress,
        timeoutMs: pollTimeout,
        intervalMs: pollInterval,
        fetchJson: request.fetchJson,
      });
      if (pollResult === null) {
        throw new SubmissionTimeoutError('Timed out waiting for identity score', {
          timeoutMs: pollTimeout,
        });
      }
      score = pollResult;
      updateStatus(request.onStatus, 'scored', {
        scopeId,
        txHash: broadcastResult.txHash,
        score,
      });
    }

    return {
      scopeId,
      txHash: broadcastResult.txHash,
      payloadHash,
      envelope,
      score,
      status: score !== undefined ? 'scored' : 'confirmed',
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Submission failed';
    const submissionError = error instanceof SubmissionError ? error : undefined;
    updateStatus(request.onStatus, 'failed', {
      message,
      error: submissionError,
    });

    if (error instanceof CaptureValidationError) throw error;
    if (error instanceof EncryptionError) throw error;
    if (error instanceof SigningError) throw error;
    if (error instanceof BroadcastError) throw error;
    if (error instanceof SubmissionTimeoutError) throw error;

    throw new CaptureValidationError(message, {
      cause: error instanceof Error ? error.message : error,
    });
  }
}
