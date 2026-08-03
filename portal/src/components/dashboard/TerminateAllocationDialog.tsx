/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/Button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import type { CustomerAllocation } from '@/types/customer';
import { useTranslation } from 'react-i18next';

interface TerminateAllocationDialogProps {
  allocation: CustomerAllocation;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (id: string) => Promise<void>;
}

export function TerminateAllocationDialog({
  allocation,
  open,
  onOpenChange,
  onConfirm,
}: TerminateAllocationDialogProps) {
  const { t } = useTranslation();
  const [isTerminating, setIsTerminating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleOpenChange = (nextOpen: boolean) => {
    if (isTerminating && !nextOpen) return;
    setError(null);
    onOpenChange(nextOpen);
  };

  const handleConfirm = async () => {
    setIsTerminating(true);
    setError(null);
    try {
      await onConfirm(allocation.id);
      handleOpenChange(false);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : t('Termination was not committed'));
    } finally {
      setIsTerminating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('Terminate Allocation')}</DialogTitle>
          <DialogDescription>
            {t('Are you sure you want to terminate {{offering}} on {{provider}}?', {
              offering: allocation.offeringName,
              provider: allocation.providerName,
            })}{' '}
            {t(
              'This action cannot be undone. The dialog closes only after the provider commits termination.'
            )}
          </DialogDescription>
        </DialogHeader>
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          <p className="font-medium">{t('Warning')}</p>
          <p>
            {t(
              'A committed termination stops workloads and releases associated resources. Until committed provider evidence is returned, the allocation remains active.'
            )}
          </p>
        </div>
        {error && (
          <div
            role="alert"
            className="rounded-md border border-destructive p-3 text-sm text-destructive"
          >
            {error}
          </div>
        )}
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isTerminating}
          >
            {t('Cancel')}
          </Button>
          <Button variant="destructive" onClick={handleConfirm} disabled={isTerminating}>
            {isTerminating ? t('Terminating…') : t('Terminate Allocation')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
