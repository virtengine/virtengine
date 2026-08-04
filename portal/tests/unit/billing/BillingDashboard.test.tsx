import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { BillingDashboard } from '@/components/billing/BillingDashboard';
import type {
  BillingWithdrawalAdapter,
  BillingWithdrawalContext,
  BillingWithdrawalRequest,
  BillingWithdrawalSubmission,
} from '@/components/billing/withdrawal-mutation';

type AuthorizedAction = () => void | Promise<void>;

const mfa = vi.hoisted(() => ({
  authorizations: [] as AuthorizedAction[],
}));

vi.mock('@/features/mfa', () => ({
  useMFAGate: () => ({
    gateAction: ({ onAuthorized }: { onAuthorized: AuthorizedAction }) => {
      mfa.authorizations.push(onAuthorized);
    },
    challengeProps: {},
  }),
}));

vi.mock('@/components/mfa', () => ({
  MFAChallenge: () => <div data-testid="mfa-challenge" />,
}));

vi.mock('@virtengine/portal/hooks/useBilling', () => ({
  useCurrentUsage: () => ({ data: undefined, isLoading: false }),
  useCostProjection: () => ({ data: undefined, isLoading: false }),
  useInvoices: () => ({ data: [], isLoading: false }),
  useUsageHistory: () => ({ data: [], isLoading: false }),
}));

vi.mock('@/components/billing/UsageSummaryCard', () => ({
  UsageSummaryCard: ({ title }: { title: string }) => <div>{title}</div>,
}));

vi.mock('@/components/billing/CostProjectionCard', () => ({
  CostProjectionCard: () => <div>Cost projection</div>,
}));

vi.mock('@/components/billing/CostTrendChart', () => ({
  CostTrendChart: () => <div>Cost trend</div>,
}));

const context: BillingWithdrawalContext = {
  chainId: 'virtengine-1',
  accountAddress: 'virt1billingauthority',
};

const deferred = <T,>() => {
  let resolve!: (value: T | PromiseLike<T>) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};

const committedEvidence = (
  request: Readonly<BillingWithdrawalRequest>,
  submission: Readonly<BillingWithdrawalSubmission>,
  overrides: Record<string, unknown> = {}
) => ({
  status: 'committed',
  code: 0,
  txHash: 'A1B2C3D4',
  blockHeight: 8675309,
  operationId: 'billing-withdrawal-1',
  requestDigest: submission.requestDigest,
  idempotencyKey: submission.idempotencyKey,
  request,
  ...overrides,
});

const clickAndAuthorize = async () => {
  fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
  const authorize = mfa.authorizations.at(-1);
  expect(authorize).toBeDefined();
  await act(async () => {
    await authorize?.();
  });
};

