export type {
  ProviderHealthStatus,
  ProviderHealth,
  LogOptions,
  DeploymentState,
  UsageMetric,
  ResourceMetrics,
  Deployment,
  DeploymentStatus,
  ServiceStatus,
  DeploymentListResponse,
  DeploymentAction,
  ProviderAPIErrorDetails,
} from "./types";

export {
  ProviderAPIClient,
  ProviderAPIError,
  LogStream,
  ShellConnection,
} from "./client";

export type { ProviderAPIClientOptions } from "./client";

export {
  ProviderShellSessionError,
  buildProviderShellWebSocketUrl,
  validateProviderShellSessionReceipt,
} from "./shell-session";
export type {
  ProviderShellSessionErrorCode,
  ProviderShellSessionCapability,
  ProviderShellSessionReceipt,
  ShellEligibilityProjection,
  ShellSessionValidationContext,
} from "./shell-session";

export {
  providerDeploymentActions,
  ProviderDeploymentActionError,
  validateProviderDeploymentActionReceipt,
} from "./deployment-actions";
export type {
  ProviderDeploymentAction,
  ProviderDeploymentActionStatus,
  ProviderDeploymentActionTxEvidence,
  ProviderDeploymentActionReceipt,
  ProviderDeploymentActionCapability,
  ProviderDeploymentActionErrorCode,
  ProviderDeploymentTxEvidenceValidator,
  ProviderDeploymentActionValidationContext,
  ProviderDeploymentActionReceiptValidator,
} from "./deployment-actions";
