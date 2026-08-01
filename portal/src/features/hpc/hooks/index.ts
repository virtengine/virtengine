/**
 * HPC Hooks
 *
 * React hooks for HPC feature interactions.
 * The default client reports typed unavailability until application wiring
 * injects authoritative adapters.
 */

import { useEffect, useState } from 'react';
import { useHPCClient } from '../context/HPCClientProvider';
import type { SubmitJobParams } from '../lib/hpc-client';
import type { Job, JobOutput, WorkloadTemplate, JobStatus } from '../types';

/**
 * Hook to fetch and manage workload templates
 */
export function useWorkloadTemplates() {
  const client = useHPCClient();
  const [templates, setTemplates] = useState<WorkloadTemplate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    client
      .listWorkloadTemplates()
      .then((data) => {
        setTemplates(data);
        setIsLoading(false);
      })
      .catch((err) => {
        setError(err as Error);
        setIsLoading(false);
      });
  }, [client]);

  return { templates, isLoading, error };
}

/**
 * Hook to fetch a single template
 */
export function useWorkloadTemplate(templateId: string | null) {
  const client = useHPCClient();
  const [template, setTemplate] = useState<WorkloadTemplate | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!templateId) {
      setTemplate(null);
      setIsLoading(false);
      return;
    }

    client
      .getWorkloadTemplate(templateId)
      .then((data) => {
        setTemplate(data);
        setIsLoading(false);
      })
      .catch((err) => {
        setError(err as Error);
        setIsLoading(false);
      });
  }, [client, templateId]);

  return { template, isLoading, error };
}

/**
 * Hook to fetch and manage jobs
 */
export function useJobs(filters?: { status?: JobStatus[] }) {
  const client = useHPCClient();
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const refetch = () => {
    setIsLoading(true);
    client
      .listJobs(filters)
      .then((data) => {
        setJobs(data);
        setIsLoading(false);
      })
      .catch((err) => {
        setError(err as Error);
        setIsLoading(false);
      });
  };

  useEffect(() => {
    refetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client, filters?.status?.join(',')]);

  return { jobs, isLoading, error, refetch };
}

/**
 * Hook to fetch a single job with auto-refresh
 */
export function useJob(jobId: string | null, autoRefresh = true) {
  const client = useHPCClient();
  const [job, setJob] = useState<Job | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setJob(null);
      setIsLoading(false);
      return;
    }

    const fetchJob = () => {
      client
        .getJob(jobId)
        .then((data) => {
          setJob(data);
          setIsLoading(false);
        })
        .catch((err) => {
          setError(err as Error);
          setIsLoading(false);
        });
    };

    fetchJob();

    // Auto-refresh every 10 seconds for running/queued jobs
    if (autoRefresh) {
      const interval = setInterval(fetchJob, 10000);
      return () => clearInterval(interval);
    }
  }, [client, jobId, autoRefresh]);

  return { job, isLoading, error };
}

/**
 * Hook for job submission
 */
export function useJobSubmission() {
  const client = useHPCClient();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const submitJob = async (params: SubmitJobParams) => {
    setIsSubmitting(true);
    setError(null);

    try {
      const result = await client.submitJob(params);
      setIsSubmitting(false);
      return result;
    } catch (err) {
      setError(err as Error);
      setIsSubmitting(false);
      throw err;
    }
  };

  return { submitJob, isSubmitting, error };
}

/**
 * Hook for job cancellation
 */
export function useJobCancellation() {
  const client = useHPCClient();
  const [isCancelling, setIsCancelling] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const cancelJob = async (jobId: string) => {
    setIsCancelling(true);
    setError(null);

    try {
      const result = await client.cancelJob(jobId);
      setIsCancelling(false);
      return result;
    } catch (err) {
      setError(err as Error);
      setIsCancelling(false);
      throw err;
    }
  };

  return { cancelJob, isCancelling, error };
}

/**
 * Hook for cost estimation
 */
export function useCostEstimation() {
  const client = useHPCClient();
  const [isEstimating, setIsEstimating] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const estimateCost = async (offeringId: string, resources: SubmitJobParams['resources']) => {
    setIsEstimating(true);
    setError(null);

    try {
      const result = await client.estimateJobCost(offeringId, resources);
      setIsEstimating(false);
      return result;
    } catch (err) {
      setError(err as Error);
      setIsEstimating(false);
      throw err;
    }
  };

  return { estimateCost, isEstimating, error };
}

/**
 * Hook for job statistics
 */
export function useJobStatistics() {
  const { jobs, isLoading, error } = useJobs();

  const stats = {
    running: jobs.filter((j) => j.status === 'running').length,
    queued: jobs.filter((j) => j.status === 'queued').length,
    completed: jobs.filter(
      (j) => j.status === 'completed' && j.completedAt && j.completedAt > Date.now() - 86400000
    ).length,
    failed: jobs.filter(
      (j) => j.status === 'failed' && j.completedAt && j.completedAt > Date.now() - 86400000
    ).length,
  };

  return { stats, isLoading, error };
}

/**
 * Hook for streaming job logs with auto-refresh
 */
export function useJobLogs(jobId: string | null, autoRefresh = true) {
  const client = useHPCClient();
  const [lines, setLines] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setLines([]);
      setIsLoading(false);
      return;
    }

    const fetchLogs = () => {
      client
        .getJobLogs(jobId, { tail: 200 })
        .then((data) => {
          setLines(data.lines);
          setIsLoading(false);
        })
        .catch((err) => {
          setError(err as Error);
          setIsLoading(false);
        });
    };

    fetchLogs();

    if (autoRefresh) {
      const interval = setInterval(fetchLogs, 5000);
      return () => clearInterval(interval);
    }
  }, [client, jobId, autoRefresh]);

  return { lines, isLoading, error };
}

/**
 * Hook for job outputs
 */
export function useJobOutputs(jobId: string | null) {
  const client = useHPCClient();
  const [outputs, setOutputs] = useState<JobOutput[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setOutputs([]);
      setIsLoading(false);
      return;
    }

    client
      .getJobOutputs(jobId)
      .then((data) => {
        setOutputs(data);
        setIsLoading(false);
      })
      .catch((err) => {
        setError(err as Error);
        setIsLoading(false);
      });
  }, [client, jobId]);

  return { outputs, isLoading, error };
}

/**
 * Hook for job resource usage with auto-refresh
 */
export function useJobUsage(jobId: string | null, autoRefresh = true) {
  const client = useHPCClient();
  const [usage, setUsage] = useState<{
    cpuPercent: number;
    memoryPercent: number;
    gpuPercent?: number;
    elapsedSeconds: number;
    estimatedRemainingSeconds?: number;
  } | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setUsage(null);
      setIsLoading(false);
      return;
    }

    const fetchUsage = () => {
      client
        .getJobUsage(jobId)
        .then((data) => {
          setUsage(data);
          setIsLoading(false);
        })
        .catch((err) => {
          setError(err as Error);
          setIsLoading(false);
        });
    };

    fetchUsage();

    if (autoRefresh) {
      const interval = setInterval(fetchUsage, 10000);
      return () => clearInterval(interval);
    }
  }, [client, jobId, autoRefresh]);

  return { usage, isLoading, error };
}
