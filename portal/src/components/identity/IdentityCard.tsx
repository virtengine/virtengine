'use client';

import { useIdentity, IdentityStatusCard, IdentityScoreDisplay } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';

interface IdentityCardProps {
  className?: string;
  showScore?: boolean;
  compact?: boolean;
}

/**
 * Identity Card Component
 * Displays user's identity verification status and score
 */
export function IdentityCard({ className, showScore = true, compact = false }: IdentityCardProps) {
  const { state } = useIdentity();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-4', className)}>
        <div className="h-4 w-24 rounded bg-muted-foreground/20" />
        <div className="mt-2 h-8 w-32 rounded bg-muted-foreground/20" />
      </div>
    );
  }

  return (
    <div className={cn('space-y-4', className)}>
      <IdentityStatusCard
        status={state.status}
        completedScopes={state.completedScopes}
        compact={compact}
      />
      {showScore && state.score && (
        <IdentityScoreDisplay
          score={state.score}
          showBreakdown={!compact}
        />
      )}
    </div>
  );
}
