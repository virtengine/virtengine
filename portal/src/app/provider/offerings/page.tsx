import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Provider Offerings',
  description: 'Manage your compute offerings',
};

export default function ProviderOfferingsPage() {
  return (
    <div className="container py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Your Offerings</h1>
          <p className="mt-1 text-muted-foreground">
            Manage your compute resource listings
          </p>
        </div>
        <Link
          href="/provider/offerings/new"
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Create Offering
        </Link>
      </div>

      {/* Offerings Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <OfferingCard key={i} index={i} />
        ))}
      </div>

      {/* Empty state */}
      {false && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-4">
            <span className="text-4xl">📦</span>
          </div>
          <h2 className="mt-4 text-lg font-medium">No offerings yet</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Create your first offering to start accepting orders
          </p>
          <Link
            href="/provider/offerings/new"
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Create Offering
          </Link>
        </div>
      )}
    </div>
  );
}

function OfferingCard({ index }: { index: number }) {
  const offerings = [
    { type: 'GPU', name: 'NVIDIA A100 80GB', active: true, leases: 3 },
    { type: 'CPU', name: 'AMD EPYC 7763', active: true, leases: 8 },
    { type: 'HPC', name: 'HPC Compute Cluster', active: false, leases: 0 },
    { type: 'GPU', name: 'NVIDIA H100', active: true, leases: 2 },
    { type: 'Storage', name: 'NVMe Storage Pool', active: true, leases: 5 },
    { type: 'CPU', name: 'Intel Xeon Platinum', active: false, leases: 0 },
  ];
  const offering = offerings[index % offerings.length]!;

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-start justify-between">
        <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
          {offering.type}
        </span>
        <span className={`flex items-center gap-1 text-sm ${offering.active ? 'text-success' : 'text-muted-foreground'}`}>
          <span className={`status-dot ${offering.active ? 'status-dot-success' : ''}`} />
          {offering.active ? 'Active' : 'Inactive'}
        </span>
      </div>
      <h3 className="mt-4 font-semibold">{offering.name}</h3>
      <p className="mt-1 text-sm text-muted-foreground">
        {offering.leases} active lease{offering.leases !== 1 ? 's' : ''}
      </p>
      <div className="mt-4 flex gap-2">
        <button
          type="button"
          className="flex-1 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
        >
          Edit
        </button>
        <button
          type="button"
          className="flex-1 rounded-lg border border-border px-3 py-2 text-sm hover:bg-accent"
        >
          {offering.active ? 'Pause' : 'Activate'}
        </button>
      </div>
    </div>
  );
}
