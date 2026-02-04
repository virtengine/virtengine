'use client';

import { useHPC, WorkloadLibrary } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

interface HPCWorkloadBrowserProps {
  className?: string;
  onTemplateSelect?: (templateId: string) => void;
}

/**
 * HPC Workload Browser Component
 * Browse and select from available workload templates
 */
export function HPCWorkloadBrowser({ className, onTemplateSelect }: HPCWorkloadBrowserProps) {
  const { state } = useHPC();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="h-32 w-full rounded bg-muted-foreground/20" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Workload Templates</CardTitle>
        <CardDescription>
          Choose from pre-configured workload templates or create a custom job
        </CardDescription>
      </CardHeader>
      <CardContent>
        <WorkloadLibrary
          templates={state.templates}
          onTemplateSelect={onTemplateSelect}
          showCategories={true}
          showSearch={true}
        />
      </CardContent>
    </Card>
  );
}
