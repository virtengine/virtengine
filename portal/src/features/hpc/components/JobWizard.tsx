'use client';

/**
 * Job Submission Wizard
 *
 * Multi-step wizard for submitting HPC jobs.
 */

import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import {
  useCostEstimation,
  useJobSubmission,
  useWorkloadTemplate,
  useWorkloadTemplates,
} from '@/features/hpc';
import { useWizardStore } from '../stores/wizardStore';
import type { JobResources } from '../types';

export function JobWizard() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const templateIdParam = searchParams.get('template');

  const {
    currentStep,
    selectedTemplate,
    manifest,
    estimatedCost,
    errors,
    nextStep,
    prevStep,
    selectTemplate,
    updateManifest,
    setEstimatedCost,
    reset,
  } = useWizardStore();

  const { template: preselectedTemplate } = useWorkloadTemplate(templateIdParam);
  const { submitJob, isSubmitting, error: submitError } = useJobSubmission();
  const { estimateCost, isEstimating } = useCostEstimation();

  // Preselect template from URL param
  useEffect(() => {
    if (preselectedTemplate && !selectedTemplate) {
      selectTemplate(preselectedTemplate);
    }
  }, [preselectedTemplate, selectedTemplate, selectTemplate]);

  // Estimate cost when resources change
  useEffect(() => {
    if (manifest.resources && currentStep === 'review') {
      estimateCost('offering-1', manifest.resources as JobResources)
        .then((cost) => {
          setEstimatedCost({
            total: cost.estimatedTotal,
            perHour: cost.pricePerHour,
            breakdown: cost.breakdown,
            denom: cost.denom,
          });
        })
        .catch(console.error);
    }
  }, [manifest.resources, currentStep, estimateCost, setEstimatedCost]);

  const handleSubmit = async () => {
    if (!manifest.name || !manifest.resources) {
      return;
    }

    try {
      const result = await submitJob({
        offeringId: 'offering-1',
        name: manifest.name,
        description: manifest.description,
        templateId: manifest.templateId,
        resources: manifest.resources as JobResources,
        command: manifest.command,
        containerImage: manifest.image,
        environment: manifest.environment,
      });

      // Success - redirect to job detail
      reset();
      router.push(`/hpc/jobs/${result.jobId}`);
    } catch (err) {
      console.error('Job submission failed:', err);
    }
  };

  return (
    <div className="mx-auto max-w-3xl">
      {/* Progress Indicator */}
      <WizardProgress currentStep={currentStep} />

      {/* Step Content */}
      <div className="mt-8">
        {currentStep === 'template' && <TemplateStep />}
        {currentStep === 'configure' && <ConfigureStep />}
        {currentStep === 'resources' && <ResourcesStep />}
        {currentStep === 'review' && (
          <ReviewStep estimatedCost={estimatedCost} isEstimating={isEstimating} />
        )}
      </div>

      {/* Navigation */}
      <div className="mt-8 flex gap-4">
        {currentStep !== 'template' && (
          <button
            type="button"
            onClick={prevStep}
            disabled={isSubmitting}
            className="flex-1 rounded-lg border border-border px-4 py-3 text-sm hover:bg-accent disabled:opacity-50"
          >
            Back
          </button>
        )}
        {currentStep === 'review' ? (
          <button
            type="button"
            onClick={handleSubmit}
            disabled={isSubmitting || !manifest.name || !manifest.resources}
            className="flex-1 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            {isSubmitting ? 'Submitting...' : 'Submit Job'}
          </button>
        ) : (
          <button
            type="button"
            onClick={nextStep}
            disabled={!canProceed(currentStep, manifest, selectedTemplate)}
            className="flex-1 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
          >
            Continue
          </button>
        )}
      </div>

      {submitError && (
        <div className="mt-4 rounded-lg border border-destructive bg-destructive/10 p-4 text-sm text-destructive">
          {submitError.message}
        </div>
      )}
    </div>
  );
}

