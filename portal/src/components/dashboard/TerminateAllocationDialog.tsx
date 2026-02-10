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

  const handleConfirm = async () => {
    setIsTerminating(true);
    try {
      await onConfirm(allocation.id);
      onOpenChange(false);
    } finally {
      setIsTerminating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('Terminate Allocation')}</DialogTitle>
          <DialogDescription>
            {t('Are you sure you want to terminate {{offering}} on {{provider}}?', {
              offering: allocation.offeringName,
              provider: allocation.providerName,
            })}{' '}
            {t('This action cannot be undone and all provisioned resources will be released.')}
          </DialogDescription>
        </DialogHeader>
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          <p className="font-medium">{t('Warning')}</p>
          <p>
            {t(
              'Terminating this allocation will immediately stop all running workloads and release associated resources. Any unsaved data may be lost.'
            )}
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isTerminating}>
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
