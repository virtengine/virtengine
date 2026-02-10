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
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
  const hasGpu = offering.category === 'gpu' || offering.category === 'ml';
  const hasGpuPricing = offering.prices?.some((p) => p.resourceType === 'gpu');

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('Compute Resources')}</CardTitle>
          <CardDescription>
            {t('Configure the resources for your deployment on {{offering}}', {
              offering: offering.name,
            })}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* CPU */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="cpu">{t('CPU')}</Label>
              <span className="text-sm text-muted-foreground">
                {t('{{count}} {{unit}}', {
                  count: resources.cpu,
                  unit: t(RESOURCE_LIMITS.cpu.unit),
                })}
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
              <span>{t('{{count}} vCPU', { count: RESOURCE_LIMITS.cpu.min })}</span>
              <span>{t('{{count}} vCPU', { count: RESOURCE_LIMITS.cpu.max })}</span>
            </div>
          </div>

          {/* Memory */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="memory">{t('Memory')}</Label>
              <span className="text-sm text-muted-foreground">
                {t('{{count}} {{unit}}', {
                  count: resources.memory,
                  unit: t(RESOURCE_LIMITS.memory.unit),
                })}
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
              <span>{t('{{count}} GB', { count: RESOURCE_LIMITS.memory.min })}</span>
              <span>{t('{{count}} GB', { count: RESOURCE_LIMITS.memory.max })}</span>
            </div>
          </div>

          {/* GPU (only for GPU/ML offerings) */}
          {(hasGpu || hasGpuPricing) && (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="gpu">{t('GPU')}</Label>
                <span className="text-sm text-muted-foreground">
                  {t('{{count}} {{unit}}', {
                    count: resources.gpu,
                    unit: t(RESOURCE_LIMITS.gpu.unit),
                  })}
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
                <span>{t('{{count}} GPU', { count: RESOURCE_LIMITS.gpu.max })}</span>
              </div>
              {offering.specifications?.gpu && (
                <p className="text-xs text-muted-foreground">
                  {t('GPU Model: {{model}}', { model: offering.specifications.gpu })}
                </p>
              )}
            </div>
          )}

          {/* Storage */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="storage">{t('Storage')}</Label>
              <span className="text-sm text-muted-foreground">
                {t('{{count}} {{unit}}', {
                  count: resources.storage,
                  unit: t(RESOURCE_LIMITS.storage.unit),
                })}
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
              <span>{t('{{count}} GB', { count: RESOURCE_LIMITS.storage.min })}</span>
              <span>{t('{{count}} GB', { count: RESOURCE_LIMITS.storage.max })}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('Duration & Region')}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Duration */}
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="duration">{t('Duration')}</Label>
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
              <Label htmlFor="duration-unit">{t('Unit')}</Label>
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
                  <SelectItem value="hours">{t('Hours')}</SelectItem>
                  <SelectItem value="days">{t('Days')}</SelectItem>
                  <SelectItem value="months">{t('Months')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Region */}
          {offering.regions && offering.regions.length > 0 && (
            <div className="space-y-2">
              <Label htmlFor="region">{t('Region')}</Label>
              <Select
                value={resources.region}
                onValueChange={(value) => onChange({ region: value })}
              >
                <SelectTrigger id="region">
                  <SelectValue placeholder={t('Select a region')} />
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
          <p className="text-sm font-medium text-destructive">{t('Please fix the following:')}</p>
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
