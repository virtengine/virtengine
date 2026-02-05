'use client';

import { useMFA, MFAEnrollmentWizard, type MFAFactorType } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

interface MFASetupProps {
  className?: string;
  onComplete?: () => void;
  onCancel?: () => void;
}

/**
 * MFA Setup Component
 * Guides users through MFA enrollment
 */
export function MFASetup({ className, onComplete, onCancel }: MFASetupProps) {
  const { state } = useMFA();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-32 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Set Up Two-Factor Authentication</CardTitle>
        <CardDescription>Add an extra layer of security to your account</CardDescription>
      </CardHeader>
      <CardContent>
        <MFAEnrollmentWizard
          allowedFactors={['totp', 'webauthn', 'sms'] as MFAFactorType[]}
          onComplete={onComplete}
          onCancel={onCancel}
        />
      </CardContent>
    </Card>
  );
}
