/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import * as React from 'react';
import { cn } from '@/lib/utils';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'circular' | 'rectangular';
  animation?: 'pulse' | 'shimmer' | 'none';
}

function Skeleton({
  className,
  variant = 'default',
  animation = 'pulse',
  ...props
}: SkeletonProps) {
  return (
    <div
      className={cn(
        'bg-muted',
        variant === 'default' && 'rounded-md',
        variant === 'circular' && 'rounded-full',
        variant === 'rectangular' && 'rounded-none',
        animation === 'pulse' && 'animate-pulse',
        animation === 'shimmer' && 'animate-shimmer bg-gradient-to-r from-muted via-muted-foreground/10 to-muted bg-[length:200%_100%]',
        className
      )}
      aria-hidden="true"
      {...props}
    />
  );
}

// Pre-built skeleton patterns
function SkeletonText({ lines = 3, className }: { lines?: number; className?: string }) {
  const lineKeys = Array.from({ length: lines }, (_, index) => `line-${index}`);
  return (
    <div className={cn('space-y-2', className)}>
      {lineKeys.map((key, i) => (
        <Skeleton
          key={key}
          className={cn(
            'h-4',
            i === lines - 1 && 'w-3/4'
          )}
        />
      ))}
    </div>
  );
}

function SkeletonCard({ className }: { className?: string }) {
  return (
    <div className={cn('rounded-lg border p-4 space-y-3', className)}>
      <Skeleton className="h-5 w-1/3" />
      <SkeletonText lines={2} />
      <div className="flex gap-2 pt-2">
        <Skeleton className="h-8 w-20" />
        <Skeleton className="h-8 w-20" />
      </div>
    </div>
  );
}

function SkeletonAvatar({ size = 'default' }: { size?: 'sm' | 'default' | 'lg' }) {
  return (
    <Skeleton
      variant="circular"
      className={cn(
        size === 'sm' && 'h-6 w-6',
        size === 'default' && 'h-10 w-10',
        size === 'lg' && 'h-14 w-14'
      )}
    />
  );
}

function SkeletonTable({ rows = 5, columns = 4 }: { rows?: number; columns?: number }) {
  const columnKeys = Array.from({ length: columns }, (_, index) => `col-${index}`);
  const rowKeys = Array.from({ length: rows }, (_, index) => `row-${index}`);
  return (
    <div className="w-full space-y-3">
      {/* Header */}
      <div className="flex gap-4">
        {columnKeys.map((key) => (
          <Skeleton key={key} className="h-4 flex-1" />
        ))}
      </div>
      {/* Rows */}
      {rowKeys.map((rowKey) => (
        <div key={rowKey} className="flex gap-4">
          {columnKeys.map((colKey) => (
            <Skeleton key={`${rowKey}-${colKey}`} className="h-10 flex-1" />
          ))}
        </div>
      ))}
    </div>
  );
}

export { Skeleton, SkeletonText, SkeletonCard, SkeletonAvatar, SkeletonTable };
