/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * VEIDScore Component
 * Displays the VEID trust score (0-100) with tier badge and optional details.
 */

'use client';

import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/Badge';
import { TIER_INFO, FEATURE_THRESHOLDS } from '@/features/veid';
import type { IdentityTier } from '@/features/veid';

interface VEIDScoreProps {
  /** Score value (0-100) */
  score: number;
  /** Identity tier */
  tier: IdentityTier;
  /** Show feature access list */
  showFeatureAccess?: boolean;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function VEIDScore({
  score,
  tier,
  showFeatureAccess = false,
  size = 'md',
  className,
}: VEIDScoreProps) {
  const info = TIER_INFO[tier];

  const ringSize = size === 'sm' ? 'h-16 w-16' : size === 'lg' ? 'h-32 w-32' : 'h-24 w-24';
  const fontSize = size === 'sm' ? 'text-lg' : size === 'lg' ? 'text-4xl' : 'text-2xl';
  const subSize = size === 'sm' ? 'text-[8px]' : size === 'lg' ? 'text-sm' : 'text-[10px]';

  const strokeColor =
    score >= 80
      ? 'text-green-500'
      : score >= 60
        ? 'text-amber-500'
        : score >= 40
          ? 'text-orange-500'
          : score > 0
            ? 'text-red-500'
            : 'text-muted';

  return (
    <div className={cn('flex flex-col items-center gap-4', className)}>
      {/* Score ring */}
      <div className={cn('relative flex flex-shrink-0 items-center justify-center', ringSize)}>
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
            className={cn('transition-all duration-700', strokeColor)}
            strokeDasharray={`${(score / 100) * 264} 264`}
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className={cn('font-bold', fontSize)} aria-label={`VEID Score: ${score}`}>
            {score}
          </span>
          <span className={cn('text-muted-foreground', subSize)}>/ 100</span>
        </div>
      </div>

      {/* Tier badge */}
      <Badge className={cn(info.bgColor, info.color, 'border', info.borderColor)}>
        {info.icon} {info.label} Tier
      </Badge>

      {size !== 'sm' && (
        <p className="text-center text-sm text-muted-foreground">{info.description}</p>
      )}

      {/* Feature access list */}
      {showFeatureAccess && (
        <div className="w-full space-y-2">
          <h4 className="text-center text-sm font-semibold">Feature Access</h4>
          <div className="space-y-1">
            {FEATURE_THRESHOLDS.map((threshold) => {
              const met = score >= threshold.minScore;
              return (
                <div key={threshold.action} className="flex items-center gap-2 text-sm">
                  <span
                    className={met ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'}
                  >
                    {met ? '✓' : '○'}
                  </span>
                  <span className={cn('flex-1', !met && 'text-muted-foreground')}>
                    {threshold.label}
                  </span>
                  <span className="text-xs text-muted-foreground">{threshold.minScore}+</span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
