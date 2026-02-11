import { create } from 'zustand';
import type { Deployment as ProviderDeployment, ResourceMetrics } from '@/lib/portal-adapter';
import { MultiProviderClient } from '@/lib/portal-adapter';
import { getPortalEndpoints } from '@/lib/config';
import { coerceNumber, coerceString, toDate } from '@/lib/api/chain';

export type DeploymentStatus =
  | 'running'
  | 'paused'
  | 'restarting'
  | 'updating'
  | 'terminated'
  | 'failed';

export type DeploymentHealth = 'healthy' | 'degraded' | 'warning' | 'critical';

export interface ResourceSpec {
  cpu: number;
  memory: number;
  storage: number;
  gpu?: number;
}

export interface ResourceUsage {
  cpu: number;
  memory: number;
  storage: number;
  gpu?: number;
}

export interface ContainerSpec {
  id: string;
  name: string;
  image: string;
  replicas: number;
  status: 'running' | 'scaled' | 'stopped';
}

export interface EnvVarSpec {
  id: string;
  key: string;
  value: string;
  scope: 'runtime' | 'build';
}

export interface PortSpec {
  id: string;
  name: string;
  port: number;
  protocol: 'tcp' | 'udp' | 'http';
  exposure: 'public' | 'private';
}

export interface DeploymentEvent {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  createdAt: Date;
}

export interface DeploymentLogLine {
  id: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  createdAt: Date;
}

export interface Deployment {
  id: string;
  name: string;
  owner: string;
  status: DeploymentStatus;
  health: DeploymentHealth;
  createdAt: Date;
  updatedAt: Date;
  uptimeSeconds: number;
  costPerHour: number;
  totalCost: number;
  resources: ResourceSpec;
  usage: ResourceUsage;
  containers: ContainerSpec[];
  env: EnvVarSpec[];
  ports: PortSpec[];
  events: DeploymentEvent[];
  logs: DeploymentLogLine[];
}

export interface DeploymentState {
  deployments: Deployment[];
  isLoading: boolean;
  error: string | null;
}

export interface DeploymentActions {
  fetchDeployment: (id: string) => Promise<void>;
  refreshDeployment: (id: string) => Promise<void>;
  stopDeployment: (id: string) => Promise<void>;
  startDeployment: (id: string) => Promise<void>;
  restartDeployment: (id: string) => Promise<void>;
  updateDeployment: (id: string, update: DeploymentUpdatePayload) => Promise<void>;
  terminateDeployment: (id: string) => Promise<void>;
}

export type DeploymentStore = DeploymentState & DeploymentActions;

export interface DeploymentUpdatePayload {
  resources: ResourceSpec;
  containers: ContainerSpec[];
  env: EnvVarSpec[];
  ports: PortSpec[];
}

const initialState: DeploymentState = {
  deployments: [],
  isLoading: false,
  error: null,
};

let providerClient: MultiProviderClient | null = null;
let providerClientInit: Promise<void> | null = null;

const getProviderClient = async () => {
  if (!providerClient) {
    providerClient = new MultiProviderClient({
      chainEndpoint: getPortalEndpoints().chainRest,
    });
  }
  if (!providerClientInit) {
    providerClientInit = providerClient.initialize().catch(() => undefined);
  }
  await providerClientInit;
  return providerClient;
};

const mapDeploymentStatus = (state: string): DeploymentStatus => {
  const normalized = state.toLowerCase();
  if (normalized.includes('pause')) return 'paused';
  if (normalized.includes('restart')) return 'restarting';
  if (normalized.includes('update')) return 'updating';
  if (normalized.includes('close') || normalized.includes('terminate')) return 'terminated';
  if (normalized.includes('fail') || normalized.includes('error')) return 'failed';
  return 'running';
};

const mapHealth = (metrics?: ResourceMetrics): DeploymentHealth => {
  if (!metrics) return 'warning';
  const cpu = metrics.cpu.limit > 0 ? metrics.cpu.usage / metrics.cpu.limit : 0;
  if (cpu > 0.9) return 'degraded';
  if (cpu > 0.75) return 'warning';
  return 'healthy';
};

