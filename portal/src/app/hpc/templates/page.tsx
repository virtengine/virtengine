import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'HPC Templates',
  description: 'Browse HPC workload templates',
};

export default function HPCTemplatesPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Workload Templates</h1>
        <p className="mt-1 text-muted-foreground">
          Pre-configured templates for common HPC workloads
        </p>
      </div>

      {/* Categories */}
      <div className="mb-6 flex flex-wrap gap-2">
        <button type="button" className="rounded-full bg-primary px-4 py-1 text-sm text-primary-foreground">
          All
        </button>
        <button type="button" className="rounded-full border border-border px-4 py-1 text-sm hover:bg-accent">
          Machine Learning
        </button>
        <button type="button" className="rounded-full border border-border px-4 py-1 text-sm hover:bg-accent">
          Scientific Computing
        </button>
        <button type="button" className="rounded-full border border-border px-4 py-1 text-sm hover:bg-accent">
          Data Processing
        </button>
        <button type="button" className="rounded-full border border-border px-4 py-1 text-sm hover:bg-accent">
          Rendering
        </button>
      </div>

      {/* Templates Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <TemplateCard
          name="PyTorch Training"
          category="Machine Learning"
          description="Train deep learning models with PyTorch. Supports distributed training across multiple GPUs."
          gpuRequired
        />
        <TemplateCard
          name="TensorFlow"
          category="Machine Learning"
          description="TensorFlow training pipeline with Keras integration and TensorBoard support."
          gpuRequired
        />
        <TemplateCard
          name="JAX/Flax"
          category="Machine Learning"
          description="High-performance ML research with JAX and Flax."
          gpuRequired
        />
        <TemplateCard
          name="OpenFOAM"
          category="Scientific Computing"
          description="Computational fluid dynamics simulations with OpenFOAM."
        />
        <TemplateCard
          name="GROMACS"
          category="Scientific Computing"
          description="Molecular dynamics simulations for computational chemistry."
          gpuRequired
        />
        <TemplateCard
          name="AlphaFold"
          category="Scientific Computing"
          description="Protein structure prediction with AlphaFold 2."
          gpuRequired
        />
        <TemplateCard
          name="Apache Spark"
          category="Data Processing"
          description="Large-scale data processing with Apache Spark."
        />
        <TemplateCard
          name="Dask"
          category="Data Processing"
          description="Parallel computing in Python with Dask."
        />
        <TemplateCard
          name="Blender Render"
          category="Rendering"
          description="3D rendering and animation with Blender."
          gpuRequired
        />
      </div>
    </div>
  );
}

interface TemplateCardProps {
  name: string;
  category: string;
  description: string;
  gpuRequired?: boolean;
}

function TemplateCard({ name, category, description, gpuRequired }: TemplateCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-start justify-between">
        <span className="rounded-full bg-muted px-2 py-1 text-xs text-muted-foreground">
          {category}
        </span>
        {gpuRequired && (
          <span className="rounded-full bg-primary/10 px-2 py-1 text-xs text-primary">
            GPU
          </span>
        )}
      </div>
      <h3 className="mt-4 text-lg font-semibold">{name}</h3>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
      <button
        type="button"
        className="mt-4 w-full rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
      >
        Use Template
      </button>
    </div>
  );
}
