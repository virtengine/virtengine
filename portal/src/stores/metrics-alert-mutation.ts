import type { Alert, AlertEvent } from '@virtengine/portal/types/metrics';

export type MetricsAlertMutationAction = 'create' | 'delete' | 'acknowledge';
type AlertInput = Omit<Alert, 'id' | 'status' | 'lastFired'>;

export type MetricsAlertMutationRequest =
  | {
      readonly action: 'create';
      readonly subject: string;
      readonly alert: AlertInput;
      readonly alertId?: never;
      readonly eventId?: never;
    }
  | {
      readonly action: 'delete';
      readonly subject: string;
      readonly alertId: string;
      readonly alert?: never;
      readonly eventId?: never;
    }
  | {
      readonly action: 'acknowledge';
      readonly subject: string;
      readonly eventId: string;
      readonly alert?: never;
      readonly alertId?: never;
    };

export interface MetricsAlertMutationSubmission {
  readonly requestDigest: string;
  readonly idempotencyKey: string;
  readonly signal: AbortSignal;
}

export interface MetricsAlertCommittedResult {
  readonly status: 'committed';
  readonly code: number;
  readonly operationId: string;
  readonly requestDigest: string;
  readonly idempotencyKey: string;
  readonly affectedId: string;
  readonly request: MetricsAlertMutationRequest;
  readonly alerts: readonly Alert[];
  readonly alertEvents: readonly AlertEvent[];
}

export interface MetricsAlertMutationAdapter {
  mutate(
    request: Readonly<MetricsAlertMutationRequest>,
    submission: Readonly<MetricsAlertMutationSubmission>
  ): Promise<unknown>;
}

export class MetricsAlertMutationError extends Error {
  constructor(
    public readonly reason:
      | 'unavailable'
      | 'invalid_request'
      | 'invalid_committed_result'
      | 'request_changed'
  ) {
    super(reason);
    this.name = 'MetricsAlertMutationError';
  }
}

type ValidationReason = 'invalid_request' | 'invalid_committed_result';

const fail = (reason: ValidationReason): never => {
  throw new MetricsAlertMutationError(reason);
};

const text = (value: unknown): value is string =>
  typeof value === 'string' && value.length > 0 && value.trim() === value;
const finite = (value: unknown): value is number =>
  typeof value === 'number' && Number.isFinite(value);

const objectValue = (value: unknown, reason: ValidationReason): Record<PropertyKey, unknown> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) fail(reason);
  return value as Record<PropertyKey, unknown>;
};

const exactKeys = (
  source: Record<PropertyKey, unknown>,
  allowed: readonly string[],
  required: readonly string[],
  reason: ValidationReason
) => {
  const keys = Reflect.ownKeys(source);
  if (
    keys.some((key) => typeof key !== 'string' || !allowed.includes(key)) ||
    required.some((key) => !keys.includes(key))
  ) {
    fail(reason);
  }
};

const snapshotChannels = (value: unknown, reason: ValidationReason): readonly string[] => {
  if (!Array.isArray(value)) fail(reason);
  const channels = Array.from(value as unknown[]);
  if (!channels.every(text)) fail(reason);
  return Object.freeze(channels as string[]);
};

const alertInputKeys = [
  'name',
  'deploymentId',
  'metric',
  'condition',
  'threshold',
  'duration',
  'notificationChannels',
] as const;

const snapshotAlertInput = (value: unknown, reason: ValidationReason): Readonly<AlertInput> => {
  const source = objectValue(value, reason);
  exactKeys(
    source,
    alertInputKeys,
    ['name', 'metric', 'condition', 'threshold', 'duration', 'notificationChannels'],
    reason
  );
  const name = source.name;
  const deploymentId = source.deploymentId;
  const metric = source.metric;
  const condition = source.condition;
  const threshold = source.threshold;
  const duration = source.duration;
  const notificationChannels = snapshotChannels(source.notificationChannels, reason);
  if (
    !text(name) ||
    (deploymentId !== undefined && !text(deploymentId)) ||
    !['cpu', 'memory', 'storage', 'network'].includes(metric as string) ||
    !['gt', 'lt', 'eq'].includes(condition as string) ||
    !finite(threshold) ||
    !finite(duration) ||
    duration <= 0
  ) {
    fail(reason);
  }
  return Object.freeze({
    name,
    deploymentId: deploymentId as string | undefined,
    metric: metric as Alert['metric'],
    condition: condition as Alert['condition'],
    threshold,
    duration,
    notificationChannels,
  }) as Readonly<AlertInput>;
};