const mapDeployment = (deployment: ProviderDeployment, metrics?: ResourceMetrics): Deployment => {
  const createdAt = toDate(deployment.createdAt ?? new Date());
  const updatedAt = new Date();
  const resources = metrics
    ? {
        cpu: metrics.cpu.limit ?? 0,
        memory: metrics.memory.limit ?? 0,
        storage: metrics.storage.limit ?? 0,
        gpu: metrics.gpu?.limit ?? undefined,
      }
    : { cpu: 0, memory: 0, storage: 0 };

  const usage = metrics
    ? {
        cpu: metrics.cpu.usage ?? 0,
        memory: metrics.memory.usage ?? 0,
        storage: metrics.storage.usage ?? 0,
        gpu: metrics.gpu?.usage ?? undefined,
      }
    : { cpu: 0, memory: 0, storage: 0 };

  return {
    id: deployment.id,
    name: deployment.id,
    owner: coerceString(deployment.owner, ''),
    status: mapDeploymentStatus(coerceString(deployment.state, 'running')),
    health: mapHealth(metrics),
    createdAt,
    updatedAt,
    uptimeSeconds: Math.max(0, Math.floor((Date.now() - createdAt.getTime()) / 1000)),
    costPerHour: coerceNumber(metrics?.cost?.amount, 0),
    totalCost: coerceNumber(metrics?.cost?.amount, 0),
    resources,
    usage,
    containers: [],
    env: [],
    ports: [],
    events: [
      {
        id: `evt-${deployment.id}`,
        type: 'info',
        message: `Deployment status: ${coerceString(deployment.state, 'unknown')}`,
        createdAt: updatedAt,
      },
    ],
    logs: [],
  };
};

export const useDeploymentStore = create<DeploymentStore>()((set, get) => ({
  ...initialState,

  fetchDeployment: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      const client = await getProviderClient();
      const deployment = await client.getDeployment(id);
      if (!deployment) {
        throw new Error('Deployment not found');
      }
      const metrics = await client.getClient(deployment.providerId)?.getDeploymentMetrics(id);
      const logs = await client
        .getClient(deployment.providerId)
        ?.getDeploymentLogs(id, { tail: 20 });

      const mapped = mapDeployment(deployment, metrics);
      mapped.logs = (logs ?? []).map((line, idx) => ({
        id: `log-${id}-${idx}`,
        level: line.toLowerCase().includes('error')
          ? 'error'
          : line.toLowerCase().includes('warn')
            ? 'warn'
            : 'info',
        message: line,
        createdAt: new Date(),
      }));

      set((state) => {
        const exists = state.deployments.some((item) => item.id === id);
        return {
          deployments: exists
            ? state.deployments.map((item) => (item.id === id ? mapped : item))
            : [...state.deployments, mapped],
          isLoading: false,
        };
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load deployment',
      });
    }
  },

  refreshDeployment: async (id: string) => {
    await get().fetchDeployment(id);
  },

  stopDeployment: async (id: string) => {
    const client = await getProviderClient();
    await client.performAction(id, 'stop');
    await get().fetchDeployment(id);
  },

  startDeployment: async (id: string) => {
    const client = await getProviderClient();
    await client.performAction(id, 'start');
    await get().fetchDeployment(id);
  },

  restartDeployment: async (id: string) => {
    const client = await getProviderClient();
    await client.performAction(id, 'restart');
    await get().fetchDeployment(id);
  },

  updateDeployment: async (id: string, _update: DeploymentUpdatePayload) => {
    const client = await getProviderClient();
    await client.performAction(id, 'update');
    await get().fetchDeployment(id);
  },

  terminateDeployment: async (id: string) => {
    const client = await getProviderClient();
    await client.performAction(id, 'terminate');
    await get().fetchDeployment(id);
  },
}));
