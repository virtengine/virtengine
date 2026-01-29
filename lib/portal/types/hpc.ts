/**
 * HPC Types
 * VE-705: Supercomputer/HPC UI (job submission, library workloads, outputs)
 *
 * @packageDocumentation
 */

/**
 * HPC state
 */
export interface HPCState {
  /**
   * Whether data is loading
   */
  isLoading: boolean;

  /**
   * Available workload templates
   */
  workloadTemplates: WorkloadTemplate[];

  /**
   * User's jobs
   */
  jobs: Job[];

  /**
   * Selected job (for detail view)
   */
  selectedJob: Job | null;

  /**
   * Current job submission state
   */
  submission: JobSubmissionState | null;

  /**
   * HPC error
   */
  error: HPCError | null;
}

/**
 * Workload template (preconfigured job type)
 */
export interface WorkloadTemplate {
  /**
   * Template ID
   */
  id: string;

  /**
   * Template name
   */
  name: string;

  /**
   * Description
   */
  description: string;

  /**
   * Template category
   */
  category: WorkloadCategory;

  /**
   * Default resources
   */
  defaultResources: JobResources;

  /**
   * Default parameters
   */
  defaultParameters: Record<string, JobParameter>;

  /**
   * Required identity score
   */
  requiredIdentityScore: number;

  /**
   * MFA required
   */
  mfaRequired: boolean;

  /**
   * Estimated cost per hour
   */
  estimatedCostPerHour: string;

  /**
   * Template version
   */
  version: string;

  /**
   * Template icon URL
   */
  iconUrl?: string;

  /**
   * Documentation URL
   */
  docsUrl?: string;
}

/**
 * Workload categories
 */
export type WorkloadCategory =
  | 'ml_training'
  | 'ml_inference'
  | 'scientific'
  | 'rendering'
  | 'simulation'
  | 'data_processing'
  | 'custom';

/**
 * Job resources
 */
export interface JobResources {
  /**
   * Number of nodes
   */
  nodes: number;

  /**
   * CPUs per node
   */
  cpusPerNode: number;

  /**
   * Memory GB per node
   */
  memoryGBPerNode: number;

  /**
   * GPUs per node
   */
  gpusPerNode?: number;

  /**
   * GPU type
   */
  gpuType?: string;

  /**
   * Maximum runtime in seconds
   */
  maxRuntimeSeconds: number;

  /**
   * Storage GB
   */
  storageGB: number;
}

/**
 * Job parameter definition
 */
export interface JobParameter {
  /**
   * Parameter name
   */
  name: string;

  /**
   * Parameter type
   */
  type: 'string' | 'number' | 'boolean' | 'select' | 'file';

  /**
   * Parameter description
   */
  description: string;

  /**
   * Whether parameter is required
   */
  required: boolean;

  /**
   * Default value
   */
  defaultValue?: string | number | boolean;

  /**
   * Options (for select type)
   */
  options?: { value: string; label: string }[];

  /**
   * Validation regex (for string type)
   */
  validationPattern?: string;

  /**
   * Min value (for number type)
   */
  min?: number;

  /**
   * Max value (for number type)
   */
  max?: number;
}

/**
 * Job manifest for submission
 */
export interface JobManifest {
  /**
   * Manifest version
   */
  version: string;

  /**
   * Job name
   */
  name: string;

  /**
   * Job description
   */
  description?: string;

  /**
   * Template ID (if using template)
   */
  templateId?: string;

  /**
   * Resources requested
   */
  resources: JobResources;

  /**
   * Job parameters
   */
  parameters: Record<string, string | number | boolean>;

  /**
   * Input data references (encrypted)
   */
  inputRefs?: string[];

  /**
   * Encrypted inputs payload (client-side encrypted JSON)
   */
  encryptedInputs?: Record<string, unknown>;

  /**
   * Environment variables (will be encrypted)
   */
  environment?: Record<string, string>;

  /**
   * Command to run (for custom jobs)
   */
  command?: string;

  /**
   * Container image (for custom jobs)
   */
  image?: string;
}

/**
 * Job submission request
 */
export interface JobSubmission {
  /**
   * Job manifest
   */
  manifest: JobManifest;

  /**
   * Offering ID to use
   */
  offeringId: string;

  /**
   * Estimated cost
   */
  estimatedCost: string;

  /**
   * Deposit amount
   */
  depositAmount: string;
}

/**
 * Job submission state
 */
export interface JobSubmissionState {
  /**
   * Submission step
   */
  step: JobSubmissionStep;

  /**
   * Job manifest being built
   */
  manifest: Partial<JobManifest>;

  /**
   * Selected template (if any)
   */
  selectedTemplate: WorkloadTemplate | null;

  /**
   * Selected offering
   */
  selectedOffering: string | null;

  /**
   * Price quote
   */
  priceQuote: JobPriceQuote | null;

  /**
   * Validation errors
   */
  validationErrors: JobValidationError[];

  /**
   * Submission error
   */
  error: HPCError | null;
}

/**
 * Job submission steps
 */
export type JobSubmissionStep =
  | 'select_template'
  | 'configure'
  | 'select_offering'
  | 'review'
  | 'mfa'
  | 'submit'
  | 'complete';

/**
 * Job price quote
 */
export interface JobPriceQuote {
  /**
   * Estimated total cost
   */
  estimatedTotal: string;

  /**
   * Deposit required
   */
  depositRequired: string;

  /**
   * Cost breakdown
   */
  breakdown: {
    compute: string;
    storage: string;
    network: string;
    gpu?: string;
  };

