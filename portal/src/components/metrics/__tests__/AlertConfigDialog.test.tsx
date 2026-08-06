import type { Alert, AlertEvent } from '@virtengine/portal/types/metrics';
import type { ReactNode } from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AlertConfigDialog } from '@/components/metrics/AlertConfigDialog';
import { AlertsPanel } from '@/components/metrics/AlertsPanel';
import type {
  MetricsAlertMutationAdapter,
  MetricsAlertMutationRequest,
  MetricsAlertMutationSubmission,
} from '@/stores/metrics-alert-mutation';
import { configureMetricsAlertMutations, useMetricsStore } from '@/stores/metricsStore';

vi.mock('@/lib/portal-adapter', () => ({
  MultiProviderClient: class MockMultiProviderClient {},
}));

vi.mock('@/components/ui/Select', () => ({
  Select: ({ children }: { children: ReactNode; value?: string; onValueChange?: () => void }) => (
    <div>{children}</div>
  ),
  SelectTrigger: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  SelectValue: ({ placeholder }: { placeholder?: string }) => <span>{placeholder}</span>,
  SelectContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  SelectItem: ({ children }: { children: ReactNode; value: string }) => <div>{children}</div>,
}));

const initialState = useMetricsStore.getState();

const adapterFrom = (
  mutate: MetricsAlertMutationAdapter['mutate']
): MetricsAlertMutationAdapter => ({
  mutate,
});

const dispatchEvent = async (event: () => boolean) => {
  await act(async () => {
    event();
    await Promise.resolve();
  });
};

const fillNameAndCreate = async (name = 'Sustained CPU') => {
  await dispatchEvent(() =>
    fireEvent.change(screen.getByLabelText('Alert Name'), { target: { value: name } })
  );
  await dispatchEvent(() => fireEvent.click(screen.getByRole('button', { name: 'Create Alert' })));
};

const renderDialog = async (onOpenChange = vi.fn()) => {
  const result = render(<AlertConfigDialog open onOpenChange={onOpenChange} />);
  await act(() => Promise.resolve());
  return result;
};

const unmountAndSettle = async (unmount: () => void) => {
  await act(async () => {
    unmount();
    await Promise.resolve();
  });
};

const authoritativeAlert: Alert = {
  id: 'alert-authoritative',
  name: 'Sustained CPU',
  metric: 'cpu',
  condition: 'gt',
  threshold: 80,
  duration: 300,
  notificationChannels: ['email'],
  status: 'active',
};

const cachedFiringAlert: Alert = {
  ...authoritativeAlert,
  id: 'alert-cached',
  name: 'Cached firing alert',
  status: 'firing',
};

const cachedEvent: AlertEvent = {
  id: 'event-cached',
  alertId: cachedFiringAlert.id,
  alertName: cachedFiringAlert.name,
  status: 'firing',
  value: 95,
  timestamp: 1_700_000_000_000,
  acknowledged: false,
};

