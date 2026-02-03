'use client';

import { useHPC, JobSubmissionForm, type JobSubmission } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

interface HPCJobSubmitProps {
  className?: string;
  templateId?: string;
  onSubmitSuccess?: (jobId: string) => void;
  onSubmitError?: (error: Error) => void;
}

/**
 * HPC Job Submit Component
 * Form for submitting new HPC jobs
 */
export function HPCJobSubmit({ className, templateId, onSubmitSuccess, onSubmitError }: HPCJobSubmitProps) {
  const { state } = useHPC();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-64 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Submit HPC Job</CardTitle>
        <CardDescription>
          Configure and submit a new high-performance computing job
        </CardDescription>
      </CardHeader>
      <CardContent>
        {state.error && (
          <Alert variant="destructive" className="mb-4">
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{state.error.message}</AlertDescription>
          </Alert>
        )}
        <JobSubmissionForm
          templates={state.templates}
          selectedTemplateId={templateId}
          onSubmit={async (submission: JobSubmission) => {
            try {
              // Submit would be handled by the hook
              onSubmitSuccess?.(submission.jobId ?? '');
            } catch (error) {
              onSubmitError?.(error as Error);
            }
          }}
        />
      </CardContent>
    </Card>
  );
}
