import type { Alert, AlertEvent } from '@virtengine/portal/types/metrics';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type {
  MetricsAlertMutationAdapter,
  MetricsAlertMutationRequest,
  MetricsAlertMutationSubmission,
} from '@/stores/metrics-alert-mutation';
import { submitMetricsAlertMutation } from '@/stores/metrics-alert-mutation';
import { configureMetricsAlertMutations, useMetricsStore } from '@/stores/metricsStore';

const { MockMultiProviderClient } = vi.hoisted(() => {
  class MockMultiProviderClient {
    initialize = vi.fn().mockResolvedValue(undefined);
    getProviders = vi.fn(() => [{ address: 've1provider', name: 'Provider One' }]);
    listAllDeployments = vi.fn().mockResolvedValue([{ id: 'dep-1', providerId: 've1provider' }]);
    getClient = vi.fn(() => ({
      getDeploymentMetrics: vi.fn().mockResolvedValue({
        cpu: { usage: 1, limit: 2 },
        memory: { usage: 2, limit: 4 },
        storage: { usage: 3, limit: 6 },
        network: { rxBytes: 10, txBytes: 20 },
      }),
    }));
  }
  return { MockMultiProviderClient };
});

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: MockMultiProviderClient,
}));

const initialState = useMetricsStore.getState();

const alertInput: Omit<Alert, 'id' | 'status' | 'lastFired'> = {
  name: 'Sustained CPU',
  deploymentId: 'deployment-1',
  metric: 'cpu',
  condition: 'gt',
  threshold: 85,
  duration: 300,
  notificationChannels: ['ops@example.com'],
};

const retainedAlert: Alert = {
  ...alertInput,
  id: 'alert-retained',
  status: 'active',
};

const retainedEvent: AlertEvent = {
  id: 'event-retained',
  alertId: retainedAlert.id,
  alertName: retainedAlert.name,
  status: 'firing',
  value: 91,
  timestamp: 1_700_000_000_000,
  acknowledged: false,
};

const committedResult = (
  request: Readonly<MetricsAlertMutationRequest>,
  submission: Readonly<MetricsAlertMutationSubmission>,
  alerts: readonly Alert[],
  alertEvents: readonly AlertEvent[]
) => ({
  status: 'committed' as const,
  code: 0,
  operationId: 'operation-committed',
  requestDigest: submission.requestDigest,
  idempotencyKey: submission.idempotencyKey,
  affectedId:
    request.action === 'create'
      ? (alerts[0]?.id ?? 'missing-created-alert')
      : request.action === 'delete'
        ? request.alertId
        : request.eventId,
  request,
  alerts,
  alertEvents,
});

const adapterFrom = (
  mutate: MetricsAlertMutationAdapter['mutate']
): MetricsAlertMutationAdapter => ({
  mutate,
});

const seedAlertState = () => {
  useMetricsStore.setState({ alerts: [retainedAlert], alertEvents: [retainedEvent] });
};

