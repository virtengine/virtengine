/**
 * MFA gating hook for sensitive actions.
 * Creates MFA challenges when policy requires verification.
 */

'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { SensitiveTransactionType } from '@/lib/portal-adapter';
import { hasValidSensitiveTransactions, useMFAStore, type MFAGateOwnerToken } from './store';

export type MFAAuthorizationErrorCode = 'policy_unavailable' | 'authorization_pending';

export class MFAAuthorizationError extends Error {
  constructor(public readonly code: MFAAuthorizationErrorCode) {
    super(code);
    this.name = 'MFAAuthorizationError';
  }
}

const POLICY_ERROR = 'MFA policy is unavailable. Try again.';
const PENDING_ERROR = 'Another MFA authorization is already in progress.';
const AUTHORIZATION_ERROR = 'MFA authorization could not be completed. Try again.';

function authorizationErrorMessage(error: unknown): string {
  if (error instanceof MFAAuthorizationError) {
    return error.code === 'policy_unavailable' ? POLICY_ERROR : PENDING_ERROR;
  }
  return AUTHORIZATION_ERROR;
}

interface MFAGateActionOptions {
  transactionType: SensitiveTransactionType;
  actionDescription?: string;
  onAuthorized: () => void | Promise<void>;
}

interface PendingAction {
  ownerToken: MFAGateOwnerToken;
  challengeId: string;
  transactionType: SensitiveTransactionType;
  policyGeneration: number;
  policy: object;
  onAuthorized: () => void | Promise<void>;
}

interface GateLease {
  ownerToken: MFAGateOwnerToken;
  policyGeneration: number;
  policy: object;
}

