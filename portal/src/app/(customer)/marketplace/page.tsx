import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Marketplace',
  description: 'Browse compute offerings from providers',
};

export default function MarketplacePage() {
  return (
    <div className="container py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Marketplace</h1>
          <p className="mt-1 text-muted-foreground">
            Browse and purchase compute resources from providers worldwide
          </p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
          >
            Filters
          </button>
          <select
            className="rounded-lg border border-border bg-background px-4 py-2 text-sm"
            aria-label="Sort offerings"
          >
            <option>Price: Low to High</option>
            <option>Price: High to Low</option>
            <option>Newest First</option>
            <option>Rating</option>
          </select>
        </div>
      </div>

      {/* Filter sidebar would go here */}
      <div className="grid gap-6 lg:grid-cols-4">
        <aside className="hidden lg:block">
          <div className="sticky top-4 space-y-6 rounded-lg border border-border p-4">
            <FilterSection title="Resource Type">
              <FilterCheckbox label="CPU" count={42} />
              <FilterCheckbox label="GPU" count={18} />
              <FilterCheckbox label="Storage" count={24} />
              <FilterCheckbox label="HPC Cluster" count={8} />
            </FilterSection>

            <FilterSection title="Region">
              <FilterCheckbox label="North America" count={32} />
              <FilterCheckbox label="Europe" count={28} />
              <FilterCheckbox label="Asia Pacific" count={15} />
            </FilterSection>

            <FilterSection title="Provider Tier">
              <FilterCheckbox label="Verified" count={45} />
              <FilterCheckbox label="Standard" count={38} />
            </FilterSection>
          </div>
        </aside>

        {/* Offerings Grid */}
        <div className="lg:col-span-3">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {/* Placeholder offering cards */}
            {Array.from({ length: 9 }, (_, index) => ({
              id: `offering-${index + 1}`,
              index,
            })).map((offering) => (
              <OfferingCardPlaceholder key={offering.id} index={offering.index} />
            ))}
          </div>

          {/* Pagination */}
          <div className="mt-8 flex justify-center gap-2">
            <button
              type="button"
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
              disabled
            >
              Previous
            </button>
            <button
              type="button"
              className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground"
            >
              1
            </button>
            <button
              type="button"
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
            >
              2
            </button>
            <button
              type="button"
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
            >
              3
            </button>
            <button
              type="button"
              className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function FilterSection({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <h3 className="mb-3 font-medium">{title}</h3>
      <div className="space-y-2">{children}</div>
    </div>
  );
}

function FilterCheckbox({ label, count }: { label: string; count: number }) {
  return (
    <label className="flex cursor-pointer items-center justify-between text-sm">
      <span className="flex items-center gap-2">
        <input type="checkbox" className="rounded border-border" />
        {label}
      </span>
      <span className="text-muted-foreground">{count}</span>
    </label>
  );
}

function OfferingCardPlaceholder({ index }: { index: number }) {
  const offerings = [
    { type: 'GPU', name: 'NVIDIA A100 Cluster', price: '2.50', provider: 'CloudCore' },
    { type: 'CPU', name: 'AMD EPYC 7763', price: '0.45', provider: 'DataNexus' },
    { type: 'HPC', name: 'HPC Compute Node', price: '8.00', provider: 'SuperCloud' },
  ];
  const offering = offerings[index % offerings.length]!;

  return (
    <Link
      href={`/marketplace/${index}`}
      className="group rounded-lg border border-border bg-card p-4 transition-all card-hover"
    >
      <div className="flex items-start justify-between">
        <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
          {offering.type}
        </span>
        <span className="flex items-center gap-1 text-sm text-muted-foreground">
          <span className="status-dot status-dot-success" />
          Available
        </span>
      </div>
      <h3 className="mt-4 font-semibold group-hover:text-primary">{offering.name}</h3>
      <p className="mt-1 text-sm text-muted-foreground">by {offering.provider}</p>
      <div className="mt-4 flex items-baseline justify-between">
        <span className="text-lg font-bold">${offering.price}</span>
        <span className="text-sm text-muted-foreground">/hour</span>
      </div>
    </Link>
  );
}
