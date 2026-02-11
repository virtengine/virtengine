/**
 * HPC Hooks
 *
 * React hooks for HPC feature interactions.
 */

import { useCallback, useEffect, useMemo, useState } from 'react';
import type { VirtEngineClient } from '@virtengine/chain-sdk';
import { useChainQuery } from '@/hooks/useChainQuery';
import { collectWorkloadTemplates, createHPCClient, mapSdkJob } from '../lib/hpc-client';
import type { JobOutput, JobStatus } from '../types';

/**
 * Hook to fetch and manage workload templates
 */
export function useWorkloadTemplates() {
  const query = useCallback(async (client: VirtEngineClient) => {
    const offerings = await client.hpc.listOfferings({ activeOnly: true });
    return collectWorkloadTemplates(offerings);
  }, []);

  const { data, isLoading, error, refetch } = useChainQuery(query, []);

  return { templates: data ?? [], isLoading, error, refetch };
}

/**
 * Hook to fetch a single template
 */
export function useWorkloadTemplate(templateId: string | null) {
  const query = useCallback(
    async (client: VirtEngineClient) => {
      if (!templateId) return null;
      const offerings = await client.hpc.listOfferings({ activeOnly: true });
      const templates = collectWorkloadTemplates(offerings);
      return templates.find((template) => template.id === templateId) ?? null;
    },
    [templateId]
  );

  const { data, isLoading, error, refetch } = useChainQuery(query, [templateId]);

  return { template: data, isLoading: templateId ? isLoading : false, error, refetch };
}

/**
 * Hook to fetch and manage jobs
 */
export function useJobs(filters?: { status?: JobStatus[] }) {
  const statusFilter = useMemo(() => filters?.status ?? [], [filters?.status]);
  const statusKey = useMemo(() => statusFilter.join(','), [statusFilter]);

  const query = useCallback(
    async (client: VirtEngineClient) => {
      const jobs = await client.hpc.listJobs();
      const mapped = jobs.map(mapSdkJob);

      if (!statusFilter.length) return mapped;
      return mapped.filter((job) => statusFilter.includes(job.status));
    },
    [statusFilter]
  );

  const { data, isLoading, error, refetch } = useChainQuery(query, [statusKey]);

  return { jobs: data ?? [], isLoading, error, refetch };
}

/**
 * Hook to fetch a single job with auto-refresh
 */
export function useJob(jobId: string | null, autoRefresh = true) {
  const query = useCallback(
    async (client: VirtEngineClient) => {
      if (!jobId) return null;
      const job = await client.hpc.getJob(jobId);
      return job ? mapSdkJob(job) : null;
    },
    [jobId]
  );

  const { data, isLoading, error, refetch } = useChainQuery(query, [jobId]);

  useEffect(() => {
    if (!jobId || !autoRefresh) return;
    const interval = setInterval(() => void refetch(), 10000);
    return () => clearInterval(interval);
  }, [jobId, autoRefresh, refetch]);

  return { job: data, isLoading: jobId ? isLoading : false, error };
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

/**
 * Hook for streaming job logs with auto-refresh
 */
export function useJobLogs(jobId: string | null, autoRefresh = true) {
  const [lines, setLines] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setLines([]);
      setIsLoading(false);
      return;
    }

    const client = createHPCClient();

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
  }, [jobId, autoRefresh]);

  return { lines, isLoading, error };
}

/**
 * Hook for job outputs
 */
export function useJobOutputs(jobId: string | null) {
  const [outputs, setOutputs] = useState<JobOutput[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!jobId) {
      setOutputs([]);
      setIsLoading(false);
      return;
    }

    const client = createHPCClient();

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
  }, [jobId]);

  return { outputs, isLoading, error };
}

/**
 * Hook for job resource usage with auto-refresh
 */
export function useJobUsage(jobId: string | null, autoRefresh = true) {
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

    const client = createHPCClient();

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
  }, [jobId, autoRefresh]);

  return { usage, isLoading, error };
}
