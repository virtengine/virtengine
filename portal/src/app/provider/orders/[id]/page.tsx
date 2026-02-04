'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import {
  useDeploymentStore,
  type Deployment,
  type DeploymentStatus,
  type DeploymentUpdatePayload,
} from '@/stores';
import { useWalletStore } from '@/stores/walletStore';
import { useWalletModal } from '@/components/wallet';
import {
  DeploymentActionsMenu,
  DeploymentUpdateModal,
  TerminateConfirmDialog,
} from '@/components/provider/deployments';
import { formatCurrency, formatDate, formatRelativeTime, sleep, truncateAddress } from '@/lib/utils';

interface DeploymentDetailPageProps {
  params: {
    id: string;
  };
}

const statusStyles: Record<DeploymentStatus, string> = {
  running: 'bg-success/10 text-success',
  paused: 'bg-warning/10 text-warning',
  restarting: 'bg-primary/10 text-primary',
  updating: 'bg-primary/10 text-primary',
  terminated: 'bg-destructive/10 text-destructive',
  failed: 'bg-destructive/10 text-destructive',
};

const healthStyles: Record<Deployment['health'], string> = {
  healthy: 'bg-success/10 text-success',
  degraded: 'bg-warning/10 text-warning',
  warning: 'bg-warning/10 text-warning',
  critical: 'bg-destructive/10 text-destructive',
};

const formatUptime = (seconds: number) => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
};

