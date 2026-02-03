import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Submit HPC Job',
  description: 'Submit a new HPC job',
};

export default function NewHPCJobPage() {
  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link
          href="/hpc/jobs"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to Jobs
        </Link>
      </div>

      <div className="mx-auto max-w-3xl">
        <div className="mb-8">
          <h1 className="text-3xl font-bold">Submit New Job</h1>
          <p className="mt-1 text-muted-foreground">
            Configure and submit a new HPC workload
          </p>
        </div>

        <form className="space-y-8">
          {/* Template Selection */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">1. Select Template</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Choose a workload template or start from scratch
            </p>

            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              <TemplateOption
                name="PyTorch Training"
                description="ML model training with PyTorch"
                selected
              />
              <TemplateOption
                name="TensorFlow"
                description="TensorFlow training pipeline"
              />
              <TemplateOption
                name="OpenFOAM"
                description="CFD simulation"
              />
              <TemplateOption
                name="Custom"
                description="Build from scratch"
              />
            </div>

            <Link
              href="/hpc/templates"
              className="mt-4 inline-block text-sm text-primary hover:underline"
            >
              Browse all templates →
            </Link>
          </div>

          {/* Job Configuration */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">2. Job Configuration</h2>
            
            <div className="mt-4 space-y-4">
              <div>
                <label htmlFor="job-name" className="text-sm font-medium">
                  Job Name
                </label>
                <input
                  type="text"
                  id="job-name"
                  placeholder="my-training-job"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                />
              </div>

              <div>
                <label htmlFor="description" className="text-sm font-medium">
                  Description (optional)
                </label>
                <textarea
                  id="description"
                  rows={2}
                  placeholder="Brief description of the job..."
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                />
              </div>
            </div>
          </div>

          {/* Resource Requirements */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">3. Resource Requirements</h2>
            
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <div>
                <label htmlFor="gpus" className="text-sm font-medium">
                  GPUs
                </label>
                <select
                  id="gpus"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                >
                  <option>1x NVIDIA A100</option>
                  <option>2x NVIDIA A100</option>
                  <option>4x NVIDIA A100</option>
                  <option>8x NVIDIA A100</option>
                  <option>1x NVIDIA H100</option>
                </select>
              </div>

              <div>
                <label htmlFor="memory" className="text-sm font-medium">
                  Memory
                </label>
                <select
                  id="memory"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                >
                  <option>32 GB</option>
                  <option>64 GB</option>
                  <option>128 GB</option>
                  <option>256 GB</option>
                </select>
              </div>

              <div>
                <label htmlFor="cpus" className="text-sm font-medium">
                  CPU Cores
                </label>
                <select
                  id="cpus"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                >
                  <option>8 cores</option>
                  <option>16 cores</option>
                  <option>32 cores</option>
                  <option>64 cores</option>
                </select>
              </div>

              <div>
                <label htmlFor="storage" className="text-sm font-medium">
                  Storage
                </label>
                <select
                  id="storage"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                >
                  <option>100 GB SSD</option>
                  <option>500 GB SSD</option>
                  <option>1 TB SSD</option>
                  <option>2 TB SSD</option>
                </select>
              </div>
            </div>
          </div>

          {/* Execution */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">4. Execution</h2>
            
            <div className="mt-4 space-y-4">
              <div>
                <label htmlFor="container" className="text-sm font-medium">
                  Container Image
                </label>
                <input
                  type="text"
                  id="container"
                  defaultValue="pytorch/pytorch:2.1.0-cuda12.1-cudnn8-runtime"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
                />
              </div>

              <div>
                <label htmlFor="command" className="text-sm font-medium">
                  Command
                </label>
                <textarea
                  id="command"
                  rows={3}
                  defaultValue="python train.py --epochs 100 --batch-size 32"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 font-mono text-sm"
                />
              </div>

              <div>
                <label htmlFor="max-runtime" className="text-sm font-medium">
                  Maximum Runtime
                </label>
                <select
                  id="max-runtime"
                  className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                >
                  <option>1 hour</option>
                  <option>4 hours</option>
                  <option>12 hours</option>
                  <option>24 hours</option>
                  <option>48 hours</option>
                  <option>72 hours</option>
                </select>
              </div>
            </div>
          </div>

          {/* Cost Estimate */}
          <div className="rounded-lg border border-primary/50 bg-primary/5 p-6">
            <h2 className="text-lg font-semibold">Cost Estimate</h2>
            <div className="mt-4 flex items-baseline justify-between">
              <span className="text-muted-foreground">Estimated cost</span>
              <div>
                <span className="text-3xl font-bold">$12.50</span>
                <span className="text-muted-foreground">/hour</span>
              </div>
            </div>
            <p className="mt-2 text-sm text-muted-foreground">
              Based on selected resources. Actual cost may vary.
            </p>
          </div>

          {/* Submit */}
          <div className="flex gap-4">
            <Link
              href="/hpc/jobs"
              className="flex-1 rounded-lg border border-border px-4 py-3 text-center text-sm hover:bg-accent"
            >
              Cancel
            </Link>
            <button
              type="submit"
              className="flex-1 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90"
            >
              Submit Job
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function TemplateOption({
  name,
  description,
  selected,
}: {
  name: string;
  description: string;
  selected?: boolean;
}) {
  return (
    <label
      className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
        selected ? 'border-primary bg-primary/5' : 'border-border hover:bg-accent'
      }`}
    >
      <input
        type="radio"
        name="template"
        defaultChecked={selected}
        className="h-4 w-4 text-primary"
        aria-label={name}
      />
      <div>
        <div className="font-medium">{name}</div>
        <div className="text-sm text-muted-foreground">{description}</div>
      </div>
    </label>
  );
}
