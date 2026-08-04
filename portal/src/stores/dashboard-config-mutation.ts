import type {
  DashboardConfig,
  DashboardWidget,
  WidgetConfig,
  WidgetPosition,
  WidgetType,
} from '@virtengine/portal/types/metrics';

export type DashboardConfigMutationRequest =
  | { readonly action: 'create'; readonly subject: string; readonly name: string }
  | { readonly action: 'delete'; readonly subject: string; readonly dashboardId: string }
  | {
      readonly action: 'add';
      readonly subject: string;
      readonly dashboardId: string;
      readonly widget: Readonly<Omit<DashboardWidget, 'id'>>;
    }
  | {
      readonly action: 'remove';
      readonly subject: string;
      readonly dashboardId: string;
      readonly widgetId: string;
    }
  | {
      readonly action: 'update-position';
      readonly subject: string;
      readonly dashboardId: string;
      readonly widgetId: string;
      readonly position: Readonly<WidgetPosition>;
    }
  | {
      readonly action: 'update-config';
      readonly subject: string;
      readonly dashboardId: string;
      readonly widgetId: string;
      readonly config: Readonly<WidgetConfig>;
    }
  | {
      readonly action: 'rename';
      readonly subject: string;
      readonly dashboardId: string;
      readonly name: string;
    };

export type DashboardConfigMutationCommand = DashboardConfigMutationRequest extends infer Request
  ? Request extends DashboardConfigMutationRequest
    ? Omit<Request, 'subject'>
    : never
  : never;

export interface DashboardConfigMutationSubmission {
  readonly requestDigest: string;
  readonly idempotencyKey: string;
  readonly signal: AbortSignal;
}

export interface DashboardConfigCommittedResult {
  readonly status: 'committed';
  readonly code: 0;
  readonly operationId: string;
  readonly revision: number;
  readonly requestDigest: string;
  readonly idempotencyKey: string;
  readonly affectedId: string;
  readonly request: DashboardConfigMutationRequest;
  readonly dashboards: readonly DashboardConfig[];
}

export interface DashboardConfigMutationAdapter {
  mutate(
    request: Readonly<DashboardConfigMutationRequest>,
    submission: Readonly<DashboardConfigMutationSubmission>
  ): Promise<unknown>;
}

export class DashboardConfigMutationError extends Error {
  constructor(
    public readonly reason:
      | 'unavailable'
      | 'invalid_request'
      | 'invalid_committed_result'
      | 'request_changed'
  ) {
    super(reason);
    this.name = 'DashboardConfigMutationError';
  }
}

type ValidationReason = 'invalid_request' | 'invalid_committed_result';
type Source = Record<PropertyKey, unknown>;

const fail = (reason: ValidationReason): never => {
  throw new DashboardConfigMutationError(reason);
};

const objectValue = (value: unknown, reason: ValidationReason): Source => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) fail(reason);
  return value as Source;
};