export function useMFAGate() {
  const { policy, policyStatus, policyGeneration, loadMFAData, createChallenge } = useMFAStore();

  const [open, setOpen] = useState(false);
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [transactionType, setTransactionType] = useState<SensitiveTransactionType | undefined>();
  const [actionDescription, setActionDescription] = useState<string | undefined>();
  const pendingActionRef = useRef<PendingAction | null>(null);
  const gateLeaseRef = useRef<GateLease | null>(null);
  const mountedRef = useRef(false);

  const releaseOwner = useCallback((ownerToken: MFAGateOwnerToken | null) => {
    if (!ownerToken) return;
    const store = useMFAStore.getState();
    if (store.ownsGate(ownerToken)) {
      store.releaseGateOwner(ownerToken);
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (policyStatus === 'idle') {
      void loadMFAData();
    }
  }, [policyStatus, loadMFAData]);

  useEffect(() => {
    const lease = gateLeaseRef.current;
    if (
      lease &&
      (policyStatus !== 'ready' ||
        lease.policyGeneration !== policyGeneration ||
        lease.policy !== policy)
    ) {
      pendingActionRef.current = null;
      gateLeaseRef.current = null;
      releaseOwner(lease.ownerToken);
      setOpen(false);
      setIsPending(false);
      setError(POLICY_ERROR);
    }
  }, [policy, policyGeneration, policyStatus, releaseOwner]);

  useEffect(() => {
    return () => {
      const lease = gateLeaseRef.current;
      pendingActionRef.current = null;
      gateLeaseRef.current = null;
      releaseOwner(lease?.ownerToken ?? null);
    };
  }, [releaseOwner]);

  const requireCurrentPolicy = useCallback(async () => {
    let current = useMFAStore.getState();
    if (current.policyStatus !== 'ready') {
      if (current.policyStatus === 'idle' || current.policyStatus === 'error') {
        await current.loadMFAData();
      } else if (current.policyStatus === 'loading') {
        await current.loadMFAData();
      }
      current = useMFAStore.getState();
    }
    if (current.policyStatus !== 'ready' || !hasValidSensitiveTransactions(current.policy)) {
      throw new MFAAuthorizationError('policy_unavailable');
    }
    return current;
  }, []);

  const clearError = useCallback(() => setError(null), []);

  const gateAction = useCallback(
    async ({ transactionType: type, actionDescription, onAuthorized }: MFAGateActionOptions) => {
      let ownerToken: MFAGateOwnerToken | null = null;
      setError(null);
      setIsPending(true);
      try {
        const current = await requireCurrentPolicy();
        if (!current.policy.sensitiveTransactions.includes(type)) {
          await onAuthorized();
          if (mountedRef.current) setIsPending(false);
          return;
        }

        ownerToken = useMFAStore.getState().acquireGateOwner();
        if (!ownerToken) throw new MFAAuthorizationError('authorization_pending');
        gateLeaseRef.current = {
          ownerToken,
          policyGeneration: current.policyGeneration,
          policy: current.policy,
        };

        const challenge = await createChallenge(ownerToken, type);
        const pending: PendingAction = Object.freeze({
          ownerToken,
          challengeId: challenge.id,
          transactionType: type,
          policyGeneration: current.policyGeneration,
          policy: current.policy,
          onAuthorized,
        });
        pendingActionRef.current = pending;
        setTransactionType(type);
        setActionDescription(actionDescription);

        const latest = useMFAStore.getState();
        if (
          !mountedRef.current ||
          pendingActionRef.current !== pending ||
          !latest.ownsGate(ownerToken, challenge.id) ||
          latest.policyStatus !== 'ready' ||
          latest.policyGeneration !== pending.policyGeneration ||
          latest.policy !== pending.policy
        ) {
          pendingActionRef.current = null;
          gateLeaseRef.current = null;
          releaseOwner(ownerToken);
          if (mountedRef.current) setIsPending(false);
          return;
        }
        setOpen(true);
      } catch (err) {
        const ownsFailedAttempt =
          ownerToken !== null && gateLeaseRef.current?.ownerToken === ownerToken;
        if (ownsFailedAttempt) {
          pendingActionRef.current = null;
          gateLeaseRef.current = null;
          releaseOwner(ownerToken);
        }
        if (mountedRef.current) {
          setIsPending(gateLeaseRef.current !== null);
          setError(authorizationErrorMessage(err));
        }
        throw err;
      }
    },
    [createChallenge, releaseOwner, requireCurrentPolicy]
  );

  const pending = pendingActionRef.current;
  const handleSuccess = useCallback(async () => {
    if (!pending || pendingActionRef.current !== pending || !mountedRef.current) return;

    const current = useMFAStore.getState();
    if (
      !current.ownsGate(pending.ownerToken, pending.challengeId) ||
      current.policyStatus !== 'ready' ||
      current.policyGeneration !== pending.policyGeneration ||
      current.policy !== pending.policy ||
      !hasValidSensitiveTransactions(current.policy) ||
      !current.policy.sensitiveTransactions.includes(pending.transactionType)
    ) {
      return;
    }

    pendingActionRef.current = null;
    gateLeaseRef.current = null;
    current.releaseGateOwner(pending.ownerToken);
    setOpen(false);
    try {
      await pending.onAuthorized();
    } catch {
      if (mountedRef.current) setError(AUTHORIZATION_ERROR);
    } finally {
      if (mountedRef.current) setIsPending(false);
    }
  }, [pending]);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!nextOpen) {
        const pending = pendingActionRef.current;
        const lease = gateLeaseRef.current;
        pendingActionRef.current = null;
        gateLeaseRef.current = null;
        releaseOwner(pending?.ownerToken ?? lease?.ownerToken ?? null);
        setIsPending(false);
      }
      setOpen(nextOpen);
    },
    [releaseOwner]
  );

  return {
    gateAction,
    isPending,
    error,
    clearError,
    challengeProps: {
      open,
      onOpenChange: handleOpenChange,
      transactionType,
      actionDescription,
      onSuccess: handleSuccess,
      onFailure: () => undefined,
    },
  };
}