export default function DeploymentDetailPage({ params }: DeploymentDetailPageProps) {
  const { id } = params;
  const {
    deployments,
    fetchDeployment,
    stopDeployment,
    startDeployment,
    restartDeployment,
    updateDeployment,
    terminateDeployment,
    tickDeployment,
    isLoading,
  } = useDeploymentStore();
  const { isConnected, walletType, address } = useWalletStore();
  const { open: openWalletModal } = useWalletModal();

  const deployment = deployments.find((item) => item.id === id);

  const [isUpdateOpen, setIsUpdateOpen] = useState(false);
  const [isTerminateOpen, setIsTerminateOpen] = useState(false);
  const [pendingAction, setPendingAction] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [lastTx, setLastTx] = useState<{
    id: string;
    type: 'MsgUpdateDeployment' | 'MsgCloseDeployment';
    memo: string;
    createdAt: Date;
  } | null>(null);

  useEffect(() => {
    void fetchDeployment(id);
  }, [fetchDeployment, id]);

  useEffect(() => {
    if (!deployment || deployment.status === 'terminated') return;
    const interval = setInterval(() => {
      tickDeployment(id, 6);
    }, 6000);
    return () => clearInterval(interval);
  }, [deployment, id, tickDeployment]);

  const resourceSummary = useMemo(() => {
    if (!deployment) return [];
    return [
      { label: 'CPU', used: deployment.usage.cpu, total: deployment.resources.cpu, unit: 'cores' },
      {
        label: 'Memory',
        used: deployment.usage.memory,
        total: deployment.resources.memory,
        unit: 'GB',
      },
      {
        label: 'Storage',
        used: deployment.usage.storage,
        total: deployment.resources.storage,
        unit: 'GB',
      },
      ...(deployment.resources.gpu
        ? [
            {
              label: 'GPU',
              used: deployment.usage.gpu ?? 0,
              total: deployment.resources.gpu,
              unit: 'cards',
            },
          ]
        : []),
    ];
  }, [deployment]);

  const runSignedAction = async (
    action: string,
    messageType: 'MsgUpdateDeployment' | 'MsgCloseDeployment',
    memo: string,
    actionFn: () => void
  ) => {
    if (!isConnected) {
      setActionError('Connect your wallet to sign this transaction.');
      openWalletModal();
      return;
    }
    setActionError(null);
    setPendingAction(action);
    await sleep(700);
    actionFn();
    setLastTx({
      id: `tx-${Math.random().toString(36).slice(2, 10)}`,
      type: messageType,
      memo,
      createdAt: new Date(),
    });
    setPendingAction(null);
  };

  if (!deployment && isLoading) {
    return (
      <div className="container py-8">
        <div className="h-8 w-56 rounded bg-muted" />
        <div className="mt-6 grid gap-4 lg:grid-cols-3">
          <div className="h-36 rounded-lg bg-muted" />
          <div className="h-36 rounded-lg bg-muted" />
          <div className="h-36 rounded-lg bg-muted" />
        </div>
      </div>
    );
  }

  if (!deployment) {
    return (
      <div className="container py-8">
        <p className="text-muted-foreground">Deployment not found.</p>
        <Link href="/provider/orders" className="mt-4 inline-flex text-sm text-primary">
          Back to orders
        </Link>
      </div>
    );
  }

  return (
    <div className="container py-8">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link href="/provider/orders" className="text-sm text-muted-foreground hover:underline">
            Provider orders
          </Link>
          <h1 className="mt-2 text-2xl font-semibold">{deployment.name}</h1>
          <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
            <span>Deployment ID: {deployment.id}</span>
            <span>Owner: {truncateAddress(deployment.owner)}</span>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <span className={`rounded-full px-3 py-1 text-xs font-medium ${statusStyles[deployment.status]}`}>
            {deployment.status}
          </span>
          <span className={`rounded-full px-3 py-1 text-xs font-medium ${healthStyles[deployment.health]}`}>
            {deployment.health}
          </span>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-5">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="text-sm font-semibold">Deployment actions</h2>
            <p className="text-xs text-muted-foreground">
              {isConnected
                ? `Signing with ${walletType ?? 'wallet'} (${truncateAddress(address ?? '')})`
                : 'Connect wallet to sign actions'}
            </p>
          </div>
          <DeploymentActionsMenu
            status={deployment.status}
            disabled={pendingAction !== null}
            onStart={() =>
              void runSignedAction('start', 'MsgUpdateDeployment', 'Start deployment', () =>
                startDeployment(id)
              )
            }
            onStop={() =>
              void runSignedAction('stop', 'MsgUpdateDeployment', 'Stop deployment', () =>
                stopDeployment(id)
              )
            }
            onRestart={() =>
              void runSignedAction('restart', 'MsgUpdateDeployment', 'Restart deployment', () =>
                restartDeployment(id)
              )
            }
            onUpdate={() => setIsUpdateOpen(true)}
            onTerminate={() => setIsTerminateOpen(true)}
          />
        </div>
        {pendingAction && (
          <div className="mt-4 rounded-lg border border-border bg-muted/40 px-4 py-3 text-sm">
            Signing transaction for {pendingAction}...
          </div>
        )}
        {actionError && (
          <div className="mt-4 rounded-lg border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {actionError}
          </div>
        )}
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-5">
          <p className="text-xs text-muted-foreground">Status</p>
          <p className="mt-2 text-lg font-semibold">{deployment.status}</p>
          <p className="mt-2 text-sm text-muted-foreground">Updated {formatRelativeTime(deployment.updatedAt)}</p>
        </div>
        <div className="rounded-xl border border-border bg-card p-5">
          <p className="text-xs text-muted-foreground">Uptime</p>
          <p className="mt-2 text-lg font-semibold">{formatUptime(deployment.uptimeSeconds)}</p>
          <p className="mt-2 text-sm text-muted-foreground">
            Started {formatDate(deployment.createdAt)}
          </p>
        </div>
        <div className="rounded-xl border border-border bg-card p-5">
          <p className="text-xs text-muted-foreground">Cost</p>
          <p className="mt-2 text-lg font-semibold">{formatCurrency(deployment.totalCost)}</p>
          <p className="mt-2 text-sm text-muted-foreground">
            {formatCurrency(deployment.costPerHour)}/hr
          </p>
        </div>
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-[2fr_1fr]">
        <div className="rounded-xl border border-border bg-card p-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold">Resource usage</h3>
              <p className="text-xs text-muted-foreground">Real-time consumption snapshot.</p>
            </div>
            <span className="text-xs text-muted-foreground">Updated moments ago</span>
          </div>
          <div className="mt-4 space-y-4">
            {resourceSummary.map((resource) => (
              <ResourceUsageBar key={resource.label} {...resource} />
            ))}
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold">Latest transaction</h3>
          {lastTx ? (
            <div className="mt-3 space-y-2 text-sm">
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>{lastTx.type}</span>
                <span>{formatRelativeTime(lastTx.createdAt)}</span>
              </div>
              <p className="text-sm">{lastTx.memo}</p>
              <p className="text-xs text-muted-foreground">Tx hash: {lastTx.id}</p>
            </div>
          ) : (
            <p className="mt-3 text-sm text-muted-foreground">No transactions submitted yet.</p>
          )}
        </div>
      </div>

      <div className="mt-6 grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold">Containers</h3>
          <div className="mt-4 space-y-3">
            {deployment.containers.map((container) => (
              <div key={container.id} className="rounded-lg border border-border p-3 text-sm">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <p className="font-medium">{container.name}</p>
                    <p className="text-xs text-muted-foreground">{container.image}</p>
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Replicas: {container.replicas} | {container.status}
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-6 grid gap-4 md:grid-cols-2">
            <div className="rounded-lg border border-border p-4">
              <h4 className="text-xs font-semibold">Environment</h4>
              <div className="mt-2 space-y-2 text-xs text-muted-foreground">
                {deployment.env.map((envVar) => (
                  <div key={envVar.id} className="flex items-center justify-between">
                    <span>{envVar.key}</span>
                    <span>{envVar.value}</span>
                  </div>
                ))}
              </div>
            </div>
            <div className="rounded-lg border border-border p-4">
              <h4 className="text-xs font-semibold">Ports</h4>
              <div className="mt-2 space-y-2 text-xs text-muted-foreground">
                {deployment.ports.map((port) => (
                  <div key={port.id} className="flex items-center justify-between">
                    <span>
                      {port.name} ({port.protocol.toUpperCase()})
                    </span>
                    <span>
                      {port.port} / {port.exposure}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-sm font-semibold">Logs preview</h3>
            <div className="mt-4 space-y-3 text-xs text-muted-foreground">
              {deployment.logs.map((log) => (
                <div key={log.id} className="rounded-md border border-border bg-muted/30 p-2">
                  <div className="flex items-center justify-between text-[11px] uppercase text-muted-foreground">
                    <span>{log.level}</span>
                    <span>{formatRelativeTime(log.createdAt)}</span>
                  </div>
                  <p className="mt-1 text-xs text-foreground">{log.message}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-6">
            <h3 className="text-sm font-semibold">Events timeline</h3>
            <div className="mt-4 space-y-3 text-xs text-muted-foreground">
              {deployment.events.map((event) => (
                <div key={event.id} className="flex gap-3">
                  <span className="mt-1 h-2 w-2 rounded-full bg-primary" />
                  <div className="flex-1">
                    <p className="text-sm text-foreground">{event.message}</p>
                    <p className="mt-1 text-[11px] text-muted-foreground">
                      {formatRelativeTime(event.createdAt)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <DeploymentUpdateModal
        isOpen={isUpdateOpen}
        onClose={() => setIsUpdateOpen(false)}
        resources={deployment.resources}
        containers={deployment.containers}
        env={deployment.env}
        ports={deployment.ports}
        onSubmit={(payload: DeploymentUpdatePayload) => {
          setIsUpdateOpen(false);
          void runSignedAction(
            'update',
            'MsgUpdateDeployment',
            'Update deployment resources',
            () => updateDeployment(id, payload)
          );
        }}
      />

      <TerminateConfirmDialog
        isOpen={isTerminateOpen}
        onClose={() => setIsTerminateOpen(false)}
        onConfirm={() => {
          setIsTerminateOpen(false);
          void runSignedAction('terminate', 'MsgCloseDeployment', 'Terminate deployment', () =>
            terminateDeployment(id)
          );
        }}
      />
    </div>
  );
}

function ResourceUsageBar({
  label,
  used,
  total,
  unit,
}: {
  label: string;
  used: number;
  total: number;
  unit: string;
}) {
  const percentage = total === 0 ? 0 : Math.min(Math.round((used / total) * 100), 100);
  const barColor =
    percentage > 85 ? 'bg-destructive' : percentage > 70 ? 'bg-warning' : 'bg-success';

  return (
    <div>
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>{label}</span>
        <span>
          {used.toFixed(1)} / {total} {unit} ({percentage}%)
        </span>
      </div>
      <div className="mt-2 h-2 rounded-full bg-muted">
        <div className={`h-full rounded-full ${barColor}`} style={{ width: `${percentage}%` }} />
      </div>
    </div>
  );
}
