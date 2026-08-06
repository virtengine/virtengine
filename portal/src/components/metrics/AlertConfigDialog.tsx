/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Alert configuration dialog for creating and editing metric alerts.
 */

'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { useMetricsStore } from '@/stores/metricsStore';
import type { AlertMetric, AlertCondition } from '@virtengine/portal/types/metrics';

interface AlertConfigDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AlertConfigDialog({ open, onOpenChange }: AlertConfigDialogProps) {
  const createAlert = useMetricsStore((s) => s.createAlert);
  const deploymentMetrics = useMetricsStore((s) => s.deploymentMetrics);
  const alertMutationPending = useMetricsStore((s) => s.alertMutationPending);
  const alertMutationsAvailable = useMetricsStore((s) => s.alertMutationsAvailable);
  const error = useMetricsStore((s) => s.error);

  const [name, setName] = useState('');
  const [metric, setMetric] = useState<AlertMetric>('cpu');
  const [condition, setCondition] = useState<AlertCondition>('gt');
  const [threshold, setThreshold] = useState('80');
  const [duration, setDuration] = useState('300');
  const [deploymentId, setDeploymentId] = useState<string>('');

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && alertMutationPending) return;
    onOpenChange(nextOpen);
  };

  async function handleSubmit() {
    if (!name.trim()) return;

    try {
      await createAlert({
        name: name.trim(),
        metric,
        condition,
        threshold: Number(threshold),
        duration: Number(duration),
        deploymentId: deploymentId || undefined,
        notificationChannels: ['email'],
      });
    } catch {
      return;
    }

    setName('');
    setMetric('cpu');
    setCondition('gt');
    setThreshold('80');
    setDuration('300');
    setDeploymentId('');
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create Alert</DialogTitle>
          <DialogDescription>Set a threshold-based alert for resource metrics.</DialogDescription>
        </DialogHeader>

        {!alertMutationsAvailable && (
          <Alert variant="warning">
            <AlertDescription>Alert persistence is unavailable.</AlertDescription>
          </Alert>
        )}
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {alertMutationPending && (
          <p role="status" aria-live="polite" className="text-sm text-muted-foreground">
            Creating alert...
          </p>
        )}

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="alert-name">Alert Name</Label>
            <Input
              id="alert-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="High CPU usage"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Metric</Label>
              <Select value={metric} onValueChange={(v) => setMetric(v as AlertMetric)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="cpu">CPU</SelectItem>
                  <SelectItem value="memory">Memory</SelectItem>
                  <SelectItem value="storage">Storage</SelectItem>
                  <SelectItem value="network">Network</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Condition</Label>
              <Select value={condition} onValueChange={(v) => setCondition(v as AlertCondition)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="gt">Greater than</SelectItem>
                  <SelectItem value="lt">Less than</SelectItem>
                  <SelectItem value="eq">Equal to</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="alert-threshold">Threshold (%)</Label>
              <Input
                id="alert-threshold"
                type="number"
                value={threshold}
                onChange={(e) => setThreshold(e.target.value)}
                min="0"
                max="100"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="alert-duration">Duration (seconds)</Label>
              <Input
                id="alert-duration"
                type="number"
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                min="60"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label>Deployment (optional)</Label>
            <Select value={deploymentId} onValueChange={setDeploymentId}>
              <SelectTrigger>
                <SelectValue placeholder="All deployments" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All deployments</SelectItem>
                {deploymentMetrics.map((d) => (
                  <SelectItem key={d.deploymentId} value={d.deploymentId}>
                    {d.deploymentId} ({d.provider})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            disabled={alertMutationPending}
            onClick={() => handleOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            onClick={() => void handleSubmit()}
            disabled={!name.trim() || alertMutationPending || !alertMutationsAvailable}
          >
            Create Alert
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
