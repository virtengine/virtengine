'use client';

import { useState } from 'react';
import type { DeploymentStatus } from '@/stores';

interface DeploymentActionsMenuProps {
  status: DeploymentStatus;
  disabled?: boolean;
  onStop: () => void;
  onStart: () => void;
  onRestart: () => void;
  onUpdate: () => void;
  onTerminate: () => void;
}

export function DeploymentActionsMenu({
  status,
  disabled,
  onStop,
  onStart,
  onRestart,
  onUpdate,
  onTerminate,
}: DeploymentActionsMenuProps) {
  const [open, setOpen] = useState(false);

  const isTerminated = status === 'terminated';
  const isRunning = status === 'running';
  const isPaused = status === 'paused';
  const isBusy = status === 'updating' || status === 'restarting';

  const actionDisabled = disabled || isBusy || isTerminated;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <button
        type="button"
        onClick={onStart}
        disabled={actionDisabled || !isPaused}
        className="rounded-full border border-border px-4 py-2 text-sm font-medium hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
      >
        Start
      </button>
      <button
        type="button"
        onClick={onStop}
        disabled={actionDisabled || !isRunning}
        className="rounded-full border border-border px-4 py-2 text-sm font-medium hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
      >
        Stop
      </button>
      <button
        type="button"
        onClick={onRestart}
        disabled={actionDisabled || !isRunning}
        className="rounded-full border border-border px-4 py-2 text-sm font-medium hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
      >
        Restart
      </button>

      <div className="relative">
        <button
          type="button"
          onClick={() => setOpen((prev) => !prev)}
          disabled={actionDisabled}
          className="rounded-full border border-border px-4 py-2 text-sm font-medium hover:bg-accent disabled:cursor-not-allowed disabled:opacity-50"
        >
          More
        </button>
        {open && (
          <div className="absolute right-0 z-20 mt-2 w-48 rounded-xl border border-border bg-card p-2 shadow-lg">
            <button
              type="button"
              onClick={() => {
                setOpen(false);
                onUpdate();
              }}
              className="w-full rounded-lg px-3 py-2 text-left text-sm hover:bg-accent"
            >
              Update resources
            </button>
            <button
              type="button"
              onClick={() => {
                setOpen(false);
                onTerminate();
              }}
              className="w-full rounded-lg px-3 py-2 text-left text-sm text-destructive hover:bg-destructive/10"
            >
              Terminate deployment
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
