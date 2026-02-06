/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Label } from '@/components/ui/Label';
import { Input } from '@/components/ui/Input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import type { Offering } from '@/types/offerings';
import { type ResourceConfig, RESOURCE_LIMITS } from '@/features/orders';

interface ResourceConfigStepProps {
  offering: Offering;
  resources: ResourceConfig;
  validationErrors: string[];
  onChange: (update: Partial<ResourceConfig>) => void;
}

/**
 * Step 1: Resource Configuration
 * Allows users to configure CPU, RAM, GPU, storage, duration and region.
 */
export function ResourceConfigStep({
  offering,
  resources,
  validationErrors,
  onChange,
}: ResourceConfigStepProps) {
  const hasGpu = offering.category === 'gpu' || offering.category === 'ml';
  const hasGpuPricing = offering.prices?.some((p) => p.resourceType === 'gpu');

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Compute Resources</CardTitle>
          <CardDescription>
            Configure the resources for your deployment on {offering.name}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* CPU */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="cpu">CPU</Label>
              <span className="text-sm text-muted-foreground">
                {resources.cpu} {RESOURCE_LIMITS.cpu.unit}
              </span>
            </div>
            <Input
              id="cpu"
              type="range"
              min={RESOURCE_LIMITS.cpu.min}
              max={RESOURCE_LIMITS.cpu.max}
              step={RESOURCE_LIMITS.cpu.step}
              value={resources.cpu}
              onChange={(e) => onChange({ cpu: parseInt(e.target.value, 10) })}
              className="h-2 cursor-pointer"
            />
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>{RESOURCE_LIMITS.cpu.min} vCPU</span>
              <span>{RESOURCE_LIMITS.cpu.max} vCPU</span>
            </div>
          </div>

          {/* Memory */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="memory">Memory</Label>
              <span className="text-sm text-muted-foreground">
                {resources.memory} {RESOURCE_LIMITS.memory.unit}
              </span>
            </div>
            <Input
              id="memory"
              type="range"
              min={RESOURCE_LIMITS.memory.min}
              max={RESOURCE_LIMITS.memory.max}
              step={RESOURCE_LIMITS.memory.step}
              value={resources.memory}
              onChange={(e) => onChange({ memory: parseInt(e.target.value, 10) })}
              className="h-2 cursor-pointer"
            />
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>{RESOURCE_LIMITS.memory.min} GB</span>
              <span>{RESOURCE_LIMITS.memory.max} GB</span>
            </div>
          </div>

          {/* GPU (only for GPU/ML offerings) */}
          {(hasGpu || hasGpuPricing) && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="gpu">GPU</Label>
                <span className="text-sm text-muted-foreground">
                  {resources.gpu} {RESOURCE_LIMITS.gpu.unit}
                </span>
              </div>
              <Input
                id="gpu"
                type="range"
                min={RESOURCE_LIMITS.gpu.min}
                max={RESOURCE_LIMITS.gpu.max}
                step={RESOURCE_LIMITS.gpu.step}
                value={resources.gpu}
                onChange={(e) => onChange({ gpu: parseInt(e.target.value, 10) })}
                className="h-2 cursor-pointer"
              />
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>{RESOURCE_LIMITS.gpu.min}</span>
                <span>{RESOURCE_LIMITS.gpu.max} GPU</span>
              </div>
              {offering.specifications?.gpu && (
                <p className="text-xs text-muted-foreground">
                  GPU Model: {offering.specifications.gpu}
                </p>
              )}
            </div>
          )}

          {/* Storage */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="storage">Storage</Label>
              <span className="text-sm text-muted-foreground">
                {resources.storage} {RESOURCE_LIMITS.storage.unit}
              </span>
            </div>
            <Input
              id="storage"
              type="range"
              min={RESOURCE_LIMITS.storage.min}
              max={RESOURCE_LIMITS.storage.max}
              step={RESOURCE_LIMITS.storage.step}
              value={resources.storage}
              onChange={(e) => onChange({ storage: parseInt(e.target.value, 10) })}
              className="h-2 cursor-pointer"
            />
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>{RESOURCE_LIMITS.storage.min} GB</span>
              <span>{RESOURCE_LIMITS.storage.max} GB</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Duration &amp; Region</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Duration */}
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="duration">Duration</Label>
              <Input
                id="duration"
                type="number"
                min={RESOURCE_LIMITS.duration.min}
                max={RESOURCE_LIMITS.duration.max}
                value={resources.duration}
                onChange={(e) =>
                  onChange({ duration: Math.max(1, parseInt(e.target.value, 10) || 1) })
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="duration-unit">Unit</Label>
              <Select
                value={resources.durationUnit}
                onValueChange={(value) =>
                  onChange({ durationUnit: value as 'hours' | 'days' | 'months' })
                }
              >
                <SelectTrigger id="duration-unit">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="hours">Hours</SelectItem>
                  <SelectItem value="days">Days</SelectItem>
                  <SelectItem value="months">Months</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Region */}
          {offering.regions && offering.regions.length > 0 && (
            <div className="space-y-2">
              <Label htmlFor="region">Region</Label>
              <Select
                value={resources.region}
                onValueChange={(value) => onChange({ region: value })}
              >
                <SelectTrigger id="region">
                  <SelectValue placeholder="Select a region" />
                </SelectTrigger>
                <SelectContent>
                  {offering.regions.map((region) => (
                    <SelectItem key={region} value={region}>
                      {region}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Validation Errors */}
      {validationErrors.length > 0 && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
          <p className="text-sm font-medium text-destructive">Please fix the following:</p>
          <ul className="mt-2 list-inside list-disc text-sm text-destructive">
            {validationErrors.map((err) => (
              <li key={err}>{err}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
