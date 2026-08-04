import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { DashboardConfig } from '@virtengine/portal/types/metrics';
import { CustomDashboard } from '@/components/metrics/CustomDashboard';
import type {
  DashboardConfigMutationAdapter,
  DashboardConfigMutationRequest,
  DashboardConfigMutationSubmission,
} from '@/stores/dashboard-config-mutation';
import {
  configureDashboardConfigMutations,
  DEFAULT_DASHBOARD,
  useDashboardConfigStore,
} from '@/stores/dashboardConfigStore';

vi.mock('@/components/metrics/MetricCard', () => ({
  MetricCard: ({ title }: { title: string }) => <div>{title}</div>,
}));
vi.mock('@/components/metrics/TimeSeriesChart', () => ({
  TimeSeriesChart: ({ title }: { title: string }) => <div>{title}</div>,
}));
vi.mock('@/components/metrics/AlertsPanel', () => ({
  AlertsPanel: () => <div>Alerts widget</div>,
}));
vi.mock('@/components/metrics/ProviderBreakdown', () => ({
  ProviderBreakdown: () => <div>Provider widget</div>,
}));

const widget = {
  id: 'widget-authoritative',
  type: 'metric-card' as const,
  title: 'Authoritative CPU',
  config: { metric: 'cpu' },
  position: { x: 0, y: 0, w: 3, h: 2 },
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

const adapterFrom = (
  mutate: DashboardConfigMutationAdapter['mutate']
): DashboardConfigMutationAdapter => ({ mutate });

const committed = (
  request: Readonly<DashboardConfigMutationRequest>,
  submission: Readonly<DashboardConfigMutationSubmission>,
  dashboards: readonly DashboardConfig[],
  affectedId: string
) => ({
  status: 'committed' as const,
  code: 0 as const,
  operationId: 'operation-committed',
  revision: 1,
  requestDigest: submission.requestDigest,
  idempotencyKey: submission.idempotencyKey,
  affectedId,
  request,
  dashboards,
});

const dispatch = async (event: () => boolean) => {
  await act(async () => {
    event();
    await Promise.resolve();
  });
};

const configure = async (adapter: DashboardConfigMutationAdapter | null) => {
  await act(async () => {
    configureDashboardConfigMutations(adapter, adapter ? 've1subject' : '');
    await Promise.resolve();
  });
};

const installCustomDashboard = async (value = dashboard()) => {
  await act(async () => {
    useDashboardConfigStore.setState({
      dashboards: [DEFAULT_DASHBOARD, value],
      activeDashboardId: value.id,
      isEditing: false,
      error: null,
    });
    await Promise.resolve();
  });
};

const openCreate = async () => {
  await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'New' })));
  await dispatch(() =>
    fireEvent.change(screen.getByRole('textbox', { name: 'Dashboard name' }), {
      target: { value: 'Operations' },
    })
  );
};

const openAddWidget = async () => {
  await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'Edit' })));
  await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'Add Widget' })));
};

