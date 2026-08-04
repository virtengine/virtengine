import { StrictMode, type ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, type Mock } from 'vitest';
import type { MFAPolicy } from '@/lib/portal-adapter';
import * as mfaApi from '@/features/mfa/api';
import { useMFAStore } from '@/features/mfa/store';
import { useMFAGate } from '@/features/mfa/useMFAGate';

vi.mock('@/features/mfa/api', () => ({
  fetchFactors: vi.fn(),
  fetchPolicy: vi.fn(),
  fetchTrustedBrowsers: vi.fn(),
  fetchAuditLog: vi.fn(),
  createChallenge: vi.fn(),
}));

const policy: MFAPolicy = {
  id: 'policy-1',
  updatedAt: Date.now(),
  requiredFactorTypes: ['otp'],
  requiredFactorCount: 1,
  sensitiveTransactions: ['withdrawal'],
  allowTrustedBrowsers: true,
  trustedBrowserDurationSeconds: 3600,
  biometricForLowValue: false,
  lowValueThreshold: '100',
};

const challenge = {
  id: 'challenge-1',
  type: 'otp' as const,
  transactionType: 'withdrawal' as const,
  availableFactors: [],
  expiresAt: Date.now() + 60_000,
  transactionSummary: 'Authorize withdrawal',
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

function mockSuccessfulLoad(loadedPolicy: unknown = policy) {
  (mfaApi.fetchFactors as Mock).mockResolvedValue([]);
  (mfaApi.fetchPolicy as Mock).mockResolvedValue(loadedPolicy);
  (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
  (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);
}

function setReady(readyPolicy: MFAPolicy = policy) {
  useMFAStore.setState((state) => ({
    policyStatus: 'ready',
    policyGeneration: state.policyGeneration + 1,
    policy: readyPolicy,
    isLoading: false,
  }));
}

describe('useMFAGate', () => {
  beforeEach(() => {
    useMFAStore.getState().reset();
    vi.clearAllMocks();
    mockSuccessfulLoad();
    (mfaApi.createChallenge as Mock).mockResolvedValue(challenge);
  });

  it('waits for authoritative loading and does not authorize early', async () => {
    const policyRequest = deferred<MFAPolicy>();
    (mfaApi.fetchPolicy as Mock).mockReturnValue(policyRequest.promise);
    const onAuthorized = vi.fn();
    const { result } = renderHook(() => useMFAGate());

    let gatePromise!: Promise<void>;
    act(() => {
      gatePromise = result.current.gateAction({
        transactionType: 'delegation_change',
        onAuthorized,
      });
    });

    expect(result.current.isPending).toBe(true);
    expect(onAuthorized).not.toHaveBeenCalled();
    policyRequest.resolve(policy);
    await act(async () => gatePromise);

    expect(onAuthorized).toHaveBeenCalledTimes(1);
    expect(result.current.isPending).toBe(false);
    expect(mfaApi.createChallenge).not.toHaveBeenCalled();
  });

  it.each([
    ['rejected load', new Error('network unavailable')],
    ['malformed policy', { ...policy, sensitiveTransactions: ['not_known'] }],
  ])('fails closed for %s', async (_name, outcome) => {
    if (outcome instanceof Error) {
      (mfaApi.fetchPolicy as Mock).mockRejectedValue(outcome);
    } else {
      (mfaApi.fetchPolicy as Mock).mockResolvedValue(outcome);
    }
    const onAuthorized = vi.fn();
    const { result } = renderHook(() => useMFAGate());

    await expect(
      act(() => result.current.gateAction({ transactionType: 'delegation_change', onAuthorized }))
    ).rejects.toMatchObject({
      name: 'MFAAuthorizationError',
      code: 'policy_unavailable',
    });
    expect(onAuthorized).not.toHaveBeenCalled();
    expect(mfaApi.createChallenge).not.toHaveBeenCalled();
  });

  it('authorizes a ready non-sensitive transaction directly', async () => {
    setReady();
    const onAuthorized = vi.fn();
    const { result } = renderHook(() => useMFAGate());

    await act(() =>
      result.current.gateAction({ transactionType: 'delegation_change', onAuthorized })
    );

    expect(onAuthorized).toHaveBeenCalledTimes(1);
    expect(mfaApi.createChallenge).not.toHaveBeenCalled();
    expect(result.current.isPending).toBe(false);
  });

  it('challenges a sensitive transaction and consumes success once', async () => {
    setReady();
    const onAuthorized = vi.fn();
    const { result } = renderHook(() => useMFAGate());

    await act(() => result.current.gateAction({ transactionType: 'withdrawal', onAuthorized }));
    expect(mfaApi.createChallenge).toHaveBeenCalledWith('withdrawal');
    expect(onAuthorized).not.toHaveBeenCalled();
    expect(result.current.challengeProps.open).toBe(true);
    expect(result.current.isPending).toBe(true);

    await act(() => result.current.challengeProps.onSuccess());
    await act(() => result.current.challengeProps.onSuccess());

    expect(onAuthorized).toHaveBeenCalledTimes(1);
    expect(result.current.challengeProps.open).toBe(false);
    expect(result.current.isPending).toBe(false);
  });

  it('serializes duplicate gate calls while a challenge is pending', async () => {
    setReady();
    const authorizeA = vi.fn();
    const authorizeB = vi.fn();
    const { result } = renderHook(() => useMFAGate());

    await act(() =>
      result.current.gateAction({ transactionType: 'withdrawal', onAuthorized: authorizeA })
    );
    const successA = result.current.challengeProps.onSuccess;
    const ownerA = useMFAStore.getState().gateOwnerToken;

    await act(async () => {
      await expect(
        result.current.gateAction({ transactionType: 'withdrawal', onAuthorized: authorizeB })
      ).rejects.toMatchObject({ code: 'authorization_pending' });
    });
    expect(result.current.error).toMatch(/already in progress/i);
    expect(mfaApi.createChallenge).toHaveBeenCalledTimes(1);
    expect(result.current.challengeProps.open).toBe(true);
    expect(useMFAStore.getState().activeChallenge?.id).toBe('challenge-1');
    expect(useMFAStore.getState().gateOwnerToken).toBe(ownerA);
    expect(result.current.isPending).toBe(true);

    await act(() => successA());

    expect(authorizeA).toHaveBeenCalledTimes(1);
    expect(authorizeB).not.toHaveBeenCalled();
    expect(useMFAStore.getState().gateOwnerToken).toBeNull();
  });

  it('serializes sensitive gates across hook instances without invoking the second callback', async () => {
    setReady();
    const firstAuthorized = vi.fn();
    const secondAuthorized = vi.fn();
    const first = renderHook(() => useMFAGate());
    const second = renderHook(() => useMFAGate());

    await act(() =>
      first.result.current.gateAction({
        transactionType: 'withdrawal',
        onAuthorized: firstAuthorized,
      })
    );
    await act(async () => {
      await expect(
        second.result.current.gateAction({
          transactionType: 'withdrawal',
          onAuthorized: secondAuthorized,
        })
      ).rejects.toMatchObject({ code: 'authorization_pending' });
    });

    expect(mfaApi.createChallenge).toHaveBeenCalledTimes(1);
    expect(firstAuthorized).not.toHaveBeenCalled();
    expect(secondAuthorized).not.toHaveBeenCalled();
  });

  it('does not let challenge A success authorize challenge B after an ABA sequence', async () => {
    setReady();
    const authorizeA = vi.fn();
    const authorizeB = vi.fn();
    const { result } = renderHook(() => useMFAGate());

    await act(() =>
      result.current.gateAction({ transactionType: 'withdrawal', onAuthorized: authorizeA })
    );
    const staleSuccessA = result.current.challengeProps.onSuccess;
    act(() => result.current.challengeProps.onOpenChange(false));

    (mfaApi.createChallenge as Mock).mockResolvedValue({ ...challenge, id: 'challenge-2' });
    await act(() =>
      result.current.gateAction({ transactionType: 'withdrawal', onAuthorized: authorizeB })
    );
    await act(() => staleSuccessA());

    expect(authorizeA).not.toHaveBeenCalled();
    expect(authorizeB).not.toHaveBeenCalled();
    expect(result.current.challengeProps.open).toBe(true);
    await act(() => result.current.challengeProps.onSuccess());
    expect(authorizeB).toHaveBeenCalledTimes(1);
  });

  it('remains mounted and functional under Strict Mode effect replay', async () => {
    setReady();
    const onAuthorized = vi.fn();
    const wrapper = ({ children }: { children: ReactNode }) => <StrictMode>{children}</StrictMode>;
    const { result } = renderHook(() => useMFAGate(), { wrapper });

    await act(() => result.current.gateAction({ transactionType: 'withdrawal', onAuthorized }));
    await act(() => result.current.challengeProps.onSuccess());

    expect(onAuthorized).toHaveBeenCalledTimes(1);
  });

  it('keeps ownership and the dialog open after verification failure', async () => {
    setReady();
    const { result } = renderHook(() => useMFAGate());
    await act(() =>
      result.current.gateAction({ transactionType: 'withdrawal', onAuthorized: vi.fn() })
    );
    const owner = useMFAStore.getState().gateOwnerToken;

    act(() => result.current.challengeProps.onFailure());

    expect(result.current.challengeProps.open).toBe(true);
    expect(useMFAStore.getState().gateOwnerToken).toBe(owner);
    expect(useMFAStore.getState().activeChallenge?.id).toBe('challenge-1');
  });

  it.each(['reset', 'policy change', 'close'])(
    '%s invalidates a pending authorization',
    async (invalidation) => {
      setReady();
      const onAuthorized = vi.fn();
      const { result } = renderHook(() => useMFAGate());
      await act(() => result.current.gateAction({ transactionType: 'withdrawal', onAuthorized }));
      const staleSuccess = result.current.challengeProps.onSuccess;

      act(() => {
        if (invalidation === 'reset') useMFAStore.getState().reset();
        else if (invalidation === 'policy change')
          setReady({ ...policy, id: 'policy-2', sensitiveTransactions: [] });
        else result.current.challengeProps.onOpenChange(false);
      });
      await act(() => staleSuccess());

      expect(onAuthorized).not.toHaveBeenCalled();
    }
  );

  it('unmount invalidates a pending authorization', async () => {
    setReady();
    const onAuthorized = vi.fn();
    const { result, unmount } = renderHook(() => useMFAGate());
    await act(() => result.current.gateAction({ transactionType: 'withdrawal', onAuthorized }));
    const staleSuccess = result.current.challengeProps.onSuccess;

    unmount();
    await staleSuccess();

    expect(onAuthorized).not.toHaveBeenCalled();
  });

  it('unmount releases ownership while challenge creation is still pending', async () => {
    setReady();
    const challengeRequest = deferred<typeof challenge>();
    (mfaApi.createChallenge as Mock).mockReturnValue(challengeRequest.promise);
    const { result, unmount } = renderHook(() => useMFAGate());
    let gatePromise!: Promise<void>;
    act(() => {
      gatePromise = result.current.gateAction({
        transactionType: 'withdrawal',
        onAuthorized: vi.fn(),
      });
    });
    await waitFor(() => expect(useMFAStore.getState().gateOwnerToken).not.toBeNull());

    unmount();
    expect(useMFAStore.getState().gateOwnerToken).toBeNull();
    challengeRequest.resolve(challenge);
    await expect(gatePromise).rejects.toThrow('Stale MFA challenge');
    expect(useMFAStore.getState().activeChallenge).toBeNull();
  });

  it('rejects challenge creation failure without authorizing', async () => {
    setReady();
    (mfaApi.createChallenge as Mock).mockRejectedValue(new Error('challenge failed'));
    const onAuthorized = vi.fn();
    const { result } = renderHook(() => useMFAGate());
    let rejection: unknown;

    await act(async () => {
      try {
        await result.current.gateAction({ transactionType: 'withdrawal', onAuthorized });
      } catch (error) {
        rejection = error;
      }
    });

    expect(rejection).toEqual(new Error('challenge failed'));
    expect(onAuthorized).not.toHaveBeenCalled();
    await waitFor(() => expect(result.current.challengeProps.open).toBe(false));
    expect(result.current.isPending).toBe(false);
    expect(result.current.error).toMatch(/could not be completed/i);

    act(() => result.current.clearError());
    expect(result.current.error).toBeNull();
  });

  it('contains an authorized callback rejection and releases the gate', async () => {
    setReady();
    const onAuthorized = vi.fn().mockRejectedValue(new Error('submission failed'));
    const { result } = renderHook(() => useMFAGate());

    await act(() => result.current.gateAction({ transactionType: 'withdrawal', onAuthorized }));
    await act(() => result.current.challengeProps.onSuccess());

    expect(onAuthorized).toHaveBeenCalledTimes(1);
    expect(result.current.error).toMatch(/could not be completed/i);
    expect(result.current.isPending).toBe(false);
    expect(useMFAStore.getState().gateOwnerToken).toBeNull();
  });
});
