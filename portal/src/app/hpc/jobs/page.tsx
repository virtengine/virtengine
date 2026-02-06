'use client';

/**
 * HPC Jobs Page
 *
 * List and manage HPC jobs
 */

import Link from 'next/link';
import { useState } from 'react';
import { JobList, JobStatistics } from '@/features/hpc';
import type { JobStatus } from '@/features/hpc';

export default function HPCJobsPage() {
  const [activeTab, setActiveTab] = useState<'all' | JobStatus>('all');

  const statusFilter = activeTab === 'all' ? undefined : [activeTab];

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
      <div className="mb-8">
        <JobStatistics />
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-4 border-b border-border">
        <TabButton
          label="All Jobs"
          active={activeTab === 'all'}
          onClick={() => setActiveTab('all')}
        />
        <TabButton
          label="Running"
          active={activeTab === 'running'}
          onClick={() => setActiveTab('running')}
        />
        <TabButton
          label="Queued"
          active={activeTab === 'queued'}
          onClick={() => setActiveTab('queued')}
        />
        <TabButton
          label="Completed"
          active={activeTab === 'completed'}
          onClick={() => setActiveTab('completed')}
        />
        <TabButton
          label="Failed"
          active={activeTab === 'failed'}
          onClick={() => setActiveTab('failed')}
        />
      </div>

      {/* Jobs List */}
      <JobList statusFilter={statusFilter} />
    </div>
  );
}

function TabButton({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium ${
        active
          ? 'border-b-2 border-primary text-primary'
          : 'text-muted-foreground hover:text-foreground'
      }`}
    >
      {label}
    </button>
  );
}