describe('BillingDashboard withdrawal mutation', () => {
  beforeEach(() => {
    mfa.authorizations.length = 0;
    vi.restoreAllMocks();
  });

  it('is unavailable by default and cannot fabricate a success', () => {
    render(<BillingDashboard />);

    expect(screen.getByRole('button', { name: 'Request Withdrawal' })).toBeDisabled();
    expect(screen.getByText(/withdrawals are unavailable/i)).toBeInTheDocument();
    expect(screen.queryByText(/withdrawal committed/i)).not.toBeInTheDocument();
    expect(mfa.authorizations).toHaveLength(0);
  });

  it('shows the exact committed receipt only after MFA authorization and adapter resolution', async () => {
    const result = deferred<unknown>();
    const requestWithdrawal = vi.fn(
      (
        request: Readonly<BillingWithdrawalRequest>,
        submission: Readonly<BillingWithdrawalSubmission>
      ) => {
        result.resolve(committedEvidence(request, submission));
        return result.promise;
      }
    );
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    render(<BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />);

    fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
    expect(requestWithdrawal).not.toHaveBeenCalled();
    expect(screen.queryByText(/withdrawal committed/i)).not.toBeInTheDocument();

    await act(async () => {
      await mfa.authorizations[0]?.();
    });

    expect(
      screen.getByText('Withdrawal committed in transaction A1B2C3D4 at height 8675309.')
    ).toBeInTheDocument();
  });

  it('does not show success while authoritative evidence is unresolved', async () => {
    const result = deferred<unknown>();
    const requestWithdrawal = vi.fn(() => result.promise);
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    render(<BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />);

    fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
    act(() => {
      void mfa.authorizations[0]?.();
    });

    expect(await screen.findByRole('button', { name: 'Submitting Withdrawal' })).toBeDisabled();
    expect(screen.queryByText(/withdrawal committed/i)).not.toBeInTheDocument();
  });

  it.each([
    ['malformed', { txHash: '' }],
    ['digest-mismatched', { requestDigest: 'wrong-digest' }],
    ['whitespace-padded transaction hash', { txHash: ' A1B2C3D4' }],
    ['whitespace-padded operation ID', { operationId: 'billing-withdrawal-1 ' }],
    [
      'whitespace-padded request chain ID',
      { request: { ...context, action: 'request_withdrawal', chainId: ' virtengine-1' } },
    ],
    [
      'whitespace-padded request account address',
      {
        request: {
          ...context,
          action: 'request_withdrawal',
          accountAddress: 'virt1billingauthority ',
        },
      },
    ],
  ])('rejects %s committed evidence', async (_label, evidenceOverride) => {
    const requestWithdrawal = vi.fn(
      (
        request: Readonly<BillingWithdrawalRequest>,
        submission: Readonly<BillingWithdrawalSubmission>
      ) => Promise.resolve(committedEvidence(request, submission, evidenceOverride))
    );
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    render(<BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />);

    await clickAndAuthorize();

    expect(screen.getByText(/withdrawal was not committed/i)).toBeInTheDocument();
    expect(screen.queryByText(/withdrawal committed in transaction/i)).not.toBeInTheDocument();
  });

  it('uses captured MFA authorizations only once after the first request commits', async () => {
    const result = deferred<unknown>();
    const requestWithdrawal = vi.fn(
      (
        request: Readonly<BillingWithdrawalRequest>,
        submission: Readonly<BillingWithdrawalSubmission>
      ) => {
        result.resolve(committedEvidence(request, submission));
        return result.promise;
      }
    );
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    render(<BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />);

    const button = screen.getByRole('button', { name: 'Request Withdrawal' });
    fireEvent.click(button);
    fireEvent.click(button);
    expect(mfa.authorizations).toHaveLength(2);
    const firstAuthorization = mfa.authorizations[1];
    const secondAuthorization = mfa.authorizations[0];

    await act(async () => {
      await firstAuthorization?.();
    });

    expect(requestWithdrawal).toHaveBeenCalledTimes(1);
    expect(screen.getAllByText(/withdrawal committed in transaction A1B2C3D4/i)).toHaveLength(1);

    await act(async () => {
      await secondAuthorization?.();
    });

    expect(requestWithdrawal).toHaveBeenCalledTimes(1);
    expect(screen.getAllByText(/withdrawal committed in transaction A1B2C3D4/i)).toHaveLength(1);
  });

  it('does not invoke stale or replacement adapters when authority changes before authorization', async () => {
    const staleRequestWithdrawal = vi.fn<BillingWithdrawalAdapter['requestWithdrawal']>();
    const replacementRequestWithdrawal = vi.fn<BillingWithdrawalAdapter['requestWithdrawal']>();
    const staleAdapter: BillingWithdrawalAdapter = {
      requestWithdrawal: staleRequestWithdrawal,
    };
    const replacementAdapter: BillingWithdrawalAdapter = {
      requestWithdrawal: replacementRequestWithdrawal,
    };
    const { rerender } = render(
      <BillingDashboard withdrawalAdapter={staleAdapter} withdrawalContext={context} />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
    const authorize = mfa.authorizations[0];
    expect(authorize).toBeDefined();

    rerender(
      <BillingDashboard
        withdrawalAdapter={replacementAdapter}
        withdrawalContext={{ ...context, accountAddress: 'virt1newauthority' }}
      />
    );
    await act(async () => {
      await authorize?.();
    });

    expect(staleRequestWithdrawal).not.toHaveBeenCalled();
    expect(replacementRequestWithdrawal).not.toHaveBeenCalled();
  });

  it.each(['context', 'adapter'] as const)(
    'rejects stale MFA authorization after an %s A-to-B-to-A cycle',
    async (authorityKind) => {
      const requestWithdrawalA = vi.fn<BillingWithdrawalAdapter['requestWithdrawal']>();
      const requestWithdrawalB = vi.fn<BillingWithdrawalAdapter['requestWithdrawal']>();
      const adapterA: BillingWithdrawalAdapter = { requestWithdrawal: requestWithdrawalA };
      const adapterB: BillingWithdrawalAdapter = { requestWithdrawal: requestWithdrawalB };
      const contextB = { ...context, accountAddress: 'virt1replacementauthority' };
      const { rerender } = render(
        <BillingDashboard withdrawalAdapter={adapterA} withdrawalContext={context} />
      );

      fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
      const authorize = mfa.authorizations[0];
      expect(authorize).toBeDefined();

      rerender(
        <BillingDashboard
          withdrawalAdapter={authorityKind === 'adapter' ? adapterB : adapterA}
          withdrawalContext={authorityKind === 'context' ? contextB : context}
        />
      );
      rerender(<BillingDashboard withdrawalAdapter={adapterA} withdrawalContext={context} />);
      await act(async () => {
        await authorize?.();
      });

      expect(requestWithdrawalA).not.toHaveBeenCalled();
      expect(requestWithdrawalB).not.toHaveBeenCalled();
    }
  );

  it('does not invoke the adapter when authorization occurs after unmount', async () => {
    const requestWithdrawal = vi.fn<BillingWithdrawalAdapter['requestWithdrawal']>();
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    const { unmount } = render(
      <BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
    const authorize = mfa.authorizations[0];
    expect(authorize).toBeDefined();

    unmount();
    await act(async () => {
      await authorize?.();
    });

    expect(requestWithdrawal).not.toHaveBeenCalled();
  });

  it('cannot commit a late noncooperative result after authority context changes', async () => {
    const result = deferred<unknown>();
    let capturedRequest!: BillingWithdrawalRequest;
    let capturedSubmission!: BillingWithdrawalSubmission;
    const requestWithdrawal = vi.fn(
      (
        request: Readonly<BillingWithdrawalRequest>,
        submission: Readonly<BillingWithdrawalSubmission>
      ) => {
        capturedRequest = request;
        capturedSubmission = submission;
        return result.promise;
      }
    );
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    const { rerender } = render(
      <BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
    act(() => {
      void mfa.authorizations[0]?.();
    });
    await vi.waitFor(() => expect(requestWithdrawal).toHaveBeenCalledTimes(1));

    rerender(
      <BillingDashboard
        withdrawalAdapter={adapter}
        withdrawalContext={{ ...context, accountAddress: 'virt1newauthority' }}
      />
    );
    await act(async () => {
      result.resolve(committedEvidence(capturedRequest, capturedSubmission));
      await result.promise;
    });

    expect(screen.queryByText(/withdrawal committed/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/withdrawal was not committed/i)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Request Withdrawal' })).toBeEnabled();
  });

  it('cannot commit when a receipt accessor changes authority during validation', async () => {
    const authorityChange: { current?: () => void } = {};
    const requestWithdrawal = vi.fn(
      (
        request: Readonly<BillingWithdrawalRequest>,
        submission: Readonly<BillingWithdrawalSubmission>
      ) => {
        const evidence = committedEvidence(request, submission);
        return Promise.resolve({
          ...evidence,
          get txHash() {
            authorityChange.current?.();
            return evidence.txHash;
          },
        });
      }
    );
    const adapter: BillingWithdrawalAdapter = { requestWithdrawal };
    const { rerender } = render(
      <BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />
    );
    authorityChange.current = () => {
      rerender(
        <BillingDashboard
          withdrawalAdapter={adapter}
          withdrawalContext={{ ...context, accountAddress: 'virt1newauthority' }}
        />
      );
    };

    await clickAndAuthorize();

    expect(requestWithdrawal).toHaveBeenCalledTimes(1);
    expect(screen.queryByText(/withdrawal committed/i)).not.toBeInTheDocument();
  });

  it('does not invoke the adapter when unmounted during request hashing', async () => {
    const digest = deferred<ArrayBuffer>();
    vi.spyOn(globalThis.crypto.subtle, 'digest').mockImplementation(() => digest.promise);
    const requestWithdrawal = vi.fn<BillingWithdrawalAdapter['requestWithdrawal']>();
    const adapter: BillingWithdrawalAdapter = {
      requestWithdrawal,
    };
    const { unmount } = render(
      <BillingDashboard withdrawalAdapter={adapter} withdrawalContext={context} />
    );

    fireEvent.click(screen.getByRole('button', { name: 'Request Withdrawal' }));
    act(() => {
      void mfa.authorizations[0]?.();
    });
    unmount();
    await act(async () => {
      digest.resolve(new ArrayBuffer(32));
      await digest.promise;
    });

    expect(requestWithdrawal).not.toHaveBeenCalled();
  });
});
