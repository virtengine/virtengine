'use client';

/**
 * Job Statistics Component
 *
 * Displays aggregate statistics for user's jobs.
 */

import { useJobStatistics } from '@/features/hpc';

export function JobStatistics() {
  const { stats, isLoading } = useJobStatistics();

  if (isLoading) {
    return (
      <div className="grid gap-4 sm:grid-cols-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-24 animate-pulse rounded-lg bg-muted" />
        ))}
      </div>
    );
  }

  return (
    <div className="grid gap-4 sm:grid-cols-4">
      <StatCard label="Running" value={stats.running} className="text-success" />
      <StatCard label="Queued" value={stats.queued} className="text-warning" />
      <StatCard label="Completed (24h)" value={stats.completed} />
      <StatCard label="Failed (24h)" value={stats.failed} className="text-destructive" />
    </div>
  );
}

function StatCard({
  label,
  value,
  className,
}: {
  label: string;
  value: number;
  className?: string;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="text-sm text-muted-foreground">{label}</div>
      <div className={`mt-1 text-2xl font-bold ${className ?? ''}`}>{value}</div>
    </div>
  );
}
