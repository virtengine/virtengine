/**
 * MFA gating hook for sensitive actions.
 * Creates MFA challenges when policy requires verification.
 */

'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import type { SensitiveTransactionType } from '@/lib/portal-adapter';
import { useMFAStore } from './store';

interface MFAGateActionOptions {
  transactionType: SensitiveTransactionType;
  actionDescription?: string;
  onAuthorized: () => void | Promise<void>;
}

export function useMFAGate() {
  const { policy, isEnabled, isLoading, loadMFAData, createChallenge, clearChallenge } =
    useMFAStore();

  const [open, setOpen] = useState(false);
  const [transactionType, setTransactionType] = useState<SensitiveTransactionType | undefined>();
  const [actionDescription, setActionDescription] = useState<string | undefined>();
  const pendingActionRef = useRef<(() => void | Promise<void>) | null>(null);

  useEffect(() => {
    if (!policy && !isLoading) {
      void loadMFAData();
    }
  }, [policy, isLoading, loadMFAData]);

  const requiresMFA = useCallback(
    (type: SensitiveTransactionType) => {
      if (policy?.sensitiveTransactions) {
        return policy.sensitiveTransactions.includes(type);
      }
      return isEnabled;
    },
    [policy, isEnabled]
  );

  const gateAction = useCallback(
    async ({ transactionType: type, actionDescription, onAuthorized }: MFAGateActionOptions) => {
      if (!requiresMFA(type)) {
        await onAuthorized();
        return;
      }

      pendingActionRef.current = onAuthorized;
      setTransactionType(type);
      setActionDescription(actionDescription);

      try {
        await createChallenge(type);
        setOpen(true);
      } catch (err) {
        pendingActionRef.current = null;
        throw err;
      }
    },
    [createChallenge, requiresMFA]
  );

  const handleSuccess = useCallback(async () => {
    const action = pendingActionRef.current;
    pendingActionRef.current = null;
    clearChallenge();
    setOpen(false);
    if (action) {
      await action();
    }
  }, [clearChallenge]);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!nextOpen) {
        pendingActionRef.current = null;
        clearChallenge();
      }
      setOpen(nextOpen);
    },
    [clearChallenge]
  );

  return {
    gateAction,
    challengeProps: {
      open,
      onOpenChange: handleOpenChange,
      transactionType,
      actionDescription,
      onSuccess: handleSuccess,
    },
  };
}
