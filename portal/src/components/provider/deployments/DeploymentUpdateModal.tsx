'use client';

import { useEffect, useState } from 'react';
import { generateId } from '@/lib/utils';
import type { ContainerSpec, EnvVarSpec, PortSpec, ResourceSpec } from '@/stores';
import { ResourceScaler } from './ResourceScaler';

interface DeploymentUpdateModalProps {
  isOpen: boolean;
  onClose: () => void;
  resources: ResourceSpec;
  containers: ContainerSpec[];
  env: EnvVarSpec[];
  ports: PortSpec[];
  onSubmit: (payload: {
    resources: ResourceSpec;
    containers: ContainerSpec[];
    env: EnvVarSpec[];
    ports: PortSpec[];
  }) => void;
}

export function DeploymentUpdateModal({
  isOpen,
  onClose,
  resources,
  containers,
  env,
  ports,
  onSubmit,
}: DeploymentUpdateModalProps) {
  const [resourceState, setResourceState] = useState<ResourceSpec>(resources);
  const [containerState, setContainerState] = useState<ContainerSpec[]>(containers);
  const [envState, setEnvState] = useState<EnvVarSpec[]>(env);
  const [portState, setPortState] = useState<PortSpec[]>(ports);

  useEffect(() => {
    if (isOpen) {
      setResourceState(resources);
      setContainerState(containers);
      setEnvState(env);
      setPortState(ports);
    }
  }, [isOpen, resources, containers, env, ports]);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = '';
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />
      <div
        role="dialog"
        aria-modal="true"
        className="relative max-h-[90vh] w-full max-w-3xl overflow-hidden rounded-2xl border border-border bg-card shadow-xl"
      >
        <div className="flex items-start justify-between border-b border-border p-6">
          <div>
            <h2 className="text-xl font-semibold">Update deployment</h2>
            <p className="text-sm text-muted-foreground">
              Adjust resources, containers, and runtime configuration.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-full border border-border p-2 text-sm hover:bg-accent"
          >
            Close
          </button>
        </div>

        <div className="max-h-[68vh] overflow-y-auto p-6">
          <div className="grid gap-4 lg:grid-cols-2">
            <ResourceScaler
              label="CPU cores"
              value={resourceState.cpu}
              unit="cores"
              min={2}
              max={64}
              step={2}
              onChange={(value) => setResourceState((prev) => ({ ...prev, cpu: value }))}
            />
            <ResourceScaler
              label="Memory"
              value={resourceState.memory}
              unit="GB"
              min={4}
              max={512}
              step={4}
              onChange={(value) => setResourceState((prev) => ({ ...prev, memory: value }))}
            />
            <ResourceScaler
              label="Storage"
              value={resourceState.storage}
              unit="GB"
              min={100}
              max={4000}
              step={100}
              onChange={(value) => setResourceState((prev) => ({ ...prev, storage: value }))}
            />
            <ResourceScaler
              label="GPU"
              value={resourceState.gpu ?? 0}
              unit="cards"
              min={0}
              max={8}
              step={1}
              onChange={(value) => setResourceState((prev) => ({ ...prev, gpu: value }))}
            />
          </div>

          <div className="mt-6 rounded-xl border border-border p-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-semibold">Containers</h3>
                <p className="text-xs text-muted-foreground">Add or remove workload containers.</p>
              </div>
              <button
                type="button"
                onClick={() =>
                  setContainerState((prev) => [
                    ...prev,
                    {
                      id: generateId('ctr'),
                      name: 'new-container',
                      image: 'virtengine/app:latest',
                      replicas: 1,
                      status: 'scaled',
                    },
                  ])
                }
                className="rounded-full border border-border px-3 py-1 text-xs hover:bg-accent"
              >
                Add container
              </button>
            </div>
            <div className="mt-3 space-y-3">
              {containerState.map((container, index) => (
                <div
                  key={container.id}
                  className="grid gap-3 rounded-lg border border-border p-3 md:grid-cols-[1fr_1.2fr_120px_auto]"
                >
                  <input
                    value={container.name}
                    onChange={(event) =>
                      setContainerState((prev) =>
                        prev.map((item, idx) =>
                          idx === index ? { ...item, name: event.target.value } : item
                        )
                      )
                    }
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                  />
                  <input
                    value={container.image}
                    onChange={(event) =>
                      setContainerState((prev) =>
                        prev.map((item, idx) =>
                          idx === index ? { ...item, image: event.target.value } : item
                        )
                      )
                    }
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                  />
                  <input
                    type="number"
                    min={1}
                    value={container.replicas}
                    onChange={(event) =>
                      setContainerState((prev) =>
                        prev.map((item, idx) =>
                          idx === index
                            ? { ...item, replicas: Number(event.target.value) }
                            : item
                        )
                      )
                    }
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                  />
                  <button
                    type="button"
                    onClick={() =>
                      setContainerState((prev) => prev.filter((item) => item.id !== container.id))
                    }
                    className="rounded-full border border-border px-3 py-2 text-xs text-destructive hover:bg-destructive/10"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
          </div>

          <div className="mt-6 grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-border p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-semibold">Environment variables</h3>
                  <p className="text-xs text-muted-foreground">Runtime configuration overrides.</p>
                </div>
                <button
                  type="button"
                  onClick={() =>
                    setEnvState((prev) => [
                      ...prev,
                      {
                        id: generateId('env'),
                        key: 'NEW_VAR',
                        value: 'value',
                        scope: 'runtime',
                      },
                    ])
                  }
                  className="rounded-full border border-border px-3 py-1 text-xs hover:bg-accent"
                >
                  Add var
                </button>
              </div>
              <div className="mt-3 space-y-2">
                {envState.map((envVar, index) => (
                  <div key={envVar.id} className="flex items-center gap-2">
                    <input
                      value={envVar.key}
                      onChange={(event) =>
                        setEnvState((prev) =>
                          prev.map((item, idx) =>
                            idx === index ? { ...item, key: event.target.value } : item
                          )
                        )
                      }
                      className="flex-1 rounded-md border border-border bg-background px-2 py-1 text-xs"
                    />
                    <input
                      value={envVar.value}
                      onChange={(event) =>
                        setEnvState((prev) =>
                          prev.map((item, idx) =>
                            idx === index ? { ...item, value: event.target.value } : item
                          )
                        )
                      }
                      className="flex-1 rounded-md border border-border bg-background px-2 py-1 text-xs"
                    />
                    <select
                      value={envVar.scope}
                      onChange={(event) =>
                        setEnvState((prev) =>
                          prev.map((item, idx) =>
                            idx === index
                              ? { ...item, scope: event.target.value as EnvVarSpec['scope'] }
                              : item
                          )
                        )
                      }
                      className="rounded-md border border-border bg-background px-2 py-1 text-xs"
                    >
                      <option value="runtime">Runtime</option>
                      <option value="build">Build</option>
                    </select>
                    <button
                      type="button"
                      onClick={() => setEnvState((prev) => prev.filter((item) => item.id !== envVar.id))}
                      className="rounded-full border border-border px-2 py-1 text-xs text-destructive hover:bg-destructive/10"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-xl border border-border p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-semibold">Exposed ports</h3>
                  <p className="text-xs text-muted-foreground">Networking configuration.</p>
                </div>
                <button
                  type="button"
                  onClick={() =>
                    setPortState((prev) => [
                      ...prev,
                      {
                        id: generateId('port'),
                        name: 'service',
                        port: 7070,
                        protocol: 'tcp',
                        exposure: 'private',
                      },
                    ])
                  }
                  className="rounded-full border border-border px-3 py-1 text-xs hover:bg-accent"
                >
                  Add port
                </button>
              </div>
              <div className="mt-3 space-y-2">
                {portState.map((port, index) => (
                  <div key={port.id} className="grid gap-2 md:grid-cols-[1fr_90px_90px_90px_auto]">
                    <input
                      value={port.name}
                      onChange={(event) =>
                        setPortState((prev) =>
                          prev.map((item, idx) =>
                            idx === index ? { ...item, name: event.target.value } : item
                          )
                        )
                      }
                      className="rounded-md border border-border bg-background px-2 py-1 text-xs"
                    />
                    <input
                      type="number"
                      value={port.port}
                      onChange={(event) =>
                        setPortState((prev) =>
                          prev.map((item, idx) =>
                            idx === index ? { ...item, port: Number(event.target.value) } : item
                          )
                        )
                      }
                      className="rounded-md border border-border bg-background px-2 py-1 text-xs"
                    />
                    <select
                      value={port.protocol}
                      onChange={(event) =>
                        setPortState((prev) =>
                          prev.map((item, idx) =>
                            idx === index
                              ? { ...item, protocol: event.target.value as PortSpec['protocol'] }
                              : item
                          )
                        )
                      }
                      className="rounded-md border border-border bg-background px-2 py-1 text-xs"
                    >
                      <option value="tcp">TCP</option>
                      <option value="udp">UDP</option>
                      <option value="http">HTTP</option>
                    </select>
                    <select
                      value={port.exposure}
                      onChange={(event) =>
                        setPortState((prev) =>
                          prev.map((item, idx) =>
                            idx === index
                              ? { ...item, exposure: event.target.value as PortSpec['exposure'] }
                              : item
                          )
                        )
                      }
                      className="rounded-md border border-border bg-background px-2 py-1 text-xs"
                    >
                      <option value="public">Public</option>
                      <option value="private">Private</option>
                    </select>
                    <button
                      type="button"
                      onClick={() =>
                        setPortState((prev) => prev.filter((item) => item.id !== port.id))
                      }
                      className="rounded-full border border-border px-2 py-1 text-xs text-destructive hover:bg-destructive/10"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-between border-t border-border p-6">
          <p className="text-xs text-muted-foreground">
            Updates trigger a signed transaction before changes apply.
          </p>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={onClose}
              className="rounded-full border border-border px-4 py-2 text-sm hover:bg-accent"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() =>
                onSubmit({
                  resources: resourceState,
                  containers: containerState,
                  env: envState,
                  ports: portState,
                })
              }
              className="rounded-full bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              Submit update
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