  /**
   * Price per hour
   */
  pricePerHour: string;

  /**
   * Maximum runtime hours
   */
  maxHours: number;

  /**
   * Currency denom
   */
  denom: string;
}

/**
 * Job validation error
 */
export interface JobValidationError {
  /**
   * Field with error
   */
  field: string;

  /**
   * Error message
   */
  message: string;
}

/**
 * Job representation
 */
export interface Job {
  /**
   * Job ID (on-chain)
   */
  id: string;

  /**
   * Job name
   */
  name: string;

  /**
   * Customer address
   */
  customerAddress: string;

  /**
   * Provider address
   */
  providerAddress: string;

  /**
   * Offering ID used
   */
  offeringId: string;

  /**
   * Template ID (if used)
   */
  templateId?: string;

  /**
   * Current job status
   */
  status: JobStatus;

  /**
   * Job creation timestamp
   */
  createdAt: number;

  /**
   * Job start timestamp
   */
  startedAt?: number;

  /**
   * Job completion timestamp
   */
  completedAt?: number;

  /**
   * Requested resources
   */
  resources: JobResources;

  /**
   * Status history
   */
  statusHistory: JobStatusChange[];

  /**
   * Job events
   */
  events: JobEvent[];

  /**
   * Output references (encrypted)
   */
  outputRefs: JobOutputReference[];

  /**
   * Total cost
   */
  totalCost: string;

  /**
   * Deposit amount
   */
  depositAmount: string;

  /**
   * Deposit status
   */
  depositStatus: 'held' | 'released' | 'forfeited';

  /**
   * Transaction hash of job creation
   */
  txHash: string;
}

/**
 * Job status
 */
export type JobStatus =
  | 'pending'      // Job submitted, awaiting scheduling
  | 'queued'       // Job queued in SLURM
  | 'running'      // Job running
  | 'completing'   // Job completing, outputs being collected
  | 'completed'    // Job completed successfully
  | 'failed'       // Job failed
  | 'cancelled'    // Job cancelled by user
  | 'timeout';     // Job exceeded max runtime

/**
 * Job status change record
 */
export interface JobStatusChange {
  /**
   * Previous status
   */
  fromStatus: JobStatus;

  /**
   * New status
   */
  toStatus: JobStatus;

  /**
   * Timestamp of change
   */
  timestamp: number;

  /**
   * Block height
   */
  blockHeight: number;

  /**
   * Transaction hash
   */
  txHash: string;

  /**
   * Reason for change (if failed/cancelled)
   */
  reason?: string;
}

/**
 * Job event
 */
export interface JobEvent {
  /**
   * Event ID
   */
  id: string;

  /**
   * Event type
   */
  type: JobEventType;

  /**
   * Event timestamp
   */
  timestamp: number;

  /**
   * Block height
   */
  blockHeight: number;

  /**
   * Event data (non-sensitive)
   */
  data: Record<string, unknown>;
}

/**
 * Job event types
 */
export type JobEventType =
  | 'job_submitted'
  | 'job_scheduled'
  | 'job_started'
  | 'checkpoint_saved'
  | 'progress_updated'
  | 'output_available'
  | 'job_completed'
  | 'job_failed'
  | 'job_cancelled'
  | 'usage_recorded'
  | 'settlement_processed';

/**
 * Job output reference (encrypted pointer)
 */
export interface JobOutputReference {
  /**
   * Reference ID
   */
  id: string;

  /**
   * Output name
   */
  name: string;

  /**
   * Output type
   */
  type: JobOutputType;

  /**
   * Size in bytes
   */
  sizeBytes: number;

  /**
   * Created at
   */
  createdAt: number;

  /**
   * Encrypted reference (only owner can decrypt)
   */
  encryptedRef: string;

  /**
   * Hash of unencrypted content (for verification)
   */
  contentHash: string;

  /**
   * Expiration timestamp
   */
  expiresAt?: number;
}

/**
 * Job output types
 */
export type JobOutputType =
  | 'model'
  | 'checkpoint'
  | 'logs'
  | 'metrics'
  | 'artifact'
  | 'data';

/**
 * Job output (decrypted, local only)
 */
export interface JobOutput {
  /**
   * Reference ID
   */
  refId: string;

  /**
   * Output name
   */
  name: string;

  /**
   * Output type
   */
  type: JobOutputType;

  /**
   * Access URL (temporary, authenticated)
   */
  accessUrl: string;

  /**
   * URL expiration
   */
  urlExpiresAt: number;

  /**
   * Size in bytes
   */
  sizeBytes: number;

  /**
   * MIME type
   */
  mimeType: string;
}

/**
 * HPC error
 */
export interface HPCError {
  /**
   * Error code
   */
  code: HPCErrorCode;

  /**
   * Human-readable message
   */
  message: string;

  /**
   * Additional details
   */
  details?: Record<string, unknown>;
}

/**
 * HPC error codes
 */
export type HPCErrorCode =
  | 'template_not_found'
  | 'invalid_manifest'
  | 'insufficient_identity'
  | 'mfa_required'
  | 'insufficient_funds'
  | 'resource_unavailable'
  | 'job_not_found'
  | 'output_access_denied'
  | 'output_expired'
  | 'cancellation_failed'
  | 'network_error'
  | 'unknown';

/**
 * Initial HPC state
 */
export const initialHPCState: HPCState = {
  isLoading: false,
  workloadTemplates: [],
  jobs: [],
  selectedJob: null,
  submission: null,
  error: null,
};
