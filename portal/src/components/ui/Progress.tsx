/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import * as React from 'react';
import * as ProgressPrimitive from '@radix-ui/react-progress';
import { cn } from '@/lib/utils';

export interface ProgressProps extends React.ComponentPropsWithoutRef<
  typeof ProgressPrimitive.Root
> {
  indicatorClassName?: string;
  showValue?: boolean;
  size?: 'sm' | 'default' | 'lg';
  variant?: 'default' | 'success' | 'warning' | 'destructive';
}

const Progress = React.forwardRef<React.ElementRef<typeof ProgressPrimitive.Root>, ProgressProps>(
  (
    {
      className,
      value,
      indicatorClassName,
      showValue,
      size = 'default',
      variant = 'default',
      ...props
    },
    ref
  ) => (
    <div className="relative w-full">
      <ProgressPrimitive.Root
        ref={ref}
        className={cn(
          'relative w-full overflow-hidden rounded-full bg-secondary',
          size === 'sm' && 'h-1.5',
          size === 'default' && 'h-3',
          size === 'lg' && 'h-4',
          className
        )}
        {...props}
      >
        <ProgressPrimitive.Indicator
          className={cn(
            'h-full w-full flex-1 transition-all',
            variant === 'default' && 'bg-primary',
            variant === 'success' && 'bg-success',
            variant === 'warning' && 'bg-warning',
            variant === 'destructive' && 'bg-destructive',
            indicatorClassName
          )}
          style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
        />
      </ProgressPrimitive.Root>
      {showValue && (
        <span
          className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-full pl-2 text-sm text-muted-foreground"
          aria-hidden="true"
        >
          {value}%
        </span>
      )}
    </div>
  )
);
Progress.displayName = ProgressPrimitive.Root.displayName;

// Circular progress variant
interface CircularProgressProps {
  value?: number;
  size?: number;
  strokeWidth?: number;
  className?: string;
  showValue?: boolean;
  variant?: 'default' | 'success' | 'warning' | 'destructive';
}

function CircularProgress({
  value = 0,
  size = 40,
  strokeWidth = 4,
  className,
  showValue,
  variant = 'default',
}: CircularProgressProps) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (value / 100) * circumference;

  const strokeColor = {
    default: 'stroke-primary',
    success: 'stroke-success',
    warning: 'stroke-warning',
    destructive: 'stroke-destructive',
  }[variant];

  return (
    <div className={cn('relative inline-flex', className)}>
      <svg
        width={size}
        height={size}
        className="-rotate-90 transform"
        aria-label={`Progress: ${value}%`}
        role="progressbar"
        aria-valuenow={value}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        <circle
          className="stroke-secondary"
          fill="none"
          strokeWidth={strokeWidth}
          r={radius}
          cx={size / 2}
          cy={size / 2}
        />
        <circle
          className={cn('transition-all duration-300 ease-in-out', strokeColor)}
          fill="none"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          r={radius}
          cx={size / 2}
          cy={size / 2}
          style={{
            strokeDasharray: circumference,
            strokeDashoffset: offset,
          }}
        />
      </svg>
      {showValue && (
        <span className="absolute inset-0 flex items-center justify-center text-xs font-medium">
          {value}%
        </span>
      )}
    </div>
  );
}

export { Progress, CircularProgress };