const snapshotAlert = (value: unknown): Readonly<Alert> => {
  const source = objectValue(value, 'invalid_committed_result');
  exactKeys(
    source,
    ['id', ...alertInputKeys, 'status', 'lastFired'],
    [
      'id',
      'name',
      'metric',
      'condition',
      'threshold',
      'duration',
      'notificationChannels',
      'status',
    ],
    'invalid_committed_result'
  );
  const id = source.id;
  const name = source.name;
  const deploymentId = source.deploymentId;
  const metric = source.metric;
  const condition = source.condition;
  const threshold = source.threshold;
  const duration = source.duration;
  const notificationChannels = snapshotChannels(
    source.notificationChannels,
    'invalid_committed_result'
  );
  const status = source.status;
  const lastFired = source.lastFired;
  if (
    !text(id) ||
    !text(name) ||
    (deploymentId !== undefined && !text(deploymentId)) ||
    !['cpu', 'memory', 'storage', 'network'].includes(metric as string) ||
    !['gt', 'lt', 'eq'].includes(condition as string) ||
    !finite(threshold) ||
    !finite(duration) ||
    duration <= 0 ||
    !['active', 'firing', 'resolved'].includes(status as string) ||
    (lastFired !== undefined && !finite(lastFired))
  ) {
    fail('invalid_committed_result');
  }
  return Object.freeze({
    id,
    name,
    deploymentId: deploymentId as string | undefined,
    metric: metric as Alert['metric'],
    condition: condition as Alert['condition'],
    threshold,
    duration,
    notificationChannels,
    status: status as Alert['status'],
    lastFired: lastFired as number | undefined,
  }) as Readonly<Alert>;
};

const snapshotEvent = (value: unknown): Readonly<AlertEvent> => {
  const source = objectValue(value, 'invalid_committed_result');
  exactKeys(
    source,
    [
      'id',
      'alertId',
      'alertName',
      'status',
      'value',
      'timestamp',
      'acknowledged',
      'acknowledgedBy',
      'acknowledgedAt',
    ],
    ['id', 'alertId', 'alertName', 'status', 'value', 'timestamp', 'acknowledged'],
    'invalid_committed_result'
  );
  const id = source.id;
  const alertId = source.alertId;
  const alertName = source.alertName;
  const status = source.status;
  const eventValue = source.value;
  const timestamp = source.timestamp;
  const acknowledged = source.acknowledged;
  const acknowledgedBy = source.acknowledgedBy;
  const acknowledgedAt = source.acknowledgedAt;
  if (
    !text(id) ||
    !text(alertId) ||
    !text(alertName) ||
    !['firing', 'resolved'].includes(status as string) ||
    !finite(eventValue) ||
    !finite(timestamp) ||
    typeof acknowledged !== 'boolean' ||
    (acknowledgedBy !== undefined && !text(acknowledgedBy)) ||
    (acknowledgedAt !== undefined && !finite(acknowledgedAt)) ||
    (acknowledged && (!text(acknowledgedBy) || !finite(acknowledgedAt)))
  ) {
    fail('invalid_committed_result');
  }
  return Object.freeze({
    id,
    alertId,
    alertName,
    status: status as AlertEvent['status'],
    value: eventValue,
    timestamp,
    acknowledged,
    acknowledgedBy: acknowledgedBy as string | undefined,
    acknowledgedAt: acknowledgedAt as number | undefined,
  });
};

const canonicalRequest = (
  value: unknown,
  reason: ValidationReason = 'invalid_request'
): Readonly<MetricsAlertMutationRequest> => {
  const source = objectValue(value, reason);
  const action = source.action;
  const subject = source.subject;
  const alert = source.alert;
  const alertId = source.alertId;
  const eventId = source.eventId;
  if (!text(subject)) fail(reason);

  if (action === 'create') {
    exactKeys(source, ['action', 'subject', 'alert'], ['action', 'subject', 'alert'], reason);
    return Object.freeze({
      action,
      subject: subject as string,
      alert: snapshotAlertInput(alert, reason),
    });
  }
  if (action === 'delete') {
    exactKeys(source, ['action', 'subject', 'alertId'], ['action', 'subject', 'alertId'], reason);
    if (!text(alertId)) fail(reason);
    return Object.freeze({ action, subject: subject as string, alertId: alertId as string });
  }
  if (action === 'acknowledge') {
    exactKeys(source, ['action', 'subject', 'eventId'], ['action', 'subject', 'eventId'], reason);
    if (!text(eventId)) fail(reason);
    return Object.freeze({ action, subject: subject as string, eventId: eventId as string });
  }
  return fail(reason);
};

