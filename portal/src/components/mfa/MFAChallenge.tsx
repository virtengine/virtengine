'use client';

import { useMFA, MFAPrompt } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/Modal';

interface MFAChallengeProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  onFailure?: (error: Error) => void;
  className?: string;
}

/**
 * MFA Challenge Component
 * Modal dialog for MFA verification during sensitive actions
 */
export function MFAChallenge({ open, onOpenChange, onSuccess, onFailure: _onFailure, className }: MFAChallengeProps) {
  const { state } = useMFA();

  const handleVerify = () => {
    onSuccess?.();
    onOpenChange(false);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={cn('sm:max-w-md', className)}>
        <DialogHeader>
          <DialogTitle>Verification Required</DialogTitle>
          <DialogDescription>
            Please complete two-factor authentication to continue
          </DialogDescription>
        </DialogHeader>
        <MFAPrompt
          factors={state.enrolledFactors}
          onVerify={handleVerify}
          onCancel={handleCancel}
        />
      </DialogContent>
    </Dialog>
  );
}