describe('AlertConfigDialog', () => {
  beforeEach(async () => {
    await act(async () => {
      useMetricsStore.setState(initialState, true);
      configureMetricsAlertMutations(null);
      await Promise.resolve();
    });
  });

  afterEach(async () => {
    await act(async () => {
      configureMetricsAlertMutations(null);
      useMetricsStore.setState(initialState, true);
      await Promise.resolve();
    });
  });

  it('shows that persistence is unavailable and disables create by default', async () => {
    const { unmount } = await renderDialog();

    expect(screen.getByText('Alert persistence is unavailable.')).toBeVisible();
    await dispatchEvent(() =>
      fireEvent.change(screen.getByLabelText('Alert Name'), {
        target: { value: 'Sustained CPU' },
      })
    );
    expect(screen.getByRole('button', { name: 'Create Alert' })).toBeDisabled();
    await unmountAndSettle(unmount);
  });

  it('masks cached authoritative alert data while persistence is unavailable', async () => {
    await act(async () => {
      useMetricsStore.setState({ alerts: [cachedFiringAlert], alertEvents: [cachedEvent] });
      await Promise.resolve();
    });

    const { unmount } = render(<AlertsPanel />);

    expect(screen.getByText('Alert changes are unavailable.')).toBeVisible();
    expect(screen.queryByText('Cached firing alert')).not.toBeInTheDocument();
    expect(screen.queryByText('1 active')).not.toBeInTheDocument();
    expect(screen.queryByText('Recent Events')).not.toBeInTheDocument();
    await unmountAndSettle(unmount);
  });

  it('closes only after the exact committed receipt installs the authoritative alert', async () => {
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
    await act(async () => {
      configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');
      await Promise.resolve();
    });
    const onOpenChange = vi.fn();
    const { unmount } = await renderDialog(onOpenChange);

    await fillNameAndCreate();

    await waitFor(() => expect(mutate).toHaveBeenCalledOnce());
    expect(onOpenChange).not.toHaveBeenCalled();
    expect(useMetricsStore.getState().alerts).toEqual([]);

    await act(async () => {
      resolveMutation?.({
        status: 'committed',
        code: 0,
        operationId: 'operation-committed',
        requestDigest: capturedSubmission?.requestDigest,
        idempotencyKey: capturedSubmission?.idempotencyKey,
        affectedId: authoritativeAlert.id,
        request: capturedRequest,
        alerts: [authoritativeAlert],
        alertEvents: [],
      });
      await Promise.resolve();
    });

    await waitFor(() => expect(useMetricsStore.getState().alerts).toEqual([authoritativeAlert]));
    expect(onOpenChange).toHaveBeenCalledOnce();
    expect(onOpenChange).toHaveBeenCalledWith(false);
    await unmountAndSettle(unmount);
  });

  it('disables create while pending and prevents a duplicate submission', async () => {
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
    await act(async () => {
      configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');
      await Promise.resolve();
    });
    const { unmount } = await renderDialog();

    await fillNameAndCreate();

    const createButton = screen.getByRole('button', { name: 'Create Alert' });
    await waitFor(() => expect(mutate).toHaveBeenCalledOnce());
    await waitFor(() => expect(createButton).toBeDisabled());
    expect(screen.getByRole('status')).toHaveTextContent('Creating alert...');
    await dispatchEvent(() => fireEvent.click(createButton));
    expect(mutate).toHaveBeenCalledOnce();
    await act(async () => {
      resolveMutation?.({
        status: 'committed',
        code: 0,
        operationId: 'operation-committed',
        requestDigest: capturedSubmission?.requestDigest,
        idempotencyKey: capturedSubmission?.idempotencyKey,
        affectedId: authoritativeAlert.id,
        request: capturedRequest,
        alerts: [authoritativeAlert],
        alertEvents: [],
      });
      await Promise.resolve();
    });
    await waitFor(() => expect(useMetricsStore.getState().alertMutationPending).toBe(false));
    await unmountAndSettle(unmount);
  });

  it('blocks cancel and dialog dismissal while a mutation is pending', async () => {
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
    await act(async () => {
      configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');
      await Promise.resolve();
    });
    const onOpenChange = vi.fn();
    const { unmount } = await renderDialog(onOpenChange);

    await fillNameAndCreate();
    await waitFor(() => expect(screen.getByRole('status')).toBeVisible());
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeDisabled();
    await dispatchEvent(() => fireEvent.keyDown(document, { key: 'Escape' }));

    expect(onOpenChange).not.toHaveBeenCalled();
    await act(async () => {
      resolveMutation?.({
        status: 'committed',
        code: 0,
        operationId: 'operation-committed',
        requestDigest: capturedSubmission?.requestDigest,
        idempotencyKey: capturedSubmission?.idempotencyKey,
        affectedId: authoritativeAlert.id,
        request: capturedRequest,
        alerts: [authoritativeAlert],
        alertEvents: [],
      });
      await Promise.resolve();
    });
    await waitFor(() => expect(useMetricsStore.getState().alertMutationPending).toBe(false));
    await unmountAndSettle(unmount);
  });

  it.each([
    ['a rejected mutation', () => Promise.reject(new Error('broadcast rejected'))],
    ['a malformed committed receipt', () => Promise.resolve({ status: 'committed', code: 0 })],
  ])('keeps the dialog open and reports failure for %s', async (_case, result) => {
    const mutate = vi.fn(result) as MetricsAlertMutationAdapter['mutate'];
    await act(async () => {
      configureMetricsAlertMutations(adapterFrom(mutate), 've1subject');
      await Promise.resolve();
    });
    const onOpenChange = vi.fn();
    const { unmount } = await renderDialog(onOpenChange);

    await fillNameAndCreate();

    expect(await screen.findByText('Alert change was not committed.')).toBeVisible();
    expect(onOpenChange).not.toHaveBeenCalled();
    expect(useMetricsStore.getState().alerts).toEqual([]);
    await unmountAndSettle(unmount);
  });
});
