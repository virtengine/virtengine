'use client';

import { useIdentity, ScopeRequirements, RemediationGuide } from '@/lib/portal-adapter';
import type { MarketplaceAction } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface IdentityRequirementsProps {
  action: string;
  className?: string;
  onStartVerification?: () => void;
}

/**
 * Identity Requirements Component
 * Shows what verification is needed for a specific action
 */
export function IdentityRequirements({ action, className, onStartVerification }: IdentityRequirementsProps) {
  const { state, actions } = useIdentity();

  const gatingError = actions.checkRequirements(action as MarketplaceAction);

  if (!gatingError) {
    return null; // User meets requirements
  }

  return (
    <Card className={cn('border-amber-500/50', className)}>
      <CardHeader>
        <CardTitle className="text-amber-600 dark:text-amber-400">
          Verification Required
        </CardTitle>
        <CardDescription>
          Additional verification is needed to perform this action
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <ScopeRequirements
          action={action as MarketplaceAction}
          completedScopes={state.completedScopes}
        />
        <RemediationGuide
          remediation={gatingError.remediation}
          onStartStep={onStartVerification}
        />
      </CardContent>
    </Card>
  );
}