describe('CustomDashboard', () => {
  beforeEach(async () => configure(null));
  afterEach(async () => configure(null));

  it('shows only the built-in Overview and masks cached custom dashboards when unavailable', async () => {
    await act(async () => {
      useDashboardConfigStore.setState({
        dashboards: [DEFAULT_DASHBOARD, dashboard({ name: 'Cached custom' })],
        activeDashboardId: 'dashboard-authoritative',
      });
      await Promise.resolve();
    });

    render(<CustomDashboard />);

    expect(screen.getByRole('alert')).toHaveTextContent('Dashboard persistence is unavailable.');
    expect(screen.getByRole('tab', { name: /Overview/ })).toBeVisible();
    expect(screen.queryByText('Cached custom')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'New' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Edit' })).toBeDisabled();
    expect(screen.queryByText('Authoritative CPU')).not.toBeInTheDocument();
  });

  it('keeps create open while pending, then closes on commit and selects the authoritative ID', async () => {
    let resolveMutation: ((value: unknown) => void) | undefined;
    let request: Readonly<DashboardConfigMutationRequest> | undefined;
    let submission: Readonly<DashboardConfigMutationSubmission> | undefined;
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn(
      (nextRequest, nextSubmission) => {
        request = nextRequest;
        submission = nextSubmission;
        return new Promise((resolve) => {
          resolveMutation = resolve;
        });
      }
    );
    await configure(adapterFrom(mutate));
    render(<CustomDashboard />);
    await openCreate();

    const createButton = screen.getByRole('button', { name: 'Create' });
    await dispatch(() => fireEvent.click(createButton));

    await waitFor(() => expect(mutate).toHaveBeenCalledOnce());
    expect(screen.getByRole('textbox', { name: 'Dashboard name' })).toHaveValue('Operations');
    expect(screen.getByRole('status')).toHaveTextContent('Saving dashboard change...');
    expect(createButton).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeDisabled();
    await dispatch(() => fireEvent.click(createButton));
    expect(mutate).toHaveBeenCalledOnce();

    await act(async () => {
      resolveMutation?.(committed(request!, submission!, [dashboard()], 'dashboard-authoritative'));
      await Promise.resolve();
    });

    await waitFor(() =>
      expect(screen.queryByRole('textbox', { name: 'Dashboard name' })).not.toBeInTheDocument()
    );
    expect(screen.getByRole('tab', { name: 'Operations' })).toHaveAttribute('data-state', 'active');
    expect(useDashboardConfigStore.getState().activeDashboardId).toBe('dashboard-authoritative');
  });

  it.each([
    ['a rejected create', () => Promise.reject(new Error('broadcast rejected'))],
    ['a malformed create', () => Promise.resolve({ status: 'committed', code: 0 })],
  ])('retains create input and reports an error for %s', async (_case, result) => {
    await configure(adapterFrom(vi.fn(result) as DashboardConfigMutationAdapter['mutate']));
    render(<CustomDashboard />);
    await openCreate();

    await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'Create' })));

    expect(await screen.findByText('Dashboard change was not committed.')).toBeVisible();
    expect(screen.getByRole('textbox', { name: 'Dashboard name' })).toHaveValue('Operations');
    expect(screen.getByRole('button', { name: 'Create' })).toBeEnabled();
  });

  it('keeps create open when the mutation generation changes during materialization', async () => {
    const adapter = adapterFrom((request, submission) => {
      const authoritative = dashboard();
      const materializedDashboard = Object.defineProperty({ ...authoritative }, 'name', {
        enumerable: true,
        get: () => {
          configureDashboardConfigMutations(null);
          return authoritative.name;
        },
      });
      return Promise.resolve(
        committed(request, submission, [materializedDashboard], 'dashboard-authoritative')
      );
    });
    await configure(adapter);
    render(<CustomDashboard />);
    await openCreate();

    await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'Create' })));

    await waitFor(() =>
      expect(screen.getByRole('textbox', { name: 'Dashboard name' })).toHaveValue('Operations')
    );
    expect(screen.getByRole('alert')).toHaveTextContent('Dashboard persistence is unavailable.');
    expect(useDashboardConfigStore.getState().dashboards).toEqual([DEFAULT_DASHBOARD]);
  });

  it('closes add only after a committed authoritative widget is installed', async () => {
    let resolveMutation: ((value: unknown) => void) | undefined;
    let request: Readonly<DashboardConfigMutationRequest> | undefined;
    let submission: Readonly<DashboardConfigMutationSubmission> | undefined;
    const mutate: DashboardConfigMutationAdapter['mutate'] = vi.fn(
      (nextRequest, nextSubmission) => {
        request = nextRequest;
        submission = nextSubmission;
        return new Promise((resolve) => {
          resolveMutation = resolve;
        });
      }
    );
    await configure(adapterFrom(mutate));
    await installCustomDashboard(dashboard({ layout: [] }));
    render(<CustomDashboard />);
    await openAddWidget();

    const preset = screen.getByRole('button', { name: 'CPU Usage' });
    await dispatch(() => fireEvent.click(preset));
    await waitFor(() => expect(mutate).toHaveBeenCalledOnce());
    expect(screen.getByText('Add Widget')).toBeVisible();
    expect(preset).toBeDisabled();

    await act(async () => {
      const addRequest = request as Extract<DashboardConfigMutationRequest, { action: 'add' }>;
      resolveMutation?.(
        committed(
          request!,
          submission!,
          [
            dashboard({
              layout: [{ id: 'widget-authoritative', ...addRequest.widget }],
            }),
          ],
          'widget-authoritative'
        )
      );
      await Promise.resolve();
    });

    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'Memory Usage' })).not.toBeInTheDocument()
    );
    expect(useDashboardConfigStore.getState().dashboards[1]?.layout[0]?.id).toBe(
      'widget-authoritative'
    );
  });

  it('retains the add panel after rejection', async () => {
    await configure(adapterFrom(() => Promise.reject(new Error('add rejected'))));
    await installCustomDashboard(dashboard({ layout: [] }));
    render(<CustomDashboard />);
    await openAddWidget();

    await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'CPU Usage' })));

    expect(await screen.findByText('Dashboard change was not committed.')).toBeVisible();
    expect(screen.getByText('Add Widget')).toBeVisible();
    expect(screen.getByRole('button', { name: 'CPU Usage' })).toBeEnabled();
  });

  it('catches delete rejection and retains the selected dashboard', async () => {
    const mutate = vi.fn(() => Promise.reject(new Error('delete rejected')));
    await configure(adapterFrom(mutate));
    await installCustomDashboard();
    render(<CustomDashboard />);

    await dispatch(() =>
      fireEvent.click(screen.getByRole('button', { name: 'Delete dashboard Operations' }))
    );

    expect(await screen.findByText('Dashboard change was not committed.')).toBeVisible();
    expect(screen.getByRole('tab', { name: 'Operations' })).toHaveAttribute('data-state', 'active');
  });

  it('catches remove rejection, retains the widget, and exposes an accessible remove control', async () => {
    const mutate = vi.fn(() => Promise.reject(new Error('remove rejected')));
    await configure(adapterFrom(mutate));
    await installCustomDashboard();
    render(<CustomDashboard />);
    await dispatch(() => fireEvent.click(screen.getByRole('button', { name: 'Edit' })));

    const removeButton = screen.getByRole('button', { name: 'Remove widget Authoritative CPU' });
    expect(removeButton).toHaveAttribute('title', 'Remove widget Authoritative CPU');
    await dispatch(() => fireEvent.click(removeButton));

    expect(await screen.findByText('Dashboard change was not committed.')).toBeVisible();
    expect(screen.getByText('Authoritative CPU')).toBeVisible();
    expect(removeButton).toBeEnabled();
  });

  it('disables conflicting controls while delete is pending and prevents duplicates', async () => {
    let rejectMutation: ((reason: unknown) => void) | undefined;
    const mutate = vi.fn(
      () =>
        new Promise((_resolve, reject) => {
          rejectMutation = reject;
        })
    );
    await configure(adapterFrom(mutate));
    await installCustomDashboard();
    render(<CustomDashboard />);

    const deleteButton = screen.getByRole('button', { name: 'Delete dashboard Operations' });
    await dispatch(() => fireEvent.click(deleteButton));

    await waitFor(() => expect(deleteButton).toBeDisabled());
    expect(screen.getByRole('button', { name: 'New' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Edit' })).toBeDisabled();
    expect(screen.getByRole('tab', { name: 'Operations' })).toBeDisabled();
    await dispatch(() => fireEvent.click(deleteButton));
    expect(mutate).toHaveBeenCalledOnce();

    await act(async () => {
      rejectMutation?.(new Error('delete rejected'));
      await Promise.resolve();
    });
    await waitFor(() => expect(deleteButton).toBeEnabled());
  });

  it('uses the wide widget class order when width is at least nine', async () => {
    await configure(adapterFrom(vi.fn()));
    await installCustomDashboard(
      dashboard({ layout: [{ ...widget, position: { ...widget.position, w: 9 } }] })
    );
    render(<CustomDashboard />);

    expect(
      screen.getByText('Authoritative CPU').closest('div[class="md:col-span-2 lg:col-span-3"]')
    ).toBeInTheDocument();
  });
});
