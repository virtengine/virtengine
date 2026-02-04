'use client';

import { useHPC, JobOutputViewer, JobCancelDialog } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { useState } from 'react';

interface HPCJobDetailsProps {
  jobId: string;
  className?: string;
  onBack?: () => void;
}

/**
 * HPC Job Details Component
 * Shows detailed information about a specific job
 */
export function HPCJobDetails({ jobId, className, onBack }: HPCJobDetailsProps) {
  const { state } = useHPC();
  const [showCancelDialog, setShowCancelDialog] = useState(false);

  const job = state.jobs.find((j) => j.id === jobId);

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-96 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  if (!job) {
    return (
      <Card className={cn(className)}>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Job not found</p>
          {onBack && (
            <Button variant="link" onClick={onBack}>
              Go back to job list
            </Button>
          )}
        </CardContent>
      </Card>
    );
  }

  const statusColor = {
    pending: 'bg-yellow-500',
    queued: 'bg-yellow-500',
    running: 'bg-blue-500',
    completing: 'bg-blue-500',
    completed: 'bg-green-500',
    failed: 'bg-red-500',
    cancelled: 'bg-gray-500',
    timeout: 'bg-red-500',
  }[job.status] ?? 'bg-gray-500';

  return (
    <div className={cn('space-y-6', className)}>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Job: {job.name ?? job.id}</CardTitle>
              <CardDescription>
                Submitted {new Date(job.createdAt).toLocaleString()}
              </CardDescription>
            </div>
            <Badge className={cn(statusColor, 'text-white')}>
              {job.status}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Template</p>
              <p>{job.templateId ?? 'Custom'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Provider</p>
              <p className="font-mono text-sm">{job.providerAddress ?? 'Pending assignment'}</p>
            </div>
          </div>
          <div className="flex gap-2">
            {onBack && (
              <Button variant="outline" onClick={onBack}>
                Back to List
              </Button>
            )}
            {job.status === 'running' && (
              <Button variant="destructive" onClick={() => setShowCancelDialog(true)}>
                Cancel Job
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {(job.status === 'completed' || job.status === 'running') && (
        <Card>
          <CardHeader>
            <CardTitle>Job Output</CardTitle>
          </CardHeader>
          <CardContent>
            <JobOutputViewer
              jobId={jobId}
              outputs={[]}
              isRunning={job.status === 'running'}
            />
          </CardContent>
        </Card>
      )}

      <JobCancelDialog
        open={showCancelDialog}
        onOpenChange={setShowCancelDialog}
        jobId={jobId}
        jobName={job.name}
      />
    </div>
  );
}
