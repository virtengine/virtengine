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

  const [name, setName] = useState('');
  const [metric, setMetric] = useState<AlertMetric>('cpu');
  const [condition, setCondition] = useState<AlertCondition>('gt');
  const [threshold, setThreshold] = useState('80');
  const [duration, setDuration] = useState('300');
  const [deploymentId, setDeploymentId] = useState<string>('');

  function handleSubmit() {
    if (!name.trim()) return;

    createAlert({
      name: name.trim(),
      metric,
      condition,
      threshold: Number(threshold),
      duration: Number(duration),
      deploymentId: deploymentId || undefined,
      notificationChannels: ['email'],
    });

    setName('');
    setMetric('cpu');
    setCondition('gt');
    setThreshold('80');
    setDuration('300');
    setDeploymentId('');
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create Alert</DialogTitle>
          <DialogDescription>Set a threshold-based alert for resource metrics.</DialogDescription>
        </DialogHeader>

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
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!name.trim()}>
            Create Alert
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
