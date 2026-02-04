import { create } from 'zustand';
import { generateId } from '@/lib/utils';

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
  stopDeployment: (id: string) => void;
  startDeployment: (id: string) => void;
  restartDeployment: (id: string) => void;
  updateDeployment: (id: string, update: DeploymentUpdatePayload) => void;
  terminateDeployment: (id: string) => void;
  tickDeployment: (id: string, seconds?: number) => void;
}

export type DeploymentStore = DeploymentState & DeploymentActions;

export interface DeploymentUpdatePayload {
  resources: ResourceSpec;
  containers: ContainerSpec[];
  env: EnvVarSpec[];
  ports: PortSpec[];
}

const estimateHourlyCost = (resources: ResourceSpec) => {
  const gpuCost = resources.gpu ? resources.gpu * 1.4 : 0;
  return resources.cpu * 0.08 + resources.memory * 0.015 + resources.storage * 0.002 + gpuCost;
};

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max);

const seedDeployment = (id: string): Deployment => {
  const resources = {
    cpu: 24,
    memory: 96,
    storage: 1200,
    gpu: 2,
  };
  const costPerHour = estimateHourlyCost(resources);

  return {
    id,
    name: 'HPC Model Serving Cluster',
    owner: 'virtengine1abc...7h3k',
    status: 'running',
    health: 'healthy',
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 26),
    updatedAt: new Date(),
    uptimeSeconds: 60 * 60 * 24,
    costPerHour,
    totalCost: costPerHour * 24,
    resources,
    usage: {
      cpu: 14,
      memory: 58,
      storage: 680,
      gpu: 1.2,
    },
    containers: [
      {
        id: generateId('ctr'),
        name: 'api-gateway',
        image: 'virtengine/gateway:v2.4.1',
        replicas: 2,
        status: 'running',
      },
      {
        id: generateId('ctr'),
        name: 'inference-worker',
        image: 'virtengine/inference:v1.8.0',
        replicas: 6,
        status: 'running',
      },
    ],
    env: [
      { id: generateId('env'), key: 'MODEL_VERSION', value: 'v5.1.2', scope: 'runtime' },
      { id: generateId('env'), key: 'CACHE_SIZE', value: '8Gi', scope: 'runtime' },
      { id: generateId('env'), key: 'LOG_LEVEL', value: 'info', scope: 'runtime' },
    ],
    ports: [
      { id: generateId('port'), name: 'http', port: 8080, protocol: 'http', exposure: 'public' },
      { id: generateId('port'), name: 'metrics', port: 9090, protocol: 'tcp', exposure: 'private' },
    ],
    events: [
      {
        id: generateId('evt'),
        type: 'success',
        message: 'Deployment started and passed health checks.',
        createdAt: new Date(Date.now() - 1000 * 60 * 45),
      },
      {
        id: generateId('evt'),
        type: 'info',
        message: 'Autoscaler adjusted worker replicas to 6.',
        createdAt: new Date(Date.now() - 1000 * 60 * 15),
      },
    ],
    logs: [
      {
        id: generateId('log'),
        level: 'info',
        message: 'Inference worker warmed with latest model weights.',
        createdAt: new Date(Date.now() - 1000 * 60 * 4),
      },
      {
        id: generateId('log'),
        level: 'warn',
        message: 'Gateway latency exceeded 200ms threshold for 2 minutes.',
        createdAt: new Date(Date.now() - 1000 * 60 * 2),
      },
      {
        id: generateId('log'),
        level: 'info',
        message: 'Cache hit ratio stabilized at 92%.',
        createdAt: new Date(Date.now() - 1000 * 60),
      },
    ],
  };
};

const updateUsageValue = (current: number, max: number, status: DeploymentStatus) => {
  const variance = status === 'running' ? 0.12 : status === 'updating' ? 0.05 : 0.02;
  const target = status === 'running' ? max * 0.55 : status === 'updating' ? max * 0.35 : max * 0.1;
  const next = current + (target - current) * variance + (Math.random() - 0.5) * max * 0.02;
  return clamp(next, max * 0.05, max * 0.92);
};

const maxLogs = 8;
const maxEvents = 10;

const createEvent = (type: DeploymentEvent['type'], message: string): DeploymentEvent => ({
  id: generateId('evt'),
  type,
  message,
  createdAt: new Date(),
});

const createLog = (level: DeploymentLogLine['level'], message: string): DeploymentLogLine => ({
  id: generateId('log'),
  level,
  message,
  createdAt: new Date(),
});

