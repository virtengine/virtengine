import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { DashboardConfig } from '@virtengine/portal/types/metrics';
import {
  submitDashboardConfigMutation,
  type DashboardConfigMutationAdapter,
  type DashboardConfigMutationRequest,
  type DashboardConfigMutationSubmission,
} from '@/stores/dashboard-config-mutation';
import {
  configureDashboardConfigMutations,
  DEFAULT_DASHBOARD,
  useDashboardConfigStore,
} from '@/stores/dashboardConfigStore';

const widget = {
  id: 'widget-authoritative',
  type: 'metric-card' as const,
  title: 'CPU',
  config: { metric: 'cpu', refreshInterval: 30 },
  position: { x: 1, y: 2, w: 3, h: 4 },
};

const dashboard = (overrides: Partial<DashboardConfig> = {}): DashboardConfig => ({
  id: 'dashboard-authoritative',
  name: 'Operations',
  isDefault: false,
  layout: [widget],
  createdAt: 1_700_000_000_000,
  updatedAt: 1_700_000_000_100,
  ...overrides,
});

const committed = (
  request: Readonly<DashboardConfigMutationRequest>,
  submission: Readonly<DashboardConfigMutationSubmission>,
  dashboards: readonly DashboardConfig[],
  affectedId = request.action === 'create'
    ? (dashboards[0]?.id ?? 'missing')
    : request.action === 'add'
      ? (dashboards[0]?.layout[0]?.id ?? 'missing')
      : request.action === 'rename' || request.action === 'delete'
        ? request.dashboardId
        : request.widgetId
) => ({
  status: 'committed' as const,
  code: 0,
  operationId: 'operation-1',
  revision: 1,
  requestDigest: submission.requestDigest,
  idempotencyKey: submission.idempotencyKey,
  affectedId,
  request,
  dashboards,
});

const adapterFrom = (mutate: DashboardConfigMutationAdapter['mutate']) => ({ mutate });

