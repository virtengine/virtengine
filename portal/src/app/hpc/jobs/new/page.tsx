'use client';

/**
 * HPC Job Submission Page
 *
 * Wizard for submitting new HPC jobs
 */

import Link from 'next/link';
import { Suspense } from 'react';
import { JobWizard } from '@/features/hpc';

export default function NewHPCJobPage() {
  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link href="/hpc/jobs" className="text-sm text-muted-foreground hover:text-foreground">
          ← Back to Jobs
        </Link>
      </div>

      <div className="mb-8">
        <h1 className="text-3xl font-bold">Submit New Job</h1>
        <p className="mt-1 text-muted-foreground">Configure and submit a new HPC workload</p>
      </div>

      <Suspense fallback={<div className="text-sm text-muted-foreground">Loading job form...</div>}>
        <JobWizard />
      </Suspense>
    </div>
  );
}
