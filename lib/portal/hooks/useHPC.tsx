// @ts-nocheck
/**
 * useHPC Hook
 * VE-705: Supercomputer/HPC UI (job submission, library workloads, outputs)
 */

import { useState, useCallback, useEffect, useContext, createContext } from 'react';
import type { ReactNode } from 'react';
import type {
  HPCState,
  WorkloadTemplate,
  JobManifest,
  JobSubmission,
  Job,
  JobStatus,
  JobOutput,
  JobOutputReference,
  JobPriceQuote,
  JobValidationError,
} from '../types/hpc';
import { initialHPCState } from '../types/hpc';
import type { QueryClient, ChainEvent } from '../types/chain';
import { sanitizePlainText, sanitizeObject } from '../utils/security';

interface HPCContextValue {
  state: HPCState;
  actions: HPCActions;
}

interface HPCActions {
  refresh: () => Promise<void>;
  getWorkloadTemplates: () => Promise<void>;
  startJobSubmission: (templateId?: string) => void;
  updateJobManifest: (manifest: Partial<JobManifest>) => void;
  selectOffering: (offeringId: string) => void;
  getQuote: () => Promise<JobPriceQuote>;
  validateJob: () => JobValidationError[];
  submitJob: () => Promise<Job>;
  cancelSubmission: () => void;
  getJobs: () => Promise<void>;
  getJob: (jobId: string) => Promise<Job>;
  cancelJob: (jobId: string) => Promise<void>;
  getOutputs: (jobId: string) => Promise<JobOutputReference[]>;
  decryptOutput: (outputRef: JobOutputReference) => Promise<JobOutput>;
  subscribeToJob: (jobId: string, callback: (event: ChainEvent) => void) => () => void;
  clearError: () => void;
}

const HPCContext = createContext<HPCContextValue | null>(null);

export interface HPCProviderProps {
  children: ReactNode;
  queryClient: QueryClient;
  accountAddress: string | null;
  getAuthHeader: () => Promise<string>;
}

