/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Verification Status Component
 * Displays current verification status, progress, and score.
 */

'use client';

import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Skeleton } from '@/components/ui/Skeleton';
import { useVerificationStatus } from '@/features/veid';

interface VerificationStatusProps {
  /** Callback to start or resume verification */
  onStartVerification?: () => void;
  /** Callback to retry verification */
  onRetryVerification?: () => void;
  /** Show compact view */
  compact?: boolean;
  className?: string;
}

export function VerificationStatus({
  onStartVerification,
  onRetryVerification,
  compact = false,
  className,
}: VerificationStatusProps) {
  const {
    displayStatus,
    tierInfo,
    featureThresholds,
    completedScopeInfo,
    missingScopeInfo,
    currentScore,
    isLoading,
    hasError,
    refresh,
  } = useVerificationStatus();

  if (isLoading) {
    return (
      <Card className={cn(className)}>
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-60" />
        </CardHeader>
        <CardContent className="space-y-4">
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-4 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (hasError) {
    return (
      <Card className={cn(className)}>
        <CardContent className="py-6">
          <Alert variant="destructive">
            <AlertTitle>Error</AlertTitle>
            <AlertDescription className="flex items-center justify-between">
              <span>Failed to load verification status.</span>
              <Button variant="outline" size="sm" onClick={() => void refresh()}>
                Retry
              </Button>
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn(displayStatus.status === 'verified' && 'border-green-500/50', className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <span>{displayStatus.icon}</span>
              Identity Verification
            </CardTitle>
            <CardDescription>{displayStatus.description}</CardDescription>
          </div>
          <Badge
            variant={
              displayStatus.status === 'verified'
                ? 'success'
                : displayStatus.status === 'rejected'
                  ? 'destructive'
                  : 'secondary'
            }
          >
            {displayStatus.label}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Score Ring */}
        <div className="flex items-center gap-6">
          <div className="relative flex h-24 w-24 flex-shrink-0 items-center justify-center">
            <svg viewBox="0 0 100 100" className="-rotate-90" aria-hidden="true">
              <circle
                cx="50"
                cy="50"
                r="42"
                fill="none"
                stroke="currentColor"
                strokeWidth="8"
                className="text-muted"
              />
              <circle
                cx="50"
                cy="50"
                r="42"
                fill="none"
                stroke="currentColor"
                strokeWidth="8"
                strokeLinecap="round"
                className={cn(
                  'transition-all duration-700',
                  currentScore >= 80
                    ? 'text-green-500'
                    : currentScore >= 60
                      ? 'text-amber-500'
                      : currentScore >= 40
                        ? 'text-orange-500'
                        : 'text-red-500'
                )}
                strokeDasharray={`${(currentScore / 100) * 264} 264`}
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-2xl font-bold">{currentScore}</span>
              <span className="text-[10px] text-muted-foreground">/ 100</span>
            </div>
          </div>
          <div className="flex-1">
            <div className={cn('text-lg font-semibold', tierInfo.color)}>{tierInfo.label} Tier</div>
            <p className="text-sm text-muted-foreground">{tierInfo.description}</p>
          </div>
        </div>

        {/* Processing indicator */}
        {displayStatus.showProgress && (
          <div className="space-y-2">
            <div className="flex items-center gap-3 rounded-lg bg-muted p-3">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
              <span className="text-sm">Processing your verification...</span>
            </div>
          </div>
        )}

        {!compact && (
          <>
            {/* Completed scopes */}
            {completedScopeInfo.length > 0 && (
              <div>
                <h4 className="mb-2 text-sm font-semibold">Completed Verifications</h4>
                <div className="space-y-2">
                  {completedScopeInfo.map((scope) => (
                    <div
                      key={scope.type}
                      className="flex items-center justify-between rounded-lg border bg-green-50 p-3 dark:bg-green-950/20"
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-green-600 dark:text-green-400">✓</span>
                        <div>
                          <span className="text-sm font-medium">{scope.label}</span>
                          <p className="text-xs text-muted-foreground">{scope.description}</p>
                        </div>
                      </div>
                      <span className="text-xs text-muted-foreground">+{scope.points} pts</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Missing scopes */}
            {missingScopeInfo.length > 0 && (
              <div>
                <h4 className="mb-2 text-sm font-semibold">Remaining Verifications</h4>
                <div className="space-y-2">
                  {missingScopeInfo.map((scope) => (
                    <div
                      key={scope.type}
                      className="flex items-center justify-between rounded-lg border p-3"
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-muted-foreground">○</span>
                        <div>
                          <span className="text-sm font-medium">{scope.label}</span>
                          <p className="text-xs text-muted-foreground">{scope.description}</p>
                        </div>
                      </div>
                      <span className="text-xs text-muted-foreground">+{scope.points} pts</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Feature thresholds */}
            <div>
              <h4 className="mb-2 text-sm font-semibold">Feature Access</h4>
              <div className="space-y-2">
                {featureThresholds.map((threshold) => (
                  <div key={threshold.action} className="flex items-center gap-3">
                    <span
                      className={cn(
                        'text-sm',
                        threshold.met
                          ? 'text-green-600 dark:text-green-400'
                          : 'text-muted-foreground'
                      )}
                    >
                      {threshold.met ? '✓' : '○'}
                    </span>
                    <span
                      className={cn('flex-1 text-sm', !threshold.met && 'text-muted-foreground')}
                    >
                      {threshold.label}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {threshold.minScore}+ score
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}

        {/* Action buttons */}
        {displayStatus.status === 'unknown' && onStartVerification && (
          <Button size="lg" className="w-full" onClick={onStartVerification}>
            Start Verification
          </Button>
        )}
        {(displayStatus.status === 'rejected' || displayStatus.status === 'expired') &&
          onRetryVerification && (
            <Button size="lg" className="w-full" onClick={onRetryVerification}>
              {displayStatus.status === 'expired' ? 'Re-Verify Identity' : 'Try Again'}
            </Button>
          )}
      </CardContent>
    </Card>
  );
}
