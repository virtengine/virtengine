import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'HPC Jobs',
  description: 'Submit and manage HPC jobs',
};

const jobPlaceholders = [
  {
    id: 'job-401',
    name: 'ML Training - ResNet50',
    template: 'PyTorch Training',
    status: 'Running' as const,
    progress: 67,
  },
  {
    id: 'job-402',
    name: 'Genome Analysis',
    template: 'Bioinformatics',
    status: 'Running' as const,
    progress: 23,
  },
  {
    id: 'job-403',
    name: 'CFD Simulation',
    template: 'OpenFOAM',
    status: 'Queued' as const,
    progress: 0,
  },
  {
    id: 'job-404',
    name: 'Protein Folding',
    template: 'AlphaFold',
    status: 'Running' as const,
    progress: 89,
  },
  {
    id: 'job-405',
    name: 'Weather Model',
    template: 'WRF Model',
    status: 'Completed' as const,
    progress: 100,
  },
  {
    id: 'job-406',
    name: 'Render Job #42',
    template: 'Blender Render',
    status: 'Completed' as const,
    progress: 100,
  },
  {
    id: 'job-407',
    name: 'Data Processing',
    template: 'Spark Cluster',
    status: 'Failed' as const,
    progress: 45,
  },
  {
    id: 'job-408',
    name: 'Neural Network',
    template: 'TensorFlow',
    status: 'Queued' as const,
    progress: 0,
  },
] as const;

export default function HPCJobsPage() {
  return (
    <div className="container py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">HPC Jobs</h1>
          <p className="mt-1 text-muted-foreground">
            Submit and track high-performance computing workloads
          </p>
        </div>
        <Link
          href="/hpc/jobs/new"
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Submit New Job
        </Link>
      </div>

      {/* Stats */}
      <div className="mb-8 grid gap-4 sm:grid-cols-4">
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Running</div>
          <div className="mt-1 text-2xl font-bold text-success">3</div>
        </div>
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Queued</div>
          <div className="mt-1 text-2xl font-bold text-warning">2</div>
        </div>
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Completed (24h)</div>
          <div className="mt-1 text-2xl font-bold">8</div>
        </div>
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Failed (24h)</div>
          <div className="mt-1 text-2xl font-bold text-destructive">1</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-4 border-b border-border">
        <button
          type="button"
          className="border-b-2 border-primary px-4 py-2 text-sm font-medium text-primary"
        >
          All Jobs
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Running
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Completed
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Failed
        </button>
      </div>

      {/* Jobs List */}
      <div className="space-y-4">
        {Array.from({ length: 8 }, (_, index) => ({ id: `job-${index + 1}`, index })).map(
          (job) => (
            <JobCard key={job.id} index={job.index} />
          )
        )}
      </div>
    </div>
  );
}

function JobCard({
  job,
}: {
  job: {
    id: string;
    name: string;
    template: string;
    status: 'Running' | 'Queued' | 'Completed' | 'Failed';
    progress: number;
  };
}) {
  const statusConfig = {
    Running: { dot: 'status-dot-success', text: 'text-success' },
    Queued: { dot: 'status-dot-pending', text: 'text-muted-foreground' },
    Completed: { dot: 'status-dot-success', text: 'text-success' },
    Failed: { dot: 'status-dot-error', text: 'text-destructive' },
  } as const;
  const config = statusConfig[job.status];

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="font-medium">{job.name}</h3>
          <p className="text-sm text-muted-foreground">{job.template}</p>
        </div>
        <span className={`flex items-center gap-2 text-sm ${config.text}`}>
          <span className={`status-dot ${config.dot}`} />
          {job.status}
        </span>
      </div>

      {(job.status === 'Running' || job.status === 'Failed') && (
        <div className="mt-4">
          <div className="flex justify-between text-sm">
            <span>Progress</span>
            <span>{job.progress}%</span>
          </div>
          <div className="mt-1 h-2 rounded-full bg-muted">
            <div
              className={`h-full rounded-full ${job.status === 'Failed' ? 'bg-destructive' : 'bg-primary'}`}
              style={{ width: `${job.progress}%` }}
            />
          </div>
        </div>
      )}

      <div className="mt-4 flex gap-2">
        <button
          type="button"
          className="rounded-lg border border-border px-3 py-1 text-sm hover:bg-accent"
        >
          View Details
        </button>
        {job.status === 'Running' && (
          <button
            type="button"
            className="rounded-lg border border-destructive px-3 py-1 text-sm text-destructive hover:bg-destructive/10"
          >
            Cancel
          </button>
        )}
        {job.status === 'Completed' && (
          <button
            type="button"
            className="rounded-lg border border-border px-3 py-1 text-sm hover:bg-accent"
          >
            Download Output
          </button>
        )}
        {job.status === 'Failed' && (
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