const serialize = (value: unknown): string =>
  JSON.stringify(value, (_key, item) => {
    if (!item || typeof item !== 'object' || Array.isArray(item)) return item;
    return Object.fromEntries(
      Object.entries(item)
        .filter(([, entry]) => entry !== undefined)
        .sort(([left], [right]) => left.localeCompare(right))
    );
  });

const digestRequest = async (request: MetricsAlertMutationRequest): Promise<string> => {
  const bytes = new TextEncoder().encode(serialize(request));
  const digest = await crypto.subtle.digest('SHA-256', bytes);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
};

const snapshotList = <T>(
  value: unknown,
  snapshot: (item: unknown) => Readonly<T>
): readonly Readonly<T>[] => {
  if (!Array.isArray(value)) fail('invalid_committed_result');
  return Object.freeze(Array.from(value as unknown[], snapshot));
};

export const submitMetricsAlertMutation = async ({
  adapter,
  request,
  signal,
  isCurrent,
}: {
  adapter: MetricsAlertMutationAdapter;
  request: unknown;
  signal: AbortSignal;
  isCurrent: () => boolean;
}): Promise<Readonly<MetricsAlertCommittedResult>> => {
  const snapshot = canonicalRequest(request);
  const digest = await digestRequest(snapshot);
  if (signal.aborted || !isCurrent()) throw new MetricsAlertMutationError('request_changed');
  const value = await adapter.mutate(snapshot, {
    requestDigest: digest,
    idempotencyKey: digest,
    signal,
  });
  if (signal.aborted || !isCurrent()) throw new MetricsAlertMutationError('request_changed');

  const source = objectValue(value, 'invalid_committed_result');
  exactKeys(
    source,
    [
      'status',
      'code',
      'operationId',
      'requestDigest',
      'idempotencyKey',
      'affectedId',
      'request',
      'alerts',
      'alertEvents',
    ],
    [
      'status',
      'code',
      'operationId',
      'requestDigest',
      'idempotencyKey',
      'affectedId',
      'request',
      'alerts',
      'alertEvents',
    ],
    'invalid_committed_result'
  );
  const status = source.status;
  const code = source.code;
  const operationId = source.operationId;
  const requestDigest = source.requestDigest;
  const idempotencyKey = source.idempotencyKey;
  const affectedId = source.affectedId;
  const echoedRequest = canonicalRequest(source.request, 'invalid_committed_result');
  const alerts = snapshotList(source.alerts, snapshotAlert);
  const alertEvents = snapshotList(source.alertEvents, snapshotEvent);
  if (
    status !== 'committed' ||
    code !== 0 ||
    !text(operationId) ||
    requestDigest !== digest ||
    idempotencyKey !== digest ||
    !text(affectedId) ||
    (snapshot.action === 'delete' && affectedId !== snapshot.alertId) ||
    (snapshot.action === 'acknowledge' && affectedId !== snapshot.eventId) ||
    serialize(echoedRequest) !== serialize(snapshot)
  ) {
    fail('invalid_committed_result');
  }
  if (snapshot.action === 'create' && !alerts.some((alert) => alert.id === affectedId)) {
    fail('invalid_committed_result');
  }
  if (snapshot.action === 'delete' && alerts.some((alert) => alert.id === affectedId)) {
    fail('invalid_committed_result');
  }
  if (snapshot.action === 'acknowledge') {
    const event = alertEvents.find((candidate) => candidate.id === affectedId);
    if (!event?.acknowledged || !text(event.acknowledgedBy) || !finite(event.acknowledgedAt)) {
      fail('invalid_committed_result');
    }
  }
  if (signal.aborted || !isCurrent()) throw new MetricsAlertMutationError('request_changed');
  return Object.freeze({
    status,
    code,
    operationId,
    requestDigest,
    idempotencyKey,
    affectedId,
    request: snapshot,
    alerts,
    alertEvents,
  }) as Readonly<MetricsAlertCommittedResult>;
};
