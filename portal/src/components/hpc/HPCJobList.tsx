'use client';

import { useHPC, JobTracker } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

interface HPCJobListProps {
  className?: string;
  onJobSelect?: (jobId: string) => void;
}

/**
 * HPC Job List Component
 * Displays list of user's HPC jobs with status
 */
export function HPCJobList({ className, onJobSelect }: HPCJobListProps) {
  const { state } = useHPC();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 w-full rounded bg-muted-foreground/20" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Your HPC Jobs</CardTitle>
        <CardDescription>
          Monitor and manage your computing jobs
        </CardDescription>
      </CardHeader>
      <CardContent>
        <JobTracker
          jobs={state.jobs}
          onJobClick={onJobSelect}
          showFilters={true}
        />
      </CardContent>
    </Card>
  );
}