export const useDeploymentStore = create<DeploymentStore>()((set, get) => ({
  deployments: [seedDeployment('ord-001')],
  isLoading: false,
  error: null,

  fetchDeployment: async (id: string) => {
    set({ isLoading: true, error: null });

    try {
      await new Promise((resolve) => setTimeout(resolve, 400));
      const { deployments } = get();
      if (deployments.some((deployment) => deployment.id === id)) {
        set({ isLoading: false });
        return;
      }

      const newDeployment = seedDeployment(id);
      set((state) => ({
        deployments: [...state.deployments, newDeployment],
        isLoading: false,
      }));
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load deployment',
      });
    }
  },

  stopDeployment: (id: string) => {
    set((state) => ({
      deployments: state.deployments.map((deployment) =>
        deployment.id === id
          ? {
              ...deployment,
              status: 'paused',
              updatedAt: new Date(),
              events: [createEvent('warning', 'Deployment paused by provider.'), ...deployment.events].slice(
                0,
                maxEvents
              ),
            }
          : deployment
      ),
    }));
  },

  startDeployment: (id: string) => {
    set((state) => ({
      deployments: state.deployments.map((deployment) =>
        deployment.id === id
          ? {
              ...deployment,
              status: 'running',
              updatedAt: new Date(),
              events: [
                createEvent('success', 'Deployment resumed and workloads are active.'),
                ...deployment.events,
              ].slice(0, maxEvents),
            }
          : deployment
      ),
    }));
  },

  restartDeployment: (id: string) => {
    set((state) => ({
      deployments: state.deployments.map((deployment) =>
        deployment.id === id
          ? {
              ...deployment,
              status: 'restarting',
              updatedAt: new Date(),
              events: [
                createEvent('info', 'Restart initiated. Draining workloads.'),
                ...deployment.events,
              ].slice(0, maxEvents),
            }
          : deployment
      ),
    }));

    setTimeout(() => {
      set((state) => ({
        deployments: state.deployments.map((deployment) =>
          deployment.id === id
            ? {
                ...deployment,
                status: 'running',
                updatedAt: new Date(),
                events: [
                  createEvent('success', 'Restart completed successfully.'),
                  ...deployment.events,
                ].slice(0, maxEvents),
              }
            : deployment
        ),
      }));
    }, 1200);
  },

  updateDeployment: (id: string, update: DeploymentUpdatePayload) => {
    set((state) => ({
      deployments: state.deployments.map((deployment) =>
        deployment.id === id
          ? {
              ...deployment,
              status: 'updating',
              updatedAt: new Date(),
              events: [
                createEvent('info', 'Update submitted. Applying configuration changes.'),
                ...deployment.events,
              ].slice(0, maxEvents),
            }
          : deployment
      ),
    }));

    setTimeout(() => {
      set((state) => ({
        deployments: state.deployments.map((deployment) => {
          if (deployment.id !== id) return deployment;

          const nextResources = update.resources;
          const nextUsage: ResourceUsage = {
            cpu: clamp(deployment.usage.cpu, 1, nextResources.cpu),
            memory: clamp(deployment.usage.memory, 1, nextResources.memory),
            storage: clamp(deployment.usage.storage, 1, nextResources.storage),
            gpu: nextResources.gpu
              ? clamp(deployment.usage.gpu ?? 0, 0, nextResources.gpu)
              : undefined,
          };

          return {
            ...deployment,
            status: 'running',
            updatedAt: new Date(),
            resources: nextResources,
            usage: nextUsage,
            containers: update.containers,
            env: update.env,
            ports: update.ports,
            costPerHour: estimateHourlyCost(update.resources),
            events: [
              createEvent('success', 'Update applied and deployment is healthy.'),
              ...deployment.events,
            ].slice(0, maxEvents),
          };
        }),
      }));
    }, 1400);
  },

  terminateDeployment: (id: string) => {
    set((state) => ({
      deployments: state.deployments.map((deployment) =>
        deployment.id === id
          ? {
              ...deployment,
              status: 'terminated',
              updatedAt: new Date(),
              events: [
                createEvent('error', 'Deployment terminated and resources released.'),
                ...deployment.events,
              ].slice(0, maxEvents),
            }
          : deployment
      ),
    }));
  },

  tickDeployment: (id: string, seconds = 5) => {
    set((state) => ({
      deployments: state.deployments.map((deployment) => {
        if (deployment.id !== id) return deployment;

        const nextUsage = {
          cpu: updateUsageValue(deployment.usage.cpu, deployment.resources.cpu, deployment.status),
          memory: updateUsageValue(
            deployment.usage.memory,
            deployment.resources.memory,
            deployment.status
          ),
          storage: updateUsageValue(
            deployment.usage.storage,
            deployment.resources.storage,
            deployment.status
          ),
          gpu:
            deployment.resources.gpu && deployment.usage.gpu
              ? updateUsageValue(deployment.usage.gpu, deployment.resources.gpu, deployment.status)
              : undefined,
        };

        const nextHealth: DeploymentHealth =
          deployment.status === 'running' && nextUsage.cpu / deployment.resources.cpu > 0.85
            ? 'degraded'
            : deployment.status === 'paused'
            ? 'warning'
            : deployment.status === 'terminated'
            ? 'critical'
            : 'healthy';

        const nextLogs =
          deployment.status === 'running' && Math.random() > 0.6
            ? [createLog('info', 'Auto-tuner adjusted batch size for throughput.'), ...deployment.logs].slice(
                0,
                maxLogs
              )
            : deployment.logs;

        return {
          ...deployment,
          usage: nextUsage,
          health: nextHealth,
          uptimeSeconds:
            deployment.status === 'running' || deployment.status === 'updating'
              ? deployment.uptimeSeconds + seconds
              : deployment.uptimeSeconds,
          totalCost:
            deployment.status === 'running' || deployment.status === 'updating'
              ? deployment.totalCost + deployment.costPerHour * (seconds / 3600)
              : deployment.totalCost,
          logs: nextLogs,
        };
      }),
    }));
  },
}));
