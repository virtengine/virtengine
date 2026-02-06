/**
 * HPC Hooks
 *
 * React hooks for HPC feature interactions.
 * Uses mock client for now, will be replaced with real SDK integration.
 */

import { useEffect, useState } from 'react';
import { createHPCClient } from '../lib/hpc-client';
import type { Job, WorkloadTemplate, JobStatus } from '../types';

/**
 * Hook to fetch and manage workload templates
 */
export function useWorkloadTemplates() {
  const [templates, setTemplates] = useState<WorkloadTemplate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const client = createHPCClient();

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
  }, []);

  return { templates, isLoading, error };
}

/**
 * Hook to fetch a single template
 */
export function useWorkloadTemplate(templateId: string | null) {
  const [template, setTemplate] = useState<WorkloadTemplate | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!templateId) {
      setTemplate(null);
      setIsLoading(false);
      return;
    }

    const client = createHPCClient();

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
  }, [templateId]);

  return { template, isLoading, error };
}

/**
 * Hook to fetch and manage jobs
 */
export function useJobs(filters?: { status?: JobStatus[] }) {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const refetch = () => {
    setIsLoading(true);
    const client = createHPCClient();

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
  }, [filters?.status?.join(',')]);

  return { jobs, isLoading, error, refetch };
}

/**
 * Hook to fetch a single job with auto-refresh
 */
export function useJob(jobId: string | null, autoRefresh = true) {
  const [job, setJob] = useState<Job | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setJob(null);
      setIsLoading(false);
      return;
    }

    const client = createHPCClient();

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
  }, [jobId, autoRefresh]);

  return { job, isLoading, error };
}

/**
 * Hook for job submission
 */
export function useJobSubmission() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const submitJob = async (
    params: Parameters<ReturnType<typeof createHPCClient>['submitJob']>[0]
  ) => {
    setIsSubmitting(true);
    setError(null);

    try {
      const client = createHPCClient();
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
  const [isCancelling, setIsCancelling] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const cancelJob = async (jobId: string) => {
    setIsCancelling(true);
    setError(null);

    try {
      const client = createHPCClient();
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
  const [isEstimating, setIsEstimating] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const estimateCost = async (
    offeringId: string,
    resources: Parameters<ReturnType<typeof createHPCClient>['estimateJobCost']>[1]
  ) => {
    setIsEstimating(true);
    setError(null);

    try {
      const client = createHPCClient();
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
