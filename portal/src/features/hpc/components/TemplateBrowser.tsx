'use client';

/**
 * Template Browser Component
 *
 * Displays workload templates with filtering.
 */

import Link from 'next/link';
import { useState } from 'react';
import { useWorkloadTemplates } from '@/features/hpc';
import type { WorkloadCategory } from '@/features/hpc';

export function TemplateBrowser() {
  const [selectedCategory, setSelectedCategory] = useState<WorkloadCategory | 'all'>('all');
  const { templates, isLoading, error } = useWorkloadTemplates();

  const filteredTemplates =
    selectedCategory === 'all'
      ? templates
      : templates.filter((t) => t.category === selectedCategory);

  if (error) {
    return (
      <div className="rounded-lg border border-destructive bg-destructive/10 p-4 text-destructive">
        <p>Error loading templates: {error.message}</p>
      </div>
    );
  }

  return (
    <div>
      {/* Category Filters */}
      <div className="mb-6 flex flex-wrap gap-2">
        <CategoryButton
          label="All"
          active={selectedCategory === 'all'}
          onClick={() => setSelectedCategory('all')}
        />
        <CategoryButton
          label="Machine Learning"
          active={selectedCategory === 'ml_training'}
          onClick={() => setSelectedCategory('ml_training')}
        />
        <CategoryButton
          label="Scientific Computing"
          active={selectedCategory === 'scientific'}
          onClick={() => setSelectedCategory('scientific')}
        />
        <CategoryButton
          label="Rendering"
          active={selectedCategory === 'rendering'}
          onClick={() => setSelectedCategory('rendering')}
        />
        <CategoryButton
          label="Data Processing"
          active={selectedCategory === 'data_processing'}
          onClick={() => setSelectedCategory('data_processing')}
        />
      </div>

      {/* Templates Grid */}
      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="h-48 animate-pulse rounded-lg bg-muted" />
          ))}
        </div>
      ) : filteredTemplates.length === 0 ? (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <p className="text-muted-foreground">No templates found in this category</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filteredTemplates.map((template) => (
            <TemplateCard key={template.id} template={template} />
          ))}
        </div>
      )}
    </div>
  );
}

function CategoryButton({
  label,
  active,
  onClick,
}: {
  label: string;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full px-4 py-1 text-sm ${
        active ? 'bg-primary text-primary-foreground' : 'border border-border hover:bg-accent'
      }`}
    >
      {label}
    </button>
  );
}

function TemplateCard({
  template,
}: {
  template: ReturnType<typeof useWorkloadTemplates>['templates'][0];
}) {
  const categoryLabels: Record<WorkloadCategory, string> = {
    ml_training: 'Machine Learning',
    ml_inference: 'ML Inference',
    scientific: 'Scientific Computing',
    rendering: 'Rendering',
    simulation: 'Simulation',
    data_processing: 'Data Processing',
    custom: 'Custom',
  };

  const hasGPU = (template.defaultResources.gpusPerNode ?? 0) > 0;

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-start justify-between">
        <span className="rounded-full bg-muted px-2 py-1 text-xs text-muted-foreground">
          {categoryLabels[template.category]}
        </span>
        {hasGPU && (
          <span className="rounded-full bg-primary/10 px-2 py-1 text-xs text-primary">GPU</span>
        )}
      </div>

      <h3 className="mt-4 text-lg font-semibold">{template.name}</h3>
      <p className="mt-2 text-sm text-muted-foreground">{template.description}</p>

      <div className="mt-4 text-sm">
        <span className="text-muted-foreground">Est. cost: </span>
        <span className="font-medium">${template.estimatedCostPerHour}/hr</span>
      </div>

      <Link
        href={`/hpc/jobs/new?template=${template.id}`}
        className="mt-4 block w-full rounded-lg border border-border py-2 text-center text-sm hover:bg-accent"
      >
        Use Template
      </Link>
    </div>
  );
}
