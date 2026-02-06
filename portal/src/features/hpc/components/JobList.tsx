'use client';

/**
 * Job List Component
 *
 * Displays a list of HPC jobs with filtering and real-time updates.
 */

import Link from 'next/link';
import { useJobs } from '@/features/hpc';
import type { Job, JobStatus } from '@/features/hpc';

interface JobListProps {
  statusFilter?: JobStatus[];
}

export function JobList({ statusFilter }: JobListProps) {
  const { jobs, isLoading, error } = useJobs(statusFilter ? { status: statusFilter } : undefined);

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4 text-destructive">
        <p>Error loading jobs: {error.message}</p>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-32 animate-pulse rounded-lg bg-muted" />
        ))}
      </div>
    );
  }

  if (jobs.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">No jobs found</p>
        <Link
          href="/hpc/jobs/new"
          className="mt-4 inline-block text-sm text-primary hover:underline"
        >
          Submit your first job →
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {jobs.map((job) => (
        <JobCard key={job.id} job={job} />
      ))}
    </div>
  );
}

function JobCard({ job }: { job: Job }) {
  const statusConfig: Record<JobStatus, { dot: string; text: string; color: string }> = {
    running: { dot: 'status-dot-success', text: 'Running', color: 'text-success' },
    queued: { dot: 'status-dot-pending', text: 'Queued', color: 'text-muted-foreground' },
    pending: { dot: 'status-dot-pending', text: 'Pending', color: 'text-muted-foreground' },
    completing: { dot: 'status-dot-success', text: 'Completing', color: 'text-success' },
    completed: { dot: 'status-dot-success', text: 'Completed', color: 'text-success' },
    failed: { dot: 'status-dot-error', text: 'Failed', color: 'text-destructive' },
    cancelled: { dot: 'status-dot-error', text: 'Cancelled', color: 'text-muted-foreground' },
    timeout: { dot: 'status-dot-error', text: 'Timeout', color: 'text-destructive' },
  };

  const config = statusConfig[job.status];

  // Calculate progress for running jobs
  const progress =
    job.status === 'running' && job.startedAt
      ? Math.min(
          Math.floor(
            ((Date.now() - job.startedAt) / (job.resources.maxRuntimeSeconds * 1000)) * 100
          ),
          99
        )
      : job.status === 'completed'
        ? 100
        : 0;

  const timeAgo = formatTimeAgo(job.createdAt);

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <Link href={`/hpc/jobs/${job.id}`} className="hover:underline">
            <h3 className="font-medium">{job.name}</h3>
          </Link>
          <p className="mt-1 text-sm text-muted-foreground">
            {job.templateId ? `Template: ${job.templateId}` : 'Custom workload'} • Created {timeAgo}
          </p>
        </div>
        <span className={`flex items-center gap-2 text-sm ${config.color}`}>
          <span className={`status-dot ${config.dot}`} />
          {config.text}
        </span>
      </div>

      {/* Resource Info */}
      <div className="mt-3 flex gap-4 text-sm text-muted-foreground">
        <span>
          {job.resources.nodes} node{job.resources.nodes > 1 ? 's' : ''}
        </span>
        <span>{job.resources.cpusPerNode} CPUs</span>
        {job.resources.gpusPerNode && <span>{job.resources.gpusPerNode} GPUs</span>}
        <span>{job.resources.memoryGBPerNode} GB RAM</span>
      </div>

      {/* Progress for running/failed jobs */}
      {(job.status === 'running' || job.status === 'failed') && progress > 0 && (
        <div className="mt-4">
          <div className="flex justify-between text-sm">
            <span>Progress</span>
            <span>{progress}%</span>
          </div>
          <div className="mt-1 h-2 rounded-full bg-muted">
            <div
              className={`h-full rounded-full ${job.status === 'failed' ? 'bg-destructive' : 'bg-primary'}`}
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="mt-4 flex gap-2">
        <Link
          href={`/hpc/jobs/${job.id}`}
          className="rounded-lg border border-border px-3 py-1 text-sm hover:bg-accent"
        >
          View Details
        </Link>
        {job.status === 'running' && (
          <button
            type="button"
            className="rounded-lg border border-destructive px-3 py-1 text-sm text-destructive hover:bg-destructive/10"
          >
            Cancel
          </button>
        )}
        {job.status === 'completed' && (
          <button
            type="button"
            className="rounded-lg border border-border px-3 py-1 text-sm hover:bg-accent"
          >
            Download Output
          </button>
        )}
        {(job.status === 'failed' || job.status === 'completed') && (
          <button
            type="button"
            className="rounded-lg border border-border px-3 py-1 text-sm hover:bg-accent"
          >
            View Logs
          </button>
        )}
      </div>
    </div>
  );
}

function formatTimeAgo(timestamp: number): string {
  const seconds = Math.floor((Date.now() - timestamp) / 1000);

  if (seconds < 60) return 'just now';
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return `${Math.floor(seconds / 86400)}d ago`;
}