function WizardProgress({ currentStep }: { currentStep: string }) {
  const steps = [
    { id: 'template', label: 'Template' },
    { id: 'configure', label: 'Configure' },
    { id: 'resources', label: 'Resources' },
    { id: 'review', label: 'Review' },
  ];

  const currentIndex = steps.findIndex((s) => s.id === currentStep);

  return (
    <div className="flex items-center gap-2">
      {steps.map((step, index) => (
        <div key={step.id} className="flex flex-1 items-center gap-2">
          <div
            className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium ${
              index <= currentIndex
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground'
            }`}
          >
            {index + 1}
          </div>
          <span
            className={`text-sm ${index <= currentIndex ? 'font-medium' : 'text-muted-foreground'}`}
          >
            {step.label}
          </span>
          {index < steps.length - 1 && <div className="h-px flex-1 bg-border" />}
        </div>
      ))}
    </div>
  );
}

function TemplateStep() {
  const { selectedTemplate, selectTemplate } = useWizardStore();
  const { templates, isLoading } = useWorkloadTemplates();

  if (isLoading) {
    return <div className="text-center text-muted-foreground">Loading templates...</div>;
  }

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h2 className="text-lg font-semibold">Select Template</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        Choose a workload template or create a custom job
      </p>

      <div className="mt-6 space-y-3">
        {templates.map((template) => (
          <TemplateOption
            key={template.id}
            template={template}
            selected={selectedTemplate?.id === template.id}
            onSelect={() => selectTemplate(template)}
          />
        ))}
        <TemplateOption
          template={null}
          selected={selectedTemplate === null}
          onSelect={() => selectTemplate(null)}
        />
      </div>
    </div>
  );
}

function TemplateOption({
  template,
  selected,
  onSelect,
}: {
  template: ReturnType<typeof useWorkloadTemplates>['templates'][0] | null;
  selected: boolean;
  onSelect: () => void;
}) {
  return (
    <label
      className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
        selected ? 'border-primary bg-primary/5' : 'border-border hover:bg-accent'
      }`}
    >
      <input type="radio" checked={selected} onChange={onSelect} className="h-4 w-4 text-primary" />
      <div className="flex-1">
        <div className="font-medium">{template ? template.name : 'Custom Workload'}</div>
        <div className="text-sm text-muted-foreground">
          {template ? template.description : 'Build from scratch with custom parameters'}
        </div>
      </div>
    </label>
  );
}

function ConfigureStep() {
  const { manifest, updateManifest } = useWizardStore();

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h2 className="text-lg font-semibold">Job Configuration</h2>

      <div className="mt-6 space-y-4">
        <div>
          <label htmlFor="job-name" className="block text-sm font-medium">
            Job Name *
          </label>
          <input
            type="text"
            id="job-name"
            value={manifest.name ?? ''}
            onChange={(e) => updateManifest({ name: e.target.value })}
            placeholder="my-training-job"
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium">
            Description (optional)
          </label>
          <textarea
            id="description"
            value={manifest.description ?? ''}
            onChange={(e) => updateManifest({ description: e.target.value })}
            rows={3}
            placeholder="Brief description of the job..."
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>
      </div>
    </div>
  );
}

function ResourcesStep() {
  const { manifest, selectedTemplate, updateManifest } = useWizardStore();
  const resources = manifest.resources ??
    selectedTemplate?.defaultResources ?? {
      nodes: 1,
      cpusPerNode: 8,
      memoryGBPerNode: 64,
      maxRuntimeSeconds: 86400,
      storageGB: 100,
    };

  const updateResources = (updates: Partial<JobResources>) => {
    updateManifest({
      resources: { ...resources, ...updates } as JobResources,
    });
  };

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h2 className="text-lg font-semibold">Resource Requirements</h2>

      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <div>
          <label htmlFor="nodes" className="block text-sm font-medium">
            Nodes
          </label>
          <input
            type="number"
            id="nodes"
            value={resources.nodes}
            onChange={(e) => updateResources({ nodes: parseInt(e.target.value) })}
            min={1}
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>

        <div>
          <label htmlFor="cpus" className="block text-sm font-medium">
            CPUs per Node
          </label>
          <input
            type="number"
            id="cpus"
            value={resources.cpusPerNode}
            onChange={(e) => updateResources({ cpusPerNode: parseInt(e.target.value) })}
            min={1}
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>

        <div>
          <label htmlFor="memory" className="block text-sm font-medium">
            Memory (GB) per Node
          </label>
          <input
            type="number"
            id="memory"
            value={resources.memoryGBPerNode}
            onChange={(e) => updateResources({ memoryGBPerNode: parseInt(e.target.value) })}
            min={1}
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>

        <div>
          <label htmlFor="storage" className="block text-sm font-medium">
            Storage (GB)
          </label>
          <input
            type="number"
            id="storage"
            value={resources.storageGB}
            onChange={(e) => updateResources({ storageGB: parseInt(e.target.value) })}
            min={1}
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>

        <div className="sm:col-span-2">
          <label htmlFor="runtime" className="block text-sm font-medium">
            Maximum Runtime (hours)
          </label>
          <input
            type="number"
            id="runtime"
            value={resources.maxRuntimeSeconds / 3600}
            onChange={(e) =>
              updateResources({ maxRuntimeSeconds: parseInt(e.target.value) * 3600 })
            }
            min={1}
            className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>
      </div>
    </div>
  );
}