export function HPCProvider({
  children,
  queryClient,
  accountAddress,
  getAuthHeader,
}: HPCProviderProps) {
  const [state, setState] = useState<HPCState>(initialHPCState);

  const getWorkloadTemplates = useCallback(async () => {
    setState(prev => ({ ...prev, isLoading: true }));

    try {
      const templates: WorkloadTemplate[] = [
        {
          id: 'pytorch-training',
          name: 'PyTorch Training',
          description: 'Train PyTorch models on distributed GPU clusters',
          category: 'ml_training',
          defaultResources: {
            nodes: 1,
            cpusPerNode: 8,
            memoryGBPerNode: 32,
            gpusPerNode: 4,
            gpuType: 'nvidia-a100',
            maxRuntimeSeconds: 86400,
            storageGB: 100,
          },
          defaultParameters: {
            model: { name: 'model', type: 'string', description: 'Model name', required: true },
            epochs: { name: 'epochs', type: 'number', description: 'Training epochs', required: true, defaultValue: 10, min: 1, max: 1000 },
          },
          requiredIdentityScore: 50,
          mfaRequired: true,
          estimatedCostPerHour: '10.00',
          version: '1.0.0',
          iconUrl: '/icons/pytorch.png',
          docsUrl: 'https://docs.virtengine.com/hpc/pytorch',
        },
        {
          id: 'molecular-dynamics',
          name: 'Molecular Dynamics',
          description: 'Run GROMACS molecular dynamics simulations',
          category: 'simulation',
          defaultResources: {
            nodes: 4,
            cpusPerNode: 32,
            memoryGBPerNode: 128,
            maxRuntimeSeconds: 604800,
            storageGB: 500,
          },
          defaultParameters: {
            structure: { name: 'structure', type: 'file', description: 'PDB structure file', required: true },
            steps: { name: 'steps', type: 'number', description: 'Simulation steps', required: true, defaultValue: 1000000 },
          },
          requiredIdentityScore: 60,
          mfaRequired: true,
          estimatedCostPerHour: '25.00',
          version: '1.0.0',
        },
        {
          id: 'rendering',
          name: 'Blender Rendering',
          description: 'Render Blender scenes on GPU clusters',
          category: 'rendering',
          defaultResources: {
            nodes: 1,
            cpusPerNode: 16,
            memoryGBPerNode: 64,
            gpusPerNode: 2,
            gpuType: 'nvidia-a100',
            maxRuntimeSeconds: 43200,
            storageGB: 200,
          },
          defaultParameters: {
            scene: { name: 'scene', type: 'file', description: 'Blender scene file', required: true },
            startFrame: { name: 'startFrame', type: 'number', description: 'Start frame', required: true, defaultValue: 1 },
            endFrame: { name: 'endFrame', type: 'number', description: 'End frame', required: true, defaultValue: 250 },
          },
          requiredIdentityScore: 40,
          mfaRequired: false,
          estimatedCostPerHour: '15.00',
          version: '1.0.0',
        },
      ];

      setState(prev => ({
        ...prev,
        isLoading: false,
        workloadTemplates: templates,
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: {
          code: 'network_error',
          message: error instanceof Error ? error.message : 'Failed to load templates',
        },
      }));
    }
  }, []);

  const refresh = useCallback(async () => {
    await Promise.all([getWorkloadTemplates(), getJobs()]);
  }, [getWorkloadTemplates]);

  const startJobSubmission = useCallback((templateId?: string) => {
    const template = templateId 
      ? state.workloadTemplates.find(t => t.id === templateId) 
      : null;

    const defaultParams: Record<string, string | number | boolean> = {};
    if (template) {
      for (const [key, param] of Object.entries(template.defaultParameters)) {
        if (param.defaultValue !== undefined) {
          defaultParams[key] = param.defaultValue;
        }
      }
    }

    setState(prev => ({
      ...prev,
      submission: {
        step: templateId ? 'configure' : 'select_template',
        manifest: template ? {
          version: '1.0.0',
          name: '',
          templateId: template.id,
          resources: template.defaultResources,
          parameters: defaultParams,
        } : {
          version: '1.0.0',
          name: '',
          resources: {
            nodes: 1,
            cpusPerNode: 4,
            memoryGBPerNode: 16,
            maxRuntimeSeconds: 3600,
            storageGB: 50,
          },
          parameters: {},
        },
        selectedTemplate: template,
        selectedOffering: null,
        priceQuote: null,
        validationErrors: [],
        error: null,
      },
    }));
  }, [state.workloadTemplates]);

  const updateJobManifest = useCallback((manifestUpdate: Partial<JobManifest>) => {
    const sanitizedUpdate: Partial<JobManifest> = { ...manifestUpdate };

    if (typeof manifestUpdate.name === 'string') {
      sanitizedUpdate.name = sanitizePlainText(manifestUpdate.name, { maxLength: 120 });
    }

    if (typeof manifestUpdate.description === 'string') {
      sanitizedUpdate.description = sanitizePlainText(manifestUpdate.description, { maxLength: 500 });
    }

    if (typeof manifestUpdate.command === 'string') {
      sanitizedUpdate.command = sanitizePlainText(manifestUpdate.command, { maxLength: 300 });
    }

    if (typeof manifestUpdate.image === 'string') {
      sanitizedUpdate.image = sanitizePlainText(manifestUpdate.image, { maxLength: 200 });
    }

    if (manifestUpdate.environment) {
      sanitizedUpdate.environment = sanitizeObject(manifestUpdate.environment, {
        maxDepth: 2,
        maxKeyLength: 64,
        maxStringLength: 256,
        escapeHtmlStrings: false,
      }) as Record<string, string>;
    }

    if (manifestUpdate.parameters) {
      sanitizedUpdate.parameters = sanitizeObject(manifestUpdate.parameters, {
        maxDepth: 2,
        maxKeyLength: 64,
        maxStringLength: 256,
        escapeHtmlStrings: false,
      }) as Record<string, string | number | boolean>;
    }

    setState(prev => ({
      ...prev,
      submission: prev.submission ? {
        ...prev.submission,
        manifest: { ...prev.submission.manifest, ...sanitizedUpdate },
      } : null,
    }));
  }, []);

  const selectOffering = useCallback((offeringId: string) => {
    setState(prev => ({
      ...prev,
      submission: prev.submission ? {
        ...prev.submission,
        selectedOffering: offeringId,
        step: 'review',
      } : null,
    }));
  }, []);

  const getQuote = useCallback(async (): Promise<JobPriceQuote> => {
    if (!state.submission?.manifest.resources) {
      throw new Error('No job configured');
    }

    const resources = state.submission.manifest.resources;
    const hours = resources.maxRuntimeSeconds / 3600;
    const computeCost = resources.nodes * resources.cpusPerNode * hours * 0.1;
    const storageCost = resources.storageGB * 0.01;
    const gpuCost = (resources.gpusPerNode || 0) * hours * 2;
    const totalCost = computeCost + storageCost + gpuCost;

    const quote: JobPriceQuote = {
      estimatedTotal: totalCost.toFixed(2),
      depositRequired: (totalCost * 1.1).toFixed(2),
      breakdown: {
        compute: computeCost.toFixed(2),
        storage: storageCost.toFixed(2),
        network: '0.00',
        gpu: gpuCost > 0 ? gpuCost.toFixed(2) : undefined,
      },
      pricePerHour: (totalCost / hours).toFixed(2),
      maxHours: hours,
      denom: 'uve',
    };

    setState(prev => ({
      ...prev,
      submission: prev.submission ? { ...prev.submission, priceQuote: quote } : null,
    }));

    return quote;
  }, [state.submission?.manifest.resources]);

  const validateJob = useCallback((): JobValidationError[] => {
    const errors: JobValidationError[] = [];
    const manifest = state.submission?.manifest;

    if (!manifest?.name) {
      errors.push({ field: 'name', message: 'Job name is required' });
    }

    if (!manifest?.resources.maxRuntimeSeconds || manifest.resources.maxRuntimeSeconds < 60) {
      errors.push({ field: 'resources.maxRuntimeSeconds', message: 'Runtime must be at least 60 seconds' });
    }

    setState(prev => ({
      ...prev,
      submission: prev.submission ? { ...prev.submission, validationErrors: errors } : null,
    }));

    return errors;
  }, [state.submission?.manifest]);

  const submitJob = useCallback(async (): Promise<Job> => {
    const errors = validateJob();
    if (errors.length > 0) {
      throw new Error('Validation failed');
    }

    setState(prev => ({
      ...prev,
      submission: prev.submission ? { ...prev.submission, step: 'submit' } : null,
    }));

    const job: Job = {
      id: `job-${Date.now()}`,
      name: state.submission?.manifest.name || 'Untitled Job',
      customerAddress: accountAddress || '',
      providerAddress: '',
      offeringId: state.submission?.selectedOffering || '',
      templateId: state.submission?.manifest.templateId,
      status: 'pending',
      createdAt: Date.now(),
      resources: state.submission?.manifest.resources || {} as any,
      statusHistory: [],
      events: [],
      outputRefs: [],
      totalCost: state.submission?.priceQuote?.estimatedTotal || '0',
      depositAmount: state.submission?.priceQuote?.depositRequired || '0',
      depositStatus: 'held',
      txHash: `0x${Math.random().toString(16).slice(2)}`,
    };

    setState(prev => ({
      ...prev,
      submission: { ...prev.submission!, step: 'complete' },
      jobs: [...prev.jobs, job],
    }));

    return job;
  }, [validateJob, state.submission, accountAddress]);

  const cancelSubmission = useCallback(() => {
    setState(prev => ({ ...prev, submission: null }));
  }, []);

  const getJobs = useCallback(async () => {
    if (!accountAddress) {
      setState(prev => ({ ...prev, jobs: [] }));
      return;
    }

    // Would fetch jobs from chain
    setState(prev => ({ ...prev, jobs: prev.jobs }));
  }, [accountAddress]);

  const getJob = useCallback(async (jobId: string): Promise<Job> => {
    const job = state.jobs.find(j => j.id === jobId);
    if (!job) throw new Error('Job not found');
    return job;
  }, [state.jobs]);

  const cancelJob = useCallback(async (jobId: string) => {
    setState(prev => ({
      ...prev,
      jobs: prev.jobs.map(j => 
        j.id === jobId ? { ...j, status: 'cancelled' as JobStatus } : j
      ),
    }));
  }, []);

  const getOutputs = useCallback(async (jobId: string): Promise<JobOutputReference[]> => {
    const job = state.jobs.find(j => j.id === jobId);
    return job?.outputRefs || [];
  }, [state.jobs]);

  const decryptOutput = useCallback(async (outputRef: JobOutputReference): Promise<JobOutput> => {
    return {
      refId: outputRef.id,
      name: outputRef.name,
      type: outputRef.type,
      accessUrl: `https://outputs.virtengine.com/${outputRef.id}`,
      urlExpiresAt: Date.now() + 3600000,
      sizeBytes: outputRef.sizeBytes,
      mimeType: 'application/octet-stream',
    };
  }, []);

  const subscribeToJob = useCallback((
    jobId: string,
    callback: (event: ChainEvent) => void
  ): (() => void) => {
    return () => {};
  }, []);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  useEffect(() => {
    getWorkloadTemplates();
  }, [getWorkloadTemplates]);

  useEffect(() => {
    if (accountAddress) {
      getJobs();
    }
  }, [accountAddress, getJobs]);

  const actions: HPCActions = {
    refresh,
    getWorkloadTemplates,
    startJobSubmission,
    updateJobManifest,
    selectOffering,
    getQuote,
    validateJob,
    submitJob,
    cancelSubmission,
    getJobs,
    getJob,
    cancelJob,
    getOutputs,
    decryptOutput,
    subscribeToJob,
    clearError,
  };

  return (
    <HPCContext.Provider value={{ state, actions }}>
      {children}
    </HPCContext.Provider>
  );
}

export function useHPC(): HPCContextValue {
  const context = useContext(HPCContext);
  if (!context) {
    throw new Error('useHPC must be used within an HPCProvider');
  }
  return context;
}
