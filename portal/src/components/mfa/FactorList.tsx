/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Enrolled MFA factors list with enable/disable, primary toggle, and removal.
 */

'use client';

import { useState, useCallback } from 'react';
import { useMFAStore } from '@/features/mfa';
import type { MFAFactor } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Modal';

interface FactorListProps {
  className?: string;
  onAddFactor?: () => void;
}

const FACTOR_ICONS: Record<string, string> = {
  otp: '📱',
  fido2: '🔐',
  sms: '💬',
  email: '📧',
  biometric: '👤',
};

const FACTOR_LABELS: Record<string, string> = {
  otp: 'Authenticator App',
  fido2: 'Security Key',
  sms: 'SMS',
  email: 'Email',
  biometric: 'Biometric',
};

function getStatusBadgeVariant(status: string) {
  switch (status) {
    case 'active':
      return 'success' as const;
    case 'suspended':
      return 'warning' as const;
    case 'expired':
      return 'destructive' as const;
    default:
      return 'secondary' as const;
  }
}

function formatDate(timestamp: number | null): string {
  if (!timestamp) return 'Never';
  return new Date(timestamp).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function FactorList({ className, onAddFactor }: FactorListProps) {
  const { factors, isMutating, error, removeFactor, toggleFactor, setPrimaryFactor, clearError } =
    useMFAStore();

  const [removeTarget, setRemoveTarget] = useState<MFAFactor | null>(null);

  const handleToggle = useCallback(
    async (factor: MFAFactor) => {
      clearError();
      await toggleFactor(factor.id, factor.status !== 'active');
    },
    [toggleFactor, clearError]
  );

  const handleSetPrimary = useCallback(
    async (factorId: string) => {
      clearError();
      await setPrimaryFactor(factorId);
    },
    [setPrimaryFactor, clearError]
  );

  const handleRemoveConfirm = useCallback(async () => {
    if (!removeTarget) return;
    clearError();
    try {
      await removeFactor(removeTarget.id);
      setRemoveTarget(null);
    } catch {
      // error is set in store
    }
  }, [removeTarget, removeFactor, clearError]);

  if (factors.length === 0) {
    return (
      <div className={cn('rounded-lg border border-dashed bg-muted/20 p-8 text-center', className)}>
        <div className="mx-auto mb-4 text-4xl" aria-hidden="true">
          🔒
        </div>
        <h3 className="text-lg font-semibold">No MFA Factors Enrolled</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Add a second factor to protect your account with multi-factor authentication.
        </p>
        {onAddFactor && (
          <Button onClick={onAddFactor} className="mt-4">
            Add First Factor
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className={cn('space-y-3', className)}>
      {error && (
        <Alert variant="destructive" onClose={clearError}>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {factors.map((factor) => (
        <div
          key={factor.id}
          className="flex flex-col gap-3 rounded-lg border bg-card p-4 sm:flex-row sm:items-center sm:justify-between"
        >
          <div className="flex items-center gap-3">
            <span className="text-2xl" aria-hidden="true">
              {FACTOR_ICONS[factor.type] ?? '🔑'}
            </span>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-medium">
                  {factor.name || FACTOR_LABELS[factor.type] || factor.type}
                </span>
                <Badge variant={getStatusBadgeVariant(factor.status)} size="sm">
                  {factor.status}
                </Badge>
                {factor.isPrimary && (
                  <Badge variant="default" size="sm">
                    Primary
                  </Badge>
                )}
              </div>
              <p className="text-xs text-muted-foreground">
                {FACTOR_LABELS[factor.type] ?? factor.type} · Enrolled{' '}
                {formatDate(factor.enrolledAt ?? null)} · Last used{' '}
                {formatDate(factor.lastUsedAt ?? null)}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {!factor.isPrimary && factor.status === 'active' && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleSetPrimary(factor.id)}
                disabled={isMutating}
              >
                Set primary
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleToggle(factor)}
              disabled={isMutating}
            >
              {factor.status === 'active' ? 'Disable' : 'Enable'}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setRemoveTarget(factor)}
              disabled={isMutating}
              className="text-destructive hover:bg-destructive/10"
            >
              Remove
            </Button>
          </div>
        </div>
      ))}

      {onAddFactor && (
        <Button variant="outline" onClick={onAddFactor} className="w-full">
          + Add Another Factor
        </Button>
      )}

      {/* Removal confirmation dialog */}
      <Dialog open={!!removeTarget} onOpenChange={(open) => !open && setRemoveTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove MFA Factor</DialogTitle>
            <DialogDescription>
              Are you sure you want to remove{' '}
              <strong>
                {removeTarget?.name || FACTOR_LABELS[removeTarget?.type ?? ''] || 'this factor'}
              </strong>
              ? This action cannot be undone.
              {factors.length === 1 && (
                <>
                  {' '}
                  This is your only enrolled factor. Removing it will disable MFA on your account.
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRemoveTarget(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRemoveConfirm} loading={isMutating}>
              Remove Factor
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