describe('dashboardConfigStore', () => {
  beforeEach(() => configureDashboardConfigMutations(null));
  afterEach(() => configureDashboardConfigMutations(null));

  it('keeps one deeply frozen deterministic default and rejects unavailable mutations', async () => {
    expect(DEFAULT_DASHBOARD.createdAt).toBe(0);
    expect(DEFAULT_DASHBOARD.updatedAt).toBe(0);
    expect(Object.isFrozen(DEFAULT_DASHBOARD)).toBe(true);
    expect(Object.isFrozen(DEFAULT_DASHBOARD.layout)).toBe(true);
    expect(Object.isFrozen(DEFAULT_DASHBOARD.layout[0]?.config)).toBe(true);
    const before = useDashboardConfigStore.getState().dashboards;
    expect(Object.isFrozen(before)).toBe(true);
    configureDashboardConfigMutations(null);
    expect(useDashboardConfigStore.getState().dashboards).toBe(before);

    await expect(useDashboardConfigStore.getState().createDashboard('New')).rejects.toMatchObject({
      reason: 'unavailable',
    });
    expect(useDashboardConfigStore.getState().dashboards).toBe(before);
    expect(useDashboardConfigStore.getState().dashboardMutationsAvailable).toBe(false);
  });

  it('uses authoritative create identity and timestamps without sending authority-owned fields', async () => {
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn((request, submission) => {
      expect(request).toEqual({ action: 'create', subject: 've1subject', name: 'Operations' });
      expect(request).not.toHaveProperty('id');
      expect(request).not.toHaveProperty('createdAt');
      expect(submission.requestDigest).toMatch(/^[a-f0-9]{64}$/);
      expect(submission.idempotencyKey).toBe(submission.requestDigest);
      return Promise.resolve(committed(request, submission, [dashboard()]));
    });
    configureDashboardConfigMutations(adapterFrom(mutate), '  ve1subject  ');

    await useDashboardConfigStore.getState().createDashboard('Operations');

    const state = useDashboardConfigStore.getState();
    expect(state.activeDashboardId).toBe('dashboard-authoritative');
    expect(state.dashboards).toEqual([DEFAULT_DASHBOARD, dashboard()]);
    expect(state.dashboards[1]?.createdAt).toBe(1_700_000_000_000);
  });

  it.each([
    ['malformed result', () => ({ status: 'committed', code: 0 })],
    [
      'digest mismatch',
      (request: DashboardConfigMutationRequest, submission: DashboardConfigMutationSubmission) => ({
        ...committed(request, submission, [dashboard()]),
        requestDigest: 'wrong',
      }),
    ],
    [
      'create postcondition',
      (request: DashboardConfigMutationRequest, submission: DashboardConfigMutationSubmission) =>
        committed(request, submission, [dashboard({ name: 'Wrong' })]),
    ],
    [
      'duplicate dashboards',
      (request: DashboardConfigMutationRequest, submission: DashboardConfigMutationSubmission) =>
        committed(request, submission, [dashboard(), dashboard()]),
    ],
    [
      'duplicate widgets',
      (request: DashboardConfigMutationRequest, submission: DashboardConfigMutationSubmission) =>
        committed(request, submission, [dashboard({ layout: [widget, widget] })]),
    ],
    [
      'default from adapter',
      (request: DashboardConfigMutationRequest, submission: DashboardConfigMutationSubmission) =>
        committed(request, submission, [dashboard({ isDefault: true })]),
    ],
    [
      'reserved default ID',
      (request: DashboardConfigMutationRequest, submission: DashboardConfigMutationSubmission) =>
        committed(
          request,
          submission,
          [dashboard({ id: 'dashboard-default' })],
          'dashboard-default'
        ),
    ],
  ])('rejects %s without changing dashboard state', async (_name, makeResult) => {
    const before = useDashboardConfigStore.getState().dashboards;
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn((request, submission) =>
      Promise.resolve(makeResult(request, submission) as unknown)
    );
    configureDashboardConfigMutations(adapterFrom(mutate), 've1subject');
    const configured = useDashboardConfigStore.getState().dashboards;

    await expect(
      useDashboardConfigStore.getState().createDashboard('Operations')
    ).rejects.toMatchObject({ reason: 'invalid_committed_result' });
    expect(useDashboardConfigStore.getState().dashboards).toBe(configured);
    expect(before).toBe(configured);
  });

  it('preserves genuine adapter errors', async () => {
    const adapterError = new Error('adapter failed');
    configureDashboardConfigMutations(
      adapterFrom(() => Promise.reject(adapterError)),
      've1subject'
    );

    await expect(useDashboardConfigStore.getState().createDashboard('Operations')).rejects.toBe(
      adapterError
    );
  });

  it.each([2, 1])(
    'rejects revision %i after accepting revision 2 without rolling back state',
    async (staleRevision) => {
      let revision = 2;
      const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn((request, submission) => {
        const authoritative =
          request.action === 'rename' ? dashboard({ name: request.name }) : dashboard();
        return Promise.resolve({
          ...committed(request, submission, [authoritative]),
          revision,
        });
      });
      configureDashboardConfigMutations(adapterFrom(mutate), 've1subject');

      await useDashboardConfigStore.getState().createDashboard('Operations');
      const accepted = useDashboardConfigStore.getState().dashboards;
      revision = staleRevision;

      await expect(
        useDashboardConfigStore.getState().renameDashboard(dashboard().id, 'Stale name')
      ).rejects.toMatchObject({ reason: 'invalid_committed_result' });
      expect(useDashboardConfigStore.getState().dashboards).toBe(accepted);
      expect(useDashboardConfigStore.getState().dashboards[1]?.name).toBe('Operations');
    }
  );

  it('rejects deleting the default dashboard without calling the adapter', async () => {
    const mutate = vi.fn();
    configureDashboardConfigMutations(adapterFrom(mutate), 've1subject');

    await expect(
      useDashboardConfigStore.getState().deleteDashboard(DEFAULT_DASHBOARD.id)
    ).rejects.toMatchObject({ reason: 'invalid_request' });
    expect(mutate).not.toHaveBeenCalled();
    expect(useDashboardConfigStore.getState().dashboards).toEqual([DEFAULT_DASHBOARD]);
  });

  it('translates an adapter rejection after authority becomes stale', async () => {
    let reject: ((reason: unknown) => void) | undefined;
    let current = true;
    const pending = submitDashboardConfigMutation({
      adapter: adapterFrom(
        () =>
          new Promise((_resolve, fail) => {
            reject = fail;
          })
      ),
      request: { action: 'create', subject: 've1subject', name: 'Operations' },
      signal: new AbortController().signal,
      isCurrent: () => current,
    });
    await vi.waitFor(() => expect(reject).toBeDefined());
    current = false;
    reject?.(new Error('stale adapter failure'));

    await expect(pending).rejects.toMatchObject({ reason: 'request_changed' });
  });

  it('commits delete, add, remove, update position, update config, and rename snapshots', async () => {
    let current = dashboard();
    let revision = 0;
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn((request, submission) => {
      if (request.action === 'delete')
        return Promise.resolve({
          ...committed(request, submission, []),
          revision: ++revision,
        });
      if (request.action === 'add')
        current = dashboard({
          layout: [
            {
              ...widget,
              id: 'widget-added',
              title: request.widget.title,
              type: request.widget.type,
              config: request.widget.config,
              position: request.widget.position,
            },
          ],
        });
      if (request.action === 'remove') current = dashboard({ layout: [] });
      if (request.action === 'update-position')
        current = dashboard({ layout: [{ ...widget, position: request.position }] });
      if (request.action === 'update-config')
        current = dashboard({ layout: [{ ...widget, config: request.config }] });
      if (request.action === 'rename') current = dashboard({ name: request.name });
      return Promise.resolve({
        ...committed(
          request,
          submission,
          [current],
          request.action === 'add' ? 'widget-added' : undefined
        ),
        revision: ++revision,
      });
    });
    configureDashboardConfigMutations(adapterFrom(mutate), 've1subject');

    await useDashboardConfigStore.getState().addWidget(current.id, 'gauge', 'Gauge', {});
    expect(useDashboardConfigStore.getState().dashboards[1]?.layout[0]?.id).toBe('widget-added');
    await useDashboardConfigStore
      .getState()
      .updateWidgetPosition(current.id, 'widget-authoritative', { x: 4, y: 5, w: 6, h: 7 });
    await useDashboardConfigStore
      .getState()
      .updateWidgetConfig(current.id, 'widget-authoritative', { metric: 'memory' });
    await useDashboardConfigStore.getState().renameDashboard(current.id, 'Renamed');
    expect(useDashboardConfigStore.getState().dashboards[1]?.name).toBe('Renamed');
    await useDashboardConfigStore.getState().removeWidget(current.id, 'widget-authoritative');
    expect(useDashboardConfigStore.getState().dashboards[1]?.layout).toEqual([]);
    await useDashboardConfigStore.getState().deleteDashboard(current.id);
    expect(useDashboardConfigStore.getState().dashboards).toEqual([DEFAULT_DASHBOARD]);
  });

  it('rejects action postcondition mismatches', async () => {
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn((request, submission) =>
      Promise.resolve(committed(request, submission, [dashboard()]))
    );
    configureDashboardConfigMutations(adapterFrom(mutate), 've1subject');

    await expect(
      useDashboardConfigStore.getState().removeWidget(dashboard().id, widget.id)
    ).rejects.toMatchObject({ reason: 'invalid_committed_result' });
    await expect(
      useDashboardConfigStore.getState().renameDashboard(dashboard().id, 'Wrong')
    ).rejects.toMatchObject({ reason: 'invalid_committed_result' });
  });

  it('allows only one in-flight mutation', async () => {
    let resolve: ((value: unknown) => void) | undefined;
    let capturedRequest: Readonly<DashboardConfigMutationRequest> | undefined;
    let capturedSubmission: Readonly<DashboardConfigMutationSubmission> | undefined;
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn((request, submission) => {
      capturedRequest = request;
      capturedSubmission = submission;
      return new Promise((done) => {
        resolve = done;
      });
    });
    configureDashboardConfigMutations(adapterFrom(mutate), 've1subject');
    const first = useDashboardConfigStore.getState().createDashboard('Operations');
    const duplicate = useDashboardConfigStore.getState().createDashboard('Operations');

    await expect(duplicate).rejects.toMatchObject({ reason: 'request_changed' });
    await vi.waitFor(() => expect(mutate).toHaveBeenCalledOnce());
    resolve?.(committed(capturedRequest!, capturedSubmission!, [dashboard()]));
    await first;
  });

  it.each(['adapter', 'subject'])(
    'aborts and rejects a late result after %s replacement, including ABA',
    async (change) => {
      let resolve: ((value: unknown) => void) | undefined;
      let request: Readonly<DashboardConfigMutationRequest> | undefined;
      let submission: Readonly<DashboardConfigMutationSubmission> | undefined;
      const adapter = adapterFrom((nextRequest, nextSubmission) => {
        request = nextRequest;
        submission = nextSubmission;
        return new Promise((done) => {
          resolve = done;
        });
      });
      configureDashboardConfigMutations(adapter, 've1subject');
      const pending = useDashboardConfigStore.getState().createDashboard('Operations');
      await vi.waitFor(() => expect(submission).toBeDefined());

      configureDashboardConfigMutations(
        change === 'adapter' ? adapterFrom(vi.fn()) : adapter,
        change === 'subject' ? 've1other' : 've1subject'
      );
      if (change === 'adapter') configureDashboardConfigMutations(adapter, 've1subject');
      expect(submission?.signal.aborted).toBe(true);
      resolve?.(committed(request!, submission!, [dashboard()]));

      await expect(pending).rejects.toMatchObject({ reason: 'request_changed' });
      expect(useDashboardConfigStore.getState().dashboards).toEqual([DEFAULT_DASHBOARD]);
    }
  );

  it('translates an adapter rejection after configure aborts it', async () => {
    let reject: ((reason: unknown) => void) | undefined;
    configureDashboardConfigMutations(
      adapterFrom(
        () =>
          new Promise((_resolve, fail) => {
            reject = fail;
          })
      ),
      've1subject'
    );
    const pending = useDashboardConfigStore.getState().createDashboard('Operations');
    await vi.waitFor(() => expect(reject).toBeDefined());

    configureDashboardConfigMutations(null);
    reject?.(new Error('aborted adapter failure'));

    await expect(pending).rejects.toMatchObject({ reason: 'request_changed' });
  });

  it('snapshots getters once, freezes recursively, and retains no mutable aliases', async () => {
    const reads = new Map<string, number>();
    const getter =
      <T>(key: string, value: T) =>
      () => {
        reads.set(key, (reads.get(key) ?? 0) + 1);
        return value;
      };
    const config = { metric: 'cpu' };
    const request = Object.defineProperties(
      {},
      {
        action: { enumerable: true, get: getter('action', 'add') },
        subject: { enumerable: true, get: getter('subject', 've1subject') },
        dashboardId: { enumerable: true, get: getter('dashboardId', dashboard().id) },
        widget: {
          enumerable: true,
          get: getter(
            'widget',
            Object.defineProperties(
              {},
              {
                type: { enumerable: true, get: getter('type', 'metric-card') },
                title: { enumerable: true, get: getter('title', 'CPU') },
                config: { enumerable: true, get: getter('config', config) },
                position: { enumerable: true, get: getter('position', { x: 0, y: 0, w: 1, h: 1 }) },
              }
            )
          ),
        },
      }
    ) as DashboardConfigMutationRequest;
    let captured: Readonly<DashboardConfigMutationRequest> | undefined;
    await submitDashboardConfigMutation({
      adapter: adapterFrom((canonical, submission) => {
        captured = canonical;
        return Promise.resolve(
          committed(canonical, submission, [
            dashboard({ layout: [{ ...widget, config, position: { x: 0, y: 0, w: 1, h: 1 } }] }),
          ])
        );
      }),
      request,
      signal: new AbortController().signal,
      isCurrent: () => true,
    });
    config.metric = 'memory';

    expect([...reads.values()]).toEqual(Array(reads.size).fill(1));
    expect(captured?.action === 'add' && captured.widget.config.metric).toBe('cpu');
    expect(captured?.action === 'add' && Object.isFrozen(captured.widget.config)).toBe(true);
  });

  it('does not select unknown dashboards and resets globals on configure', () => {
    const state = useDashboardConfigStore.getState();
    state.setActiveDashboard('unknown');
    expect(useDashboardConfigStore.getState().activeDashboardId).toBe(DEFAULT_DASHBOARD.id);
    configureDashboardConfigMutations(adapterFrom(vi.fn()), 've1subject');
    expect(useDashboardConfigStore.getState().dashboardMutationsAvailable).toBe(true);
    configureDashboardConfigMutations(null);
    expect(useDashboardConfigStore.getState().dashboardMutationsAvailable).toBe(false);
  });
});
