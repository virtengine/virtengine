'use client';

/**
 * HPC Job Detail Client Component
 *
 * Displays job details including status, logs, outputs, and resource usage.
 */

import Link from 'next/link';
import { useParams } from 'next/navigation';
import { useCallback, useRef, useEffect, useState, type ReactNode } from 'react';
import { useJob, useJobLogs, useJobOutputs, useJobUsage, useJobCancellation } from '@/features/hpc';
import type { JobStatus } from '@/features/hpc';
import { txLink } from '@/lib/explorer';

export default function JobDetailClient() {
  const params = useParams();
  const jobId = params.jobId as string;

  const { job, isLoading, error } = useJob(jobId);
  const isRunning = job?.status === 'running' || job?.status === 'queued';
  const { lines: logLines, isLoading: logsLoading } = useJobLogs(jobId, isRunning);
  const { outputs, isLoading: outputsLoading } = useJobOutputs(jobId);
  const { usage } = useJobUsage(jobId, isRunning);
  const { cancelJob, isCancelling } = useJobCancellation();
  const [activeTab, setActiveTab] = useState<'logs' | 'outputs' | 'usage'>('logs');

  const handleCancel = useCallback(async () => {
    if (!jobId) return;
    try {
      await cancelJob(jobId);
    } catch {
      // Error handled by hook
    }
  }, [jobId, cancelJob]);

  if (isLoading) {
    return (
      <div className="container py-8">
        <div className="mb-6">
          <div className="h-4 w-24 animate-pulse rounded bg-muted" />
        </div>
        <div className="mb-8 space-y-2">
          <div className="h-8 w-64 animate-pulse rounded bg-muted" />
          <div className="h-4 w-96 animate-pulse rounded bg-muted" />
        </div>
        <div className="h-96 animate-pulse rounded-lg bg-muted" />
      </div>
    );
  }

  if (error || !job) {
    return (
      <div className="container py-8">
        <div className="mb-6">
          <Link href="/hpc/jobs" className="text-sm text-muted-foreground hover:text-foreground">
            ← Back to Jobs
          </Link>
        </div>
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">{error?.message ?? 'Job not found'}</p>
        </div>
      </div>
    );
  }

  const statusConfig: Record<JobStatus, { label: string; color: string; bg: string }> = {
    running: {
      label: 'Running',
      color: 'text-blue-700 dark:text-blue-400',
      bg: 'bg-blue-100 dark:bg-blue-900/30',
    },
    queued: {
      label: 'Queued',
      color: 'text-yellow-700 dark:text-yellow-400',
      bg: 'bg-yellow-100 dark:bg-yellow-900/30',
    },
    pending: {
      label: 'Pending',
      color: 'text-yellow-700 dark:text-yellow-400',
      bg: 'bg-yellow-100 dark:bg-yellow-900/30',
    },
    completing: {
      label: 'Completing',
      color: 'text-blue-700 dark:text-blue-400',
      bg: 'bg-blue-100 dark:bg-blue-900/30',
    },
    completed: {
      label: 'Completed',
      color: 'text-green-700 dark:text-green-400',
      bg: 'bg-green-100 dark:bg-green-900/30',
    },
    failed: {
      label: 'Failed',
      color: 'text-red-700 dark:text-red-400',
      bg: 'bg-red-100 dark:bg-red-900/30',
    },
    cancelled: {
      label: 'Cancelled',
      color: 'text-gray-700 dark:text-gray-400',
      bg: 'bg-gray-100 dark:bg-gray-800',
    },
    timeout: {
      label: 'Timeout',
      color: 'text-orange-700 dark:text-orange-400',
      bg: 'bg-orange-100 dark:bg-orange-900/30',
    },
  };

  const config = statusConfig[job.status];

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

  return (
    <div className="container py-8">
      {/* Breadcrumb */}
      <div className="mb-6">
        <Link href="/hpc/jobs" className="text-sm text-muted-foreground hover:text-foreground">
          ← Back to Jobs
        </Link>
      </div>

      {/* Header */}
      <div className="mb-8 flex items-start justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-3xl font-bold">{job.name}</h1>
            <span
              className={`rounded-full px-3 py-1 text-xs font-medium ${config.color} ${config.bg}`}
            >
              {config.label}
            </span>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            {job.templateId ? `Template: ${job.templateId}` : 'Custom workload'} • Created{' '}
            {new Date(job.createdAt).toLocaleString()}
          </p>
        </div>
        {(job.status === 'running' || job.status === 'queued') && (
          <button
            type="button"
            onClick={handleCancel}
            disabled={isCancelling}
            className="rounded-lg border border-destructive px-4 py-2 text-sm text-destructive hover:bg-destructive/10 disabled:opacity-50"
          >
            {isCancelling ? 'Cancelling...' : 'Cancel Job'}
          </button>
        )}
      </div>

      {/* Main Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left Column */}
        <div className="space-y-6 lg:col-span-2">
          {/* Progress */}
          {(job.status === 'running' || job.status === 'completed') && (
            <div className="rounded-lg border border-border bg-card p-6">
              <div className="flex justify-between text-sm">
                <span className="font-medium">Progress</span>
                <span>{progress}%</span>
              </div>
              <div className="mt-2 h-2 rounded-full bg-muted">
                <div
                  className={`h-full rounded-full transition-all ${
                    job.status === 'completed' ? 'bg-green-500' : 'bg-primary'
                  }`}
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          )}

          {/* Tab Content */}
          <div>
            <div className="flex gap-4 border-b border-border">
              <TabButton
                label="Logs"
                active={activeTab === 'logs'}
                onClick={() => setActiveTab('logs')}
              />
              <TabButton
                label="Outputs"
                active={activeTab === 'outputs'}
                onClick={() => setActiveTab('outputs')}
                badge={outputs.length > 0 ? outputs.length : undefined}
              />
              <TabButton
                label="Usage"
                active={activeTab === 'usage'}
                onClick={() => setActiveTab('usage')}
              />
            </div>

            <div className="mt-4">
              {activeTab === 'logs' && (
                <LogViewer lines={logLines} isLoading={logsLoading} isStreaming={isRunning} />
              )}
              {activeTab === 'outputs' && (
                <OutputsList outputs={outputs} isLoading={outputsLoading} jobStatus={job.status} />
              )}
              {activeTab === 'usage' && (
                <UsageMetrics usage={usage} resources={job.resources} jobStatus={job.status} />
              )}
            </div>
          </div>
        </div>

        {/* Right Column */}
        <div className="space-y-6">
          {/* Resources */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h3 className="font-semibold">Resources</h3>
            <div className="mt-4 space-y-3 text-sm">
              <InfoRow label="Nodes" value={String(job.resources.nodes)} />
              <InfoRow label="CPUs / Node" value={String(job.resources.cpusPerNode)} />
              <InfoRow label="Memory / Node" value={`${job.resources.memoryGBPerNode} GB`} />
              <InfoRow label="Storage" value={`${job.resources.storageGB} GB`} />
              {job.resources.gpusPerNode && (
                <InfoRow
                  label="GPUs / Node"
                  value={`${job.resources.gpusPerNode} × ${job.resources.gpuType ?? 'GPU'}`}
                />
              )}
              <InfoRow
                label="Max Runtime"
                value={formatDuration(job.resources.maxRuntimeSeconds)}
              />
            </div>
          </div>

          {/* Cost */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h3 className="font-semibold">Cost</h3>
            <div className="mt-4 space-y-3 text-sm">
              <InfoRow label="Total Cost" value={`$${job.totalCost}`} />
              <InfoRow label="Deposit" value={`$${job.depositAmount}`} />
              <InfoRow label="Deposit Status" value={job.depositStatus} />
              <InfoRow
                label="Tx Hash"
                value={
                  <a
                    className="font-medium text-primary hover:underline"
                    href={txLink(job.txHash)}
                    rel="noopener noreferrer"
                    target="_blank"
                  >
                    {truncate(job.txHash, 16)}
                  </a>
                }
              />
            </div>
          </div>

          {/* Timeline */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h3 className="font-semibold">Timeline</h3>
            <div className="mt-4 space-y-3 text-sm">
              <InfoRow label="Created" value={new Date(job.createdAt).toLocaleString()} />
              {job.startedAt && (
                <InfoRow label="Started" value={new Date(job.startedAt).toLocaleString()} />
              )}
              {job.completedAt && (
                <InfoRow label="Completed" value={new Date(job.completedAt).toLocaleString()} />
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function TabButton({
  label,
  active,
  onClick,
  badge,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
  badge?: number;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-2 text-sm font-medium ${
        active
          ? 'border-b-2 border-primary text-primary'
          : 'text-muted-foreground hover:text-foreground'
      }`}
    >
      {label}
      {badge !== undefined && (
        <span className="rounded-full bg-muted px-2 py-0.5 text-xs">{badge}</span>
      )}
    </button>
  );
}

function LogViewer({
  lines,
  isLoading,
  isStreaming,
}: {
  lines: string[];
  isLoading: boolean;
  isStreaming: boolean;
}) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [lines]);

  if (isLoading) {
    return <div className="h-96 animate-pulse rounded-lg bg-muted" />;
  }

  if (lines.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">
          {isStreaming ? 'Waiting for log output...' : 'No logs available'}
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-border bg-gray-950">
      {isStreaming && (
        <div className="flex items-center gap-2 border-b border-gray-800 px-4 py-2 text-xs text-green-400">
          <span className="h-2 w-2 animate-pulse rounded-full bg-green-400" />
          Live streaming
        </div>
      )}
      <div
        ref={containerRef}
        className="h-96 overflow-auto p-4 font-mono text-xs leading-5 text-gray-300"
      >
        {lines.map((line) => (
          <div
            key={line}
            className={`whitespace-pre-wrap ${
              line.includes('ERROR') || line.includes('FATAL')
                ? 'text-red-400'
                : line.includes('WARN')
                  ? 'text-yellow-400'
                  : ''
            }`}
          >
            {line}
          </div>
        ))}
      </div>
    </div>
  );
}

function OutputsList({
  outputs,
  isLoading,
  jobStatus,
}: {
  outputs: {
    refId: string;
    name: string;
    type: string;
    accessUrl: string;
    sizeBytes: number;
    mimeType: string;
  }[];
  isLoading: boolean;
  jobStatus: JobStatus;
}) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-16 animate-pulse rounded-lg bg-muted" />
        ))}
      </div>
    );
  }

  if (outputs.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">
          {jobStatus === 'completed'
            ? 'No output files available'
            : 'Outputs will appear here when the job completes'}
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {outputs.map((output) => (
        <div
          key={output.refId}
          className="flex items-center justify-between rounded-lg border border-border bg-card p-4"
        >
          <div className="flex items-center gap-3">
            <FileIcon type={output.type} />
            <div>
              <p className="font-medium">{output.name}</p>
              <p className="text-xs text-muted-foreground">
                {formatBytes(output.sizeBytes)} • {output.type}
              </p>
            </div>
          </div>
          <a
            href={output.accessUrl}
            className="rounded-lg border border-border px-3 py-1 text-sm hover:bg-accent"
            download={output.name}
          >
            Download
          </a>
        </div>
      ))}
    </div>
  );
}

function UsageMetrics({
  usage,
  resources,
  jobStatus,
}: {
  usage: {
    cpuPercent: number;
    memoryPercent: number;
    gpuPercent?: number;
    elapsedSeconds: number;
    estimatedRemainingSeconds?: number;
  } | null;
  resources: {
    nodes: number;
    cpusPerNode: number;
    memoryGBPerNode: number;
    gpusPerNode?: number;
    storageGB: number;
    maxRuntimeSeconds: number;
  };
  jobStatus: JobStatus;
}) {
  if (!usage || jobStatus === 'pending' || jobStatus === 'queued') {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">
          Usage metrics will appear when the job starts running
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Resource Utilization */}
      <div className="rounded-lg border border-border bg-card p-6">
        <h4 className="mb-4 font-semibold">Resource Utilization</h4>
        <div className="space-y-4">
          <UsageBar label="CPU" percent={usage.cpuPercent} color="bg-blue-500" />
          <UsageBar label="Memory" percent={usage.memoryPercent} color="bg-purple-500" />
          {usage.gpuPercent !== undefined && (
            <UsageBar label="GPU" percent={usage.gpuPercent} color="bg-green-500" />
          )}
        </div>
      </div>

      {/* Time */}
      <div className="rounded-lg border border-border bg-card p-6">
        <h4 className="mb-4 font-semibold">Runtime</h4>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="rounded-lg bg-muted/50 p-4 text-center">
            <div className="text-2xl font-bold">{formatDuration(usage.elapsedSeconds)}</div>
            <div className="mt-1 text-sm text-muted-foreground">Elapsed</div>
          </div>
          {usage.estimatedRemainingSeconds !== undefined && (
            <div className="rounded-lg bg-muted/50 p-4 text-center">
              <div className="text-2xl font-bold">
                {formatDuration(usage.estimatedRemainingSeconds)}
              </div>
              <div className="mt-1 text-sm text-muted-foreground">Remaining</div>
            </div>
          )}
        </div>
      </div>

      {/* Allocation Summary */}
      <div className="rounded-lg border border-border bg-card p-6">
        <h4 className="mb-4 font-semibold">Allocation</h4>
        <div className="grid gap-3 text-sm sm:grid-cols-2">
          <InfoRow label="Total CPUs" value={String(resources.nodes * resources.cpusPerNode)} />
          <InfoRow
            label="Total Memory"
            value={`${resources.nodes * resources.memoryGBPerNode} GB`}
          />
          <InfoRow label="Storage" value={`${resources.storageGB} GB`} />
          {resources.gpusPerNode && (
            <InfoRow label="Total GPUs" value={String(resources.nodes * resources.gpusPerNode)} />
          )}
        </div>
      </div>
    </div>
  );
}

function UsageBar({ label, percent, color }: { label: string; percent: number; color: string }) {
  return (
    <div>
      <div className="flex justify-between text-sm">
        <span>{label}</span>
        <span className="font-medium">{percent}%</span>
      </div>
      <div className="mt-1 h-2 rounded-full bg-muted">
        <div
          className={`h-full rounded-full transition-all ${color}`}
          style={{ width: `${percent}%` }}
        />
      </div>
    </div>
  );
}

function FileIcon({ type }: { type: string }) {
  const iconMap: Record<string, string> = {
    model: '🧠',
    logs: '📄',
    metrics: '📊',
    checkpoint: '💾',
    artifact: '📦',
    data: '📁',
  };
  return <span className="text-xl">{iconMap[type] ?? '📄'}</span>;
}

function InfoRow({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="flex justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  );
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function truncate(str: string, len: number): string {
  return str.length > len ? `${str.slice(0, len)}...` : str;
}
