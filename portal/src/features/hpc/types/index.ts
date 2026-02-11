/**
 * HPC Feature Types
 *
 * Re-exports from portal-lib with additional portal-specific types
 */

// Import only types needed for local interface definitions
import type { HPCCluster, HPCJob, HPCOffering, PreconfiguredWorkload } from '@virtengine/chain-sdk';
import type { JobStatus, WorkloadCategory } from '@/lib/portal-adapter';

// Re-export all types from portal-lib (via portal-adapter)
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
} from '@/lib/portal-adapter';

/**
 * HPC SDK types (from @virtengine/chain-sdk)
 */
export type SDKJob = HPCJob;
export type SDKOffering = HPCOffering;
export type SDKWorkloadTemplate = PreconfiguredWorkload;
export type SDKCluster = HPCCluster;

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