const exactKeys = (
  source: Source,
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

const text = (value: unknown): value is string =>
  typeof value === 'string' && value.length > 0 && value.trim() === value;
const integer = (value: unknown): value is number => Number.isSafeInteger(value);
const nonNegativeInteger = (value: unknown): value is number => integer(value) && value >= 0;
const positiveInteger = (value: unknown): value is number => integer(value) && value > 0;

const widgetTypes: readonly WidgetType[] = [
  'metric-card',
  'time-series-chart',
  'gauge',
  'table',
  'heatmap',
  'alert-list',
];
const timeRanges = ['1h', '6h', '24h', '7d', '30d'] as const;

const snapshotConfig = (value: unknown, reason: ValidationReason): Readonly<WidgetConfig> => {
  const source = objectValue(value, reason);
  exactKeys(source, ['metric', 'deploymentId', 'timeRange', 'refreshInterval'], [], reason);
  const metric = source.metric;
  const deploymentId = source.deploymentId;
  const timeRange = source.timeRange;
  const refreshInterval = source.refreshInterval;
  if (
    (metric !== undefined && !text(metric)) ||
    (deploymentId !== undefined && !text(deploymentId)) ||
    (timeRange !== undefined && !timeRanges.includes(timeRange as (typeof timeRanges)[number])) ||
    (refreshInterval !== undefined && !positiveInteger(refreshInterval))
  ) {
    fail(reason);
  }
  return Object.freeze({
    ...(metric === undefined ? {} : { metric }),
    ...(deploymentId === undefined ? {} : { deploymentId }),
    ...(timeRange === undefined ? {} : { timeRange }),
    ...(refreshInterval === undefined ? {} : { refreshInterval }),
  }) as Readonly<WidgetConfig>;
};

const snapshotPosition = (value: unknown, reason: ValidationReason): Readonly<WidgetPosition> => {
  const source = objectValue(value, reason);
  exactKeys(source, ['x', 'y', 'w', 'h'], ['x', 'y', 'w', 'h'], reason);
  const x = source.x;
  const y = source.y;
  const w = source.w;
  const h = source.h;
  if (
    !nonNegativeInteger(x) ||
    !nonNegativeInteger(y) ||
    !positiveInteger(w) ||
    !positiveInteger(h)
  ) {
    fail(reason);
  }
  return Object.freeze({ x, y, w, h });
};

const snapshotWidgetInput = (
  value: unknown,
  reason: ValidationReason
): Readonly<Omit<DashboardWidget, 'id'>> => {
  const source = objectValue(value, reason);
  exactKeys(
    source,
    ['type', 'title', 'config', 'position'],
    ['type', 'title', 'config', 'position'],
    reason
  );
  const type = source.type;
  const title = source.title;
  const config = source.config;
  const position = source.position;
  if (!widgetTypes.includes(type as WidgetType) || !text(title)) fail(reason);
  return Object.freeze({
    type: type as WidgetType,
    title,
    config: snapshotConfig(config, reason),
    position: snapshotPosition(position, reason),
  });
};

const snapshotWidget = (value: unknown): Readonly<DashboardWidget> => {
  const source = objectValue(value, 'invalid_committed_result');
  exactKeys(
    source,
    ['id', 'type', 'title', 'config', 'position'],
    ['id', 'type', 'title', 'config', 'position'],
    'invalid_committed_result'
  );
  const id = source.id;
  const type = source.type;
  const title = source.title;
  const config = source.config;
  const position = source.position;
  if (!text(id) || !widgetTypes.includes(type as WidgetType) || !text(title)) {
    fail('invalid_committed_result');
  }
  return Object.freeze({
    id,
    type: type as WidgetType,
    title,
    config: snapshotConfig(config, 'invalid_committed_result'),
    position: snapshotPosition(position, 'invalid_committed_result'),
  });
};

const snapshotDashboard = (value: unknown): Readonly<DashboardConfig> => {
  const source = objectValue(value, 'invalid_committed_result');
  exactKeys(
    source,
    ['id', 'name', 'isDefault', 'layout', 'createdAt', 'updatedAt'],
    ['id', 'name', 'isDefault', 'layout', 'createdAt', 'updatedAt'],
    'invalid_committed_result'
  );
  const id = source.id;
  const name = source.name;
  const isDefault = source.isDefault;
  const layoutValue = source.layout;
  const createdAt = source.createdAt;
  const updatedAt = source.updatedAt;
  if (
    !text(id) ||
    id === 'dashboard-default' ||
    !text(name) ||
    isDefault !== false ||
    !Array.isArray(layoutValue)
  ) {
    fail('invalid_committed_result');
  }
  const layout = Object.freeze(Array.from(layoutValue as unknown[], snapshotWidget));
  if (
    !nonNegativeInteger(createdAt) ||
    !nonNegativeInteger(updatedAt) ||
    updatedAt < createdAt ||
    new Set(layout.map((widget) => widget.id)).size !== layout.length
  ) {
    fail('invalid_committed_result');
  }
  return Object.freeze({ id, name, isDefault, layout, createdAt, updatedAt });
};

const canonicalRequest = (
  value: unknown,
  reason: ValidationReason = 'invalid_request'
): Readonly<DashboardConfigMutationRequest> => {
  const source = objectValue(value, reason);
  const action = source.action;
  const subject = source.subject;
  if (!text(subject)) fail(reason);

  if (action === 'create') {
    exactKeys(source, ['action', 'subject', 'name'], ['action', 'subject', 'name'], reason);
    const name = source.name;
    if (!text(name)) fail(reason);
    return Object.freeze({ action, subject: subject as string, name: name as string });
  }

  const dashboardId = source.dashboardId;
  if (!text(dashboardId)) fail(reason);
  if (action === 'delete') {
    exactKeys(
      source,
      ['action', 'subject', 'dashboardId'],
      ['action', 'subject', 'dashboardId'],
      reason
    );
    return Object.freeze({
      action,
      subject: subject as string,
      dashboardId: dashboardId as string,
    });
  }
  if (action === 'add') {
    exactKeys(
      source,
      ['action', 'subject', 'dashboardId', 'widget'],
      ['action', 'subject', 'dashboardId', 'widget'],
      reason
    );
    const widget = source.widget;
    return Object.freeze({
      action,
      subject: subject as string,
      dashboardId: dashboardId as string,
      widget: snapshotWidgetInput(widget, reason),
    });
  }

  const widgetId = source.widgetId;
  if (action === 'remove') {
    exactKeys(
      source,
      ['action', 'subject', 'dashboardId', 'widgetId'],
      ['action', 'subject', 'dashboardId', 'widgetId'],
      reason
    );
    if (!text(widgetId)) fail(reason);
    return Object.freeze({
      action,
      subject: subject as string,
      dashboardId: dashboardId as string,
      widgetId: widgetId as string,
    });
  }
  if (action === 'update-position') {
    exactKeys(
      source,
      ['action', 'subject', 'dashboardId', 'widgetId', 'position'],
      ['action', 'subject', 'dashboardId', 'widgetId', 'position'],
      reason
    );
    const position = source.position;
    if (!text(widgetId)) fail(reason);
    return Object.freeze({
      action,
      subject: subject as string,
      dashboardId: dashboardId as string,
      widgetId: widgetId as string,
      position: snapshotPosition(position, reason),
    });
  }
  if (action === 'update-config') {
    exactKeys(
      source,
      ['action', 'subject', 'dashboardId', 'widgetId', 'config'],
      ['action', 'subject', 'dashboardId', 'widgetId', 'config'],
      reason
    );
    const config = source.config;
    if (!text(widgetId)) fail(reason);
    return Object.freeze({
      action,
      subject: subject as string,
      dashboardId: dashboardId as string,
      widgetId: widgetId as string,
      config: snapshotConfig(config, reason),
    });
  }
  if (action === 'rename') {
    exactKeys(
      source,
      ['action', 'subject', 'dashboardId', 'name'],
      ['action', 'subject', 'dashboardId', 'name'],
      reason
    );
    const name = source.name;
    if (!text(name)) fail(reason);
    return Object.freeze({
      action,
      subject: subject as string,
      dashboardId: dashboardId as string,
      name: name as string,
    });
  }
  return fail(reason);
};

export const serializeDashboardConfigRequest = (value: unknown): string =>
  JSON.stringify(value, (_key, item) => {
    if (!item || typeof item !== 'object' || Array.isArray(item)) return item;
    return Object.fromEntries(
      Object.entries(item)
        .filter(([, entry]) => entry !== undefined)
        .sort(([left], [right]) => left.localeCompare(right))
    );
  });

const digestRequest = async (request: DashboardConfigMutationRequest): Promise<string> => {
  const bytes = new TextEncoder().encode(serializeDashboardConfigRequest(request));
  const digest = await crypto.subtle.digest('SHA-256', bytes);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
};

const validatePostcondition = (
  request: Readonly<DashboardConfigMutationRequest>,
  affectedId: string,
  dashboards: readonly Readonly<DashboardConfig>[]
) => {
  if (request.action === 'create') {
    const dashboard = dashboards.find((candidate) => candidate.id === affectedId);
    if (!dashboard || dashboard.name !== request.name) fail('invalid_committed_result');
    return;
  }
  if (request.action === 'delete') {
    if (
      affectedId !== request.dashboardId ||
      dashboards.some((candidate) => candidate.id === affectedId)
    ) {
      fail('invalid_committed_result');
    }
    return;
  }
  const dashboard = dashboards.find((candidate) => candidate.id === request.dashboardId);
  if (!dashboard) fail('invalid_committed_result');
  const authoritativeDashboard = dashboard as Readonly<DashboardConfig>;
  if (request.action === 'add') {
    const widget = authoritativeDashboard.layout.find((candidate) => candidate.id === affectedId);
    if (
      !widget ||
      widget.type !== request.widget.type ||
      widget.title !== request.widget.title ||
      serializeDashboardConfigRequest(widget.config) !==
        serializeDashboardConfigRequest(request.widget.config) ||
      serializeDashboardConfigRequest(widget.position) !==
        serializeDashboardConfigRequest(request.widget.position)
    ) {
      fail('invalid_committed_result');
    }
    return;
  }
  if (affectedId !== (request.action === 'rename' ? request.dashboardId : request.widgetId)) {
    fail('invalid_committed_result');
  }
  if (request.action === 'remove') {
    if (authoritativeDashboard.layout.some((candidate) => candidate.id === request.widgetId))
      fail('invalid_committed_result');
    return;
  }
  if (request.action === 'rename') {
    if (authoritativeDashboard.name !== request.name) fail('invalid_committed_result');
    return;
  }
  const widget = authoritativeDashboard.layout.find(
    (candidate) => candidate.id === request.widgetId
  );
  if (!widget) fail('invalid_committed_result');
  const actual = request.action === 'update-position' ? widget.position : widget.config;
  const expected = request.action === 'update-position' ? request.position : request.config;
  if (serializeDashboardConfigRequest(actual) !== serializeDashboardConfigRequest(expected)) {
    fail('invalid_committed_result');
  }
};

export const submitDashboardConfigMutation = async ({
  adapter,
  request,
  signal,
  isCurrent,
}: {
  adapter: DashboardConfigMutationAdapter;
  request: unknown;
  signal: AbortSignal;
  isCurrent: () => boolean;
}): Promise<Readonly<DashboardConfigCommittedResult>> => {
  const snapshot = canonicalRequest(request);
  const digest = await digestRequest(snapshot);
  if (signal.aborted || !isCurrent()) throw new DashboardConfigMutationError('request_changed');
  let value: unknown;
  try {
    value = await adapter.mutate(
      snapshot,
      Object.freeze({ requestDigest: digest, idempotencyKey: digest, signal })
    );
  } catch (error) {
    if (signal.aborted || !isCurrent()) {
      throw new DashboardConfigMutationError('request_changed');
    }
    throw error;
  }
  if (signal.aborted || !isCurrent()) throw new DashboardConfigMutationError('request_changed');

  const source = objectValue(value, 'invalid_committed_result');
  const keys = [
    'status',
    'code',
    'operationId',
    'revision',
    'requestDigest',
    'idempotencyKey',
    'affectedId',
    'request',
    'dashboards',
  ] as const;
  exactKeys(source, keys, keys, 'invalid_committed_result');
  const status = source.status;
  const code = source.code;
  const operationId = source.operationId;
  const revision = source.revision;
  const requestDigest = source.requestDigest;
  const idempotencyKey = source.idempotencyKey;
  const affectedId = source.affectedId;
  const echoedRequestValue = source.request;
  const dashboardValues = source.dashboards;
  const echoedRequest = canonicalRequest(echoedRequestValue, 'invalid_committed_result');
  if (!Array.isArray(dashboardValues)) fail('invalid_committed_result');
  const dashboards = Object.freeze(Array.from(dashboardValues as unknown[], snapshotDashboard));
  if (
    status !== 'committed' ||
    code !== 0 ||
    !text(operationId) ||
    !positiveInteger(revision) ||
    requestDigest !== digest ||
    idempotencyKey !== digest ||
    !text(affectedId) ||
    serializeDashboardConfigRequest(echoedRequest) !== serializeDashboardConfigRequest(snapshot) ||
    new Set(dashboards.map((dashboard) => dashboard.id)).size !== dashboards.length
  ) {
    fail('invalid_committed_result');
  }
  validatePostcondition(snapshot, affectedId as string, dashboards);
  if (signal.aborted || !isCurrent()) throw new DashboardConfigMutationError('request_changed');
  return Object.freeze({
    status,
    code,
    operationId,
    revision,
    requestDigest,
    idempotencyKey,
    affectedId,
    request: snapshot,
    dashboards,
  }) as Readonly<DashboardConfigCommittedResult>;
};