function ReviewStep({
  estimatedCost,
  isEstimating,
}: {
  estimatedCost: ReturnType<typeof useWizardStore>['estimatedCost'];
  isEstimating: boolean;
}) {
  const { manifest, selectedTemplate } = useWizardStore();

  return (
    <div className="space-y-6">
      {/* Job Summary */}
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold">Job Summary</h2>

        <div className="mt-4 space-y-3 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Name:</span>
            <span className="font-medium">{manifest.name}</span>
          </div>
          {selectedTemplate && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Template:</span>
              <span className="font-medium">{selectedTemplate.name}</span>
            </div>
          )}
          {manifest.description && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Description:</span>
              <span className="font-medium">{manifest.description}</span>
            </div>
          )}
        </div>
      </div>

      {/* Resources */}
      {manifest.resources && (
        <div className="rounded-lg border border-border bg-card p-6">
          <h3 className="font-semibold">Resources</h3>
          <div className="mt-4 grid gap-3 text-sm sm:grid-cols-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Nodes:</span>
              <span>{manifest.resources.nodes}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">CPUs:</span>
              <span>{manifest.resources.cpusPerNode} per node</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Memory:</span>
              <span>{manifest.resources.memoryGBPerNode} GB per node</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Storage:</span>
              <span>{manifest.resources.storageGB} GB</span>
            </div>
            <div className="flex justify-between sm:col-span-2">
              <span className="text-muted-foreground">Max Runtime:</span>
              <span>{manifest.resources.maxRuntimeSeconds / 3600} hours</span>
            </div>
          </div>
        </div>
      )}

      {/* Cost Estimate */}
      <div className="rounded-lg border border-primary/50 bg-primary/5 p-6">
        <h3 className="font-semibold">Cost Estimate</h3>

        {isEstimating ? (
          <div className="mt-4 text-sm text-muted-foreground">Calculating...</div>
        ) : estimatedCost ? (
          <div className="mt-4">
            <div className="flex items-baseline justify-between">
              <span className="text-muted-foreground">Estimated total</span>
              <div>
                <span className="text-3xl font-bold">${estimatedCost.total}</span>
                <span className="ml-2 text-sm text-muted-foreground">{estimatedCost.denom}</span>
              </div>
            </div>
            <div className="mt-2 text-sm text-muted-foreground">${estimatedCost.perHour}/hour</div>
          </div>
        ) : null}
      </div>
    </div>
  );
}

function canProceed(step: string, manifest: any, template: any): boolean {
  switch (step) {
    case 'template':
      return true; // Can always proceed from template selection
    case 'configure':
      return !!manifest.name && manifest.name.trim().length > 0;
    case 'resources':
      return !!manifest.resources;
    default:
      return false;
  }
}
