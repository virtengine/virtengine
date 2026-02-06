'use client';

/**
 * HPC Templates Page
 *
 * Browse and select workload templates
 */

import { TemplateBrowser } from '@/features/hpc';

export default function HPCTemplatesPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Workload Templates</h1>
        <p className="mt-1 text-muted-foreground">
          Pre-configured templates for common HPC workloads
        </p>
      </div>

      <TemplateBrowser />
    </div>
  );
}