describe('metricsStore', () => {
  beforeEach(() => {
    useMetricsStore.setState(initialState, true);
    configureMetricsAlertMutations(null);
  });

  afterEach(() => {
    configureMetricsAlertMutations(null);
    useMetricsStore.setState(initialState, true);
  });

  it('aggregates provider metrics from daemon clients', async () => {
    await useMetricsStore.getState().fetchMetrics();

    const state = useMetricsStore.getState();
    expect(state.summary).not.toBeNull();
    expect(state.deploymentMetrics).toHaveLength(1);
    expect(state.summary?.totalProviders).toBe(1);
  });

  it('rejects all alert actions when persistence is unavailable without changing alert state', async () => {
    seedAlertState();
    const before = useMetricsStore.getState();

    await expect(useMetricsStore.getState().createAlert(alertInput)).rejects.toMatchObject({
      reason: 'unavailable',
    });
    await expect(useMetricsStore.getState().deleteAlert(retainedAlert.id)).rejects.toMatchObject({
      reason: 'unavailable',
    });
    await expect(
      useMetricsStore.getState().acknowledgeAlertEvent(retainedEvent.id)
    ).rejects.toMatchObject({ reason: 'unavailable' });

    const state = useMetricsStore.getState();
    expect(state.alerts).toBe(before.alerts);
    expect(state.alertEvents).toBe(before.alertEvents);
    expect(state.alertMutationPending).toBe(false);
    expect(state.alertMutationsAvailable).toBe(false);
  });

  it('creates from the authoritative committed snapshot and echoes request integrity fields', async () => {
    const authoritativeAlert: Alert = {
      ...alertInput,
      id: 'alert-authoritative',
      status: 'firing',
      lastFired: 1_700_000_000_100,
    };
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) => {
      expect(request).toEqual({ action: 'create', subject: 've1subject', alert: alertInput });
      expect(submission.requestDigest).toMatch(/^[a-f0-9]{64}$/);
      expect(submission.idempotencyKey).toBe(submission.requestDigest);
      return Promise.resolve(
        committedResult(request, submission, [authoritativeAlert], [retainedEvent])
      );
    });
    configureMetricsAlertMutations(adapterFrom(mutate), '  ve1subject  ');

    await useMetricsStore.getState().createAlert(alertInput);

    expect(mutate).toHaveBeenCalledOnce();
    expect(useMetricsStore.getState().alerts).toEqual([authoritativeAlert]);
    expect(useMetricsStore.getState().alertEvents).toEqual([retainedEvent]);
  });

  it.each([
    ['nonfinite threshold', { ...alertInput, threshold: Number.POSITIVE_INFINITY }],
    ['blank channel', { ...alertInput, notificationChannels: [''] }],
    ['extraneous field', { ...alertInput, status: 'active' }],
  ])(
    'rejects an invalid create payload with a %s before adapter submission',
    async (_case, alert) => {
      const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn();
      configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

      await expect(
        useMetricsStore.getState().createAlert(alert as typeof alertInput)
      ).rejects.toMatchObject({ reason: 'invalid_request' });
      expect(mutate).not.toHaveBeenCalled();
    }
  );

  it('snapshots request getters once and does not retain mutable request aliases', async () => {
    const reads = new Map<string, number>();
    const getter =
      <T>(key: string, value: T) =>
      () => {
        reads.set(key, (reads.get(key) ?? 0) + 1);
        return value;
      };
    const channels = ['ops@example.com'];
    const alert = Object.defineProperties(
      {},
      {
        name: { enumerable: true, get: getter('name', alertInput.name) },
        deploymentId: { enumerable: true, get: getter('deploymentId', alertInput.deploymentId) },
        metric: { enumerable: true, get: getter('metric', alertInput.metric) },
        condition: { enumerable: true, get: getter('condition', alertInput.condition) },
        threshold: { enumerable: true, get: getter('threshold', alertInput.threshold) },
        duration: { enumerable: true, get: getter('duration', alertInput.duration) },
        notificationChannels: { enumerable: true, get: getter('notificationChannels', channels) },
      }
    ) as typeof alertInput;
    const requestPrototype = Object.defineProperties(
      {},
      {
        alertId: { get: getter('alertId', undefined) },
        eventId: { get: getter('eventId', undefined) },
      }
    );
    const request = Object.defineProperties(Object.create(requestPrototype), {
      action: { enumerable: true, get: getter('action', 'create' as const) },
      subject: { enumerable: true, get: getter('subject', 've1subject') },
      alert: { enumerable: true, get: getter('alert', alert) },
    }) as MetricsAlertMutationRequest;
    let capturedRequest: Readonly<MetricsAlertMutationRequest> | undefined;
    const adapter = adapterFrom((canonical, submission) => {
      capturedRequest = canonical;
      return Promise.resolve(committedResult(canonical, submission, [retainedAlert], []));
    });

    await submitMetricsAlertMutation({
      adapter,
      request,
      signal: new AbortController().signal,
      isCurrent: () => true,
    });
    channels[0] = 'mutated@example.com';

    expect([...reads.values()]).toEqual(Array(reads.size).fill(1));
    expect(
      capturedRequest?.action === 'create' && capturedRequest.alert.notificationChannels
    ).toEqual(['ops@example.com']);
    expect(Object.isFrozen(capturedRequest)).toBe(true);
  });

  it.each([
    ['create missing affected alert', { action: 'create', affectedId: 'alert-missing' }],
    ['delete retaining affected alert', { action: 'delete', affectedId: retainedAlert.id }],
    ['acknowledge missing actor and time', { action: 'acknowledge', affectedId: retainedEvent.id }],
  ])('rejects a committed %s postcondition mismatch', async (_case, mismatch) => {
    seedAlertState();
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) => {
      const alerts = mismatch.action === 'delete' ? [retainedAlert] : [];
      const alertEvents = mismatch.action === 'acknowledge' ? [retainedEvent] : [];
      return Promise.resolve({
        ...committedResult(request, submission, alerts, alertEvents),
        affectedId: mismatch.affectedId,
      });
    });
    configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

    const operation =
      mismatch.action === 'create'
        ? useMetricsStore.getState().createAlert(alertInput)
        : mismatch.action === 'delete'
          ? useMetricsStore.getState().deleteAlert(retainedAlert.id)
          : useMetricsStore.getState().acknowledgeAlertEvent(retainedEvent.id);
    await expect(operation).rejects.toMatchObject({ reason: 'invalid_committed_result' });
  });

  it.each([
    ['malformed', { status: 'committed', code: 0 }],
    ['mismatched', 'mismatched'],
  ])('leaves alert state unchanged for a %s committed receipt', async (_case, receipt) => {
    seedAlertState();
    const before = useMetricsStore.getState();
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) =>
      Promise.resolve(
        receipt === 'mismatched'
          ? {
              ...committedResult(request, submission, [], []),
              requestDigest: 'not-the-request-digest',
            }
          : receipt
      )
    );
    configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

    await expect(useMetricsStore.getState().createAlert(alertInput)).rejects.toMatchObject({
      reason: 'invalid_committed_result',
    });

    const state = useMetricsStore.getState();
    expect(state.alerts).toBe(before.alerts);
    expect(state.alertEvents).toBe(before.alertEvents);
    expect(state.alertMutationPending).toBe(false);
  });

  it('preserves alerts and events when delete is rejected by the adapter', async () => {
    seedAlertState();
    const before = useMetricsStore.getState();
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn(() =>
      Promise.reject(new Error('broadcast rejected'))
    );
    configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

    await expect(useMetricsStore.getState().deleteAlert(retainedAlert.id)).rejects.toThrow(
      'broadcast rejected'
    );

    const state = useMetricsStore.getState();
    expect(state.alerts).toBe(before.alerts);
    expect(state.alertEvents).toBe(before.alertEvents);
  });

  it('adopts the exact authoritative snapshot after a committed delete', async () => {
    seedAlertState();
    const authoritativeAlert: Alert = {
      ...retainedAlert,
      id: 'alert-server-retained',
      status: 'resolved',
    };
    const authoritativeEvent: AlertEvent = {
      ...retainedEvent,
      id: 'event-server-retained',
      alertId: authoritativeAlert.id,
    };
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) => {
      expect(request).toEqual({
        action: 'delete',
        subject: 've1subject',
        alertId: retainedAlert.id,
      });
      return Promise.resolve(
        committedResult(request, submission, [authoritativeAlert], [authoritativeEvent])
      );
    });
    configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

    await useMetricsStore.getState().deleteAlert(retainedAlert.id);

    expect(useMetricsStore.getState().alerts).toEqual([authoritativeAlert]);
    expect(useMetricsStore.getState().alertEvents).toEqual([authoritativeEvent]);
  });

  it('adopts the authoritative acknowledgement actor and time', async () => {
    seedAlertState();
    const acknowledgedEvent: AlertEvent = {
      ...retainedEvent,
      acknowledged: true,
      acknowledgedBy: 've1authoritative-actor',
      acknowledgedAt: 1_700_000_000_900,
    };
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) => {
      expect(request).toEqual({
        action: 'acknowledge',
        subject: 've1subject',
        eventId: retainedEvent.id,
      });
      return Promise.resolve(
        committedResult(request, submission, [retainedAlert], [acknowledgedEvent])
      );
    });
    configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

    await useMetricsStore.getState().acknowledgeAlertEvent(retainedEvent.id);

    expect(useMetricsStore.getState().alertEvents).toEqual([acknowledgedEvent]);
  });

  it('allows only one adapter call for duplicate concurrent operations', async () => {
    let resolveMutation: ((value: unknown) => void) | undefined;
    let capturedRequest: Readonly<MetricsAlertMutationRequest> | undefined;
    let capturedSubmission: Readonly<MetricsAlertMutationSubmission> | undefined;
    const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) => {
      capturedRequest = request;
      capturedSubmission = submission;
      return new Promise((resolve) => {
        resolveMutation = resolve;
      });
    });
    configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');

    const first = useMetricsStore.getState().createAlert(alertInput);
    const duplicate = useMetricsStore.getState().createAlert(alertInput);

    await expect(duplicate).rejects.toMatchObject({ reason: 'request_changed' });
    await vi.waitFor(() => expect(mutate).toHaveBeenCalledOnce());
    expect(capturedRequest).toBeDefined();
    expect(capturedSubmission).toBeDefined();
    resolveMutation?.(
      committedResult(capturedRequest!, capturedSubmission!, [retainedAlert], [retainedEvent])
    );
    await first;

    expect(mutate).toHaveBeenCalledOnce();
  });

  it.each(['adapter', 'subject'])(
    'aborts an in-flight operation when the %s changes and ignores a late result',
    async (change) => {
      seedAlertState();
      const before = useMetricsStore.getState();
      let resolveMutation: ((value: unknown) => void) | undefined;
      let capturedRequest: Readonly<MetricsAlertMutationRequest> | undefined;
      let capturedSubmission: Readonly<MetricsAlertMutationSubmission> | undefined;
      const mutate: MetricsAlertMutationAdapter['mutate'] = vi.fn((request, submission) => {
        capturedRequest = request;
        capturedSubmission = submission;
        return new Promise((resolve) => {
          resolveMutation = resolve;
        });
      });
      const adapter = adapterFrom(mutate);
      configureMetricsAlertMutations(adapter, 've1subject');
      const pending = useMetricsStore.getState().deleteAlert(retainedAlert.id);
      await vi.waitFor(() => expect(mutate).toHaveBeenCalledOnce());

      const signal = capturedSubmission?.signal;
      configureMetricsAlertMutations(
        change === 'adapter' ? adapterFrom(vi.fn(() => Promise.resolve(undefined))) : adapter,
        change === 'subject' ? 've1different-subject' : 've1subject'
      );
      expect(signal?.aborted).toBe(true);
      resolveMutation?.(committedResult(capturedRequest!, capturedSubmission!, [], []));

      await expect(pending).rejects.toMatchObject({ reason: 'request_changed' });
      const state = useMetricsStore.getState();
      expect(state.alerts).toBe(before.alerts);
      expect(state.alertEvents).toBe(before.alertEvents);
      expect(state.alertMutationPending).toBe(false);
    }
  );
});
