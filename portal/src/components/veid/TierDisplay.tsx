/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Tier Display Component
 * Shows identity tier badge and score as a compact display.
 */

'use client';

import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/Badge';
import { TIER_INFO } from '@/features/veid';
import type { IdentityTier } from '@/features/veid';

interface TierDisplayProps {
  /** Score value (0-100) */
  score: number;
  /** Identity tier */
  tier: IdentityTier;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Show score number */
  showScore?: boolean;
  className?: string;
}

export function TierDisplay({
  score,
  tier,
  size = 'md',
  showScore = true,
  className,
}: TierDisplayProps) {
  const info = TIER_INFO[tier];

  const ringSize = size === 'sm' ? 'h-12 w-12' : size === 'lg' ? 'h-24 w-24' : 'h-16 w-16';
  const fontSize = size === 'sm' ? 'text-sm' : size === 'lg' ? 'text-3xl' : 'text-xl';
  const labelSize = size === 'sm' ? 'text-xs' : size === 'lg' ? 'text-base' : 'text-sm';

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
    <div className={cn('flex items-center gap-3', className)}>
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
        {showScore && (
          <span className={cn('absolute font-bold', fontSize)} aria-label={`Score: ${score}`}>
            {score}
          </span>
        )}
      </div>
      <div>
        <Badge className={cn(info.bgColor, info.color, 'border', info.borderColor)}>
          {info.icon} {info.label}
        </Badge>
        {size !== 'sm' && (
          <p className={cn('mt-1 text-muted-foreground', labelSize)}>{info.description}</p>
        )}
      </div>
    </div>
  );
}
