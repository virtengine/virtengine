import Link from 'next/link';

export default function HomePage() {
  return (
    <main id="main-content" className="flex min-h-screen flex-col">
      {/* Hero Section */}
      <section className="relative flex flex-1 flex-col items-center justify-center px-4 py-20">
        <div className="absolute inset-0 -z-10 bg-gradient-to-b from-primary/5 via-transparent to-transparent" />

        <div className="text-center">
          <h1 className="text-balance text-4xl font-bold tracking-tight sm:text-6xl">
            <span className="gradient-text">VirtEngine</span> Portal
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-muted-foreground">
            Decentralized cloud computing marketplace with ML-powered identity verification. Deploy
            workloads, manage HPC jobs, and access compute resources securely.
          </p>

          <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
            <Link
              href="/connect"
              className="inline-flex items-center justify-center rounded-lg bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              Connect Wallet
            </Link>
            <Link
              href="/marketplace"
              className="inline-flex items-center justify-center rounded-lg border border-border bg-background px-6 py-3 text-sm font-semibold shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              Browse Marketplace
            </Link>
          </div>
        </div>

        {/* Feature Cards */}
        <div className="mt-20 grid max-w-5xl gap-6 px-4 sm:grid-cols-2 lg:grid-cols-3">
          <FeatureCard
            title="Marketplace"
            description="Browse and purchase compute resources from providers worldwide."
            href="/marketplace"
            icon="🛒"
          />
          <FeatureCard
            title="Identity (VEID)"
            description="Complete identity verification with ML-powered scoring."
            href="/identity"
            icon="🔐"
          />
          <FeatureCard
            title="HPC Jobs"
            description="Submit and manage high-performance computing workloads."
            href="/hpc/jobs"
            icon="⚡"
          />
          <FeatureCard
            title="Provider Console"
            description="Manage your offerings and infrastructure as a provider."
            href="/provider/dashboard"
            icon="🖥️"
          />
          <FeatureCard
            title="Governance"
            description="Participate in protocol governance and voting."
            href="/governance/proposals"
            icon="🗳️"
          />
          <FeatureCard
            title="Orders"
            description="Track your orders and manage deployments."
            href="/orders"
            icon="📋"
          />
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border px-4 py-8">
        <div className="mx-auto max-w-7xl text-center text-sm text-muted-foreground">
          <p>© {new Date().getFullYear()} VirtEngine. All rights reserved.</p>
          <div className="mt-4 flex justify-center gap-6">
            <a
              href="https://docs.virtengine.com"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground"
            >
              Documentation
            </a>
            <a
              href="https://support.virtengine.com"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground"
            >
              Support
            </a>
            <a
              href="https://github.com/virtengine"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-foreground"
            >
              GitHub
            </a>
          </div>
        </div>
      </footer>
    </main>
  );
}

interface FeatureCardProps {
  title: string;
  description: string;
  href: string;
  icon: string;
}

function FeatureCard({ title, description, href, icon }: FeatureCardProps) {
  return (
    <Link
      href={href}
      className="card-hover group rounded-xl border border-border bg-card p-6 shadow-sm transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
    >
      <div className="mb-4 text-3xl">{icon}</div>
      <h2 className="text-lg font-semibold group-hover:text-primary">{title}</h2>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
    </Link>
  );
}
