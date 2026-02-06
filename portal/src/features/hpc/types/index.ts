/**
 * HPC Feature Types
 *
 * Re-exports from portal-lib with additional portal-specific types
 */

// Re-export all types from portal-lib
export type {
  HPCState,
  WorkloadTemplate,
  WorkloadCategory,
  JobResources,
  JobParameter,
  JobManifest,
  JobSubmission,
  JobSubmissionState,
  JobSubmissionStep,
  JobPriceQuote,
  JobValidationError,
  Job,
  JobStatus,
  JobStatusChange,
  JobEvent,
  JobEventType,
  JobOutputReference,
  JobOutputType,
  JobOutput,
  HPCError,
  HPCErrorCode,
} from 'virtengine-portal-lib/types/hpc';

/**
 * HPC SDK types (these will come from @virtengine/chain-sdk when integrated)
 */
export interface SDKJob {
  jobId: string;
  offeringId: string;
  clusterId: string;
  providerAddress: string;
  customerAddress: string;
  slurmJobId?: string;
  state: string;
  queueName: string;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
  totalCost: string;
}

export interface SDKOffering {
  offeringId: string;
  clusterId: string;
  providerAddress: string;
  name: string;
  description: string;
  pricing: {
    baseNodeHourPrice: string;
    cpuCoreHourPrice: string;
    memoryGbHourPrice: string;
    storageGbPrice: string;
    networkGbPrice: string;
    currency: string;
  };
  maxRuntimeSeconds: number;
  supportsCustomWorkloads: boolean;
  preconfiguredWorkloads: SDKWorkloadTemplate[];
}

export interface SDKWorkloadTemplate {
  templateId: string;
  name: string;
  description: string;
  workloadType: string;
  containerImage: string;
  defaultCommand: string;
  approvalStatus: string;
  requiredIdentityThreshold: number;
}

export interface SDKCluster {
  clusterId: string;
  providerAddress: string;
  name: string;
  description: string;
  region: string;
  state: string;
  totalNodes: number;
  totalGpus: number;
}

/**
 * Job filter options
 */
export interface JobFilters {
  status?: JobStatus[];
  templateId?: string;
  dateFrom?: number;
  dateTo?: number;
  searchQuery?: string;
}

/**
 * Job sort options
 */
export type JobSortField = 'createdAt' | 'name' | 'status' | 'totalCost';
export type JobSortOrder = 'asc' | 'desc';

export interface JobSort {
  field: JobSortField;
  order: JobSortOrder;
}

/**
 * Template filter options
 */
export interface TemplateFilters {
  category?: WorkloadCategory[];
  gpuRequired?: boolean;
  searchQuery?: string;
}

/**
 * Offering filter options
 */
export interface OfferingFilters {
  region?: string;
  gpuType?: string;
  maxPricePerHour?: string;
}
