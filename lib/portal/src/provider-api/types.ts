export type ProviderHealthStatus = "ok" | "degraded" | "down";

export interface ProviderHealth {
  status: ProviderHealthStatus;
  version?: string;
  uptime?: number;
}

export interface LogOptions {
  follow?: boolean;
  tail?: number;
  since?: Date;
  timestamps?: boolean;
  level?: string;
  search?: string;
}

export type DeploymentState = "pending" | "active" | "closed" | string;

export interface UsageMetric {
  usage: number;
  limit: number;
}

export interface ResourceMetrics {
  cpu: UsageMetric;
  memory: UsageMetric;
  storage: UsageMetric;
  network?: {
    rxBytes?: number;
    txBytes?: number;
  };
  gpu?: {
    usage?: number;
    limit?: number;
  };
  cost?: {
    amount: string;
    currency: string;
  };
  timestamp?: Date;
}

export interface Deployment {
  id: string;
  owner?: string;
  provider?: string;
  state: DeploymentState;
  createdAt?: Date;
  resources?: ResourceMetrics;
}

export interface DeploymentStatus {
  leaseId: string;
  state: string;
  replicas: { ready: number; total: number };
  services: ServiceStatus[];
  lastUpdated?: Date;
}

export interface ServiceStatus {
  name: string;
  state: string;
  replicas: number;
  ports?: Array<{ port: number; protocol: string }>;
}

export interface DeploymentListResponse {
  deployments: Deployment[];
  nextCursor?: string | null;
}

export type DeploymentAction = "start" | "stop" | "restart" | string;

export interface ShellSessionResponse {
  token: string;
  expiresAt: string;
  deployment: string;
  container?: string;
  sessionTtl?: number;
}

export interface ProviderAPIErrorDetails {
  code?: string;
  message?: string;
  [key: string]: unknown;
}
