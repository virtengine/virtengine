import crypto from 'crypto';
import type { DeletionRequest } from '@/types/consent';

export interface DeletionSubmissionRequest {
  dataSubject: string;
  idempotencyKey: string;
}

export interface DeletionWorkflowAdapter {
  submitDeletion(request: DeletionSubmissionRequest): Promise<unknown>;
  listDeletions(dataSubject: string): Promise<unknown>;
}

export class DeletionWorkflowError extends Error {
  constructor(
    readonly code:
      | 'feature_unavailable'
      | 'invalid_subject'
      | 'invalid_receipt'
      | 'workflow_failed'
      | 'submission_in_progress'
  ) {
    super(code);
    this.name = 'DeletionWorkflowError';
  }
}

const normalizeSubject = (value: string): string => {
  const subject = value.trim();
  if (!subject || subject.length > 256) throw new DeletionWorkflowError('invalid_subject');
  return subject;
};

let configuredAdapter: DeletionWorkflowAdapter | undefined;

export const configureDeletionWorkflowAdapter = (adapter: DeletionWorkflowAdapter | undefined) => {
  configuredAdapter = adapter;
};

export const getDeletionWorkflowAdapter = () => configuredAdapter;

const materializeDeletion = (
  value: unknown,
  dataSubject: string,
  expectedIdempotencyKey?: string
): DeletionRequest => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new DeletionWorkflowError('invalid_receipt');
  }
  const source = value as Partial<DeletionRequest>;
  const requestedAt = typeof source.requestedAt === 'string' ? new Date(source.requestedAt) : null;
  if (
    typeof source.id !== 'string' ||
    !source.id.trim() ||
    source.dataSubject !== dataSubject ||
    !requestedAt ||
    Number.isNaN(requestedAt.getTime()) ||
    !['pending', 'blocked', 'processing', 'complete', 'failed'].includes(source.status ?? '') ||
    (source.blockers !== undefined &&
      (!Array.isArray(source.blockers) ||
        source.blockers.some((blocker) => typeof blocker !== 'string' || !blocker.trim())))
  ) {
    throw new DeletionWorkflowError('invalid_receipt');
  }
  const evidence = source as Partial<DeletionRequest> & { idempotencyKey?: unknown };
  if (
    (expectedIdempotencyKey !== undefined && evidence.idempotencyKey !== expectedIdempotencyKey) ||
    (source.error !== undefined && (typeof source.error !== 'string' || !source.error.trim())) ||
    (source.completedAt !== undefined &&
      (typeof source.completedAt !== 'string' ||
        Number.isNaN(new Date(source.completedAt).getTime()))) ||
    (source.status === 'complete' && source.completedAt === undefined) ||
    (source.status === 'failed' && source.error === undefined) ||
    (!['complete', 'failed'].includes(source.status ?? '') && source.completedAt !== undefined) ||
    (source.status === 'blocked' && (!source.blockers || source.blockers.length === 0)) ||
    (source.status !== 'blocked' && source.blockers !== undefined) ||
    (source.status !== 'failed' && source.error !== undefined)
  ) {
    throw new DeletionWorkflowError('invalid_receipt');
  }
  return Object.freeze({
    id: source.id.trim(),
    dataSubject,
    requestedAt: requestedAt.toISOString(),
    status: source.status as DeletionRequest['status'],
    blockers: source.blockers ? Object.freeze([...source.blockers]) : undefined,
    error: source.error,
    completedAt: source.completedAt,
  }) as DeletionRequest;
};

const pendingSubjects = new Set<string>();
const idempotencyKeyFor = (dataSubject: string) =>
  crypto.createHash('sha256').update(`consent-deletion:v1:${dataSubject}`).digest('hex');

export async function submitDeletionRequest(
  adapter: DeletionWorkflowAdapter | undefined,
  dataSubject: string
): Promise<DeletionRequest> {
  const subject = normalizeSubject(dataSubject);
  if (!adapter) throw new DeletionWorkflowError('feature_unavailable');
  if (pendingSubjects.has(subject)) throw new DeletionWorkflowError('submission_in_progress');
  const idempotencyKey = idempotencyKeyFor(subject);
  pendingSubjects.add(subject);
  try {
    return materializeDeletion(
      await adapter.submitDeletion({ dataSubject: subject, idempotencyKey }),
      subject,
      idempotencyKey
    );
  } catch (error) {
    if (error instanceof DeletionWorkflowError) throw error;
    throw new DeletionWorkflowError('workflow_failed');
  } finally {
    pendingSubjects.delete(subject);
  }
}

export async function listDeletionRequests(
  adapter: DeletionWorkflowAdapter | undefined,
  dataSubject: string
): Promise<readonly DeletionRequest[]> {
  const subject = normalizeSubject(dataSubject);
  if (!adapter) throw new DeletionWorkflowError('feature_unavailable');
  let value: unknown;
  try {
    value = await adapter.listDeletions(subject);
  } catch (error) {
    if (error instanceof DeletionWorkflowError) throw error;
    throw new DeletionWorkflowError('workflow_failed');
  }
  if (!Array.isArray(value)) throw new DeletionWorkflowError('invalid_receipt');
  const records = value.map((item) => materializeDeletion(item, subject));
  if (new Set(records.map((item) => item.id)).size !== records.length) {
    throw new DeletionWorkflowError('invalid_receipt');
  }
  return Object.freeze(records);
}
