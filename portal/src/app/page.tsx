import Link from 'next/link';
import { MarketingLayout } from '@/layouts';

export default function HomePage() {
  return (
    <MarketingLayout>
      <section className="relative flex flex-1 flex-col items-center justify-center px-4 py-12 sm:py-20">
        <div className="absolute inset-0 -z-10 bg-gradient-to-b from-primary/5 via-transparent to-transparent" />

        <div className="text-center">
          <h1 className="text-balance text-3xl font-bold tracking-tight sm:text-4xl md:text-6xl">
            <span className="gradient-text">VirtEngine</span> Portal
          </h1>
          <p className="mx-auto mt-4 max-w-2xl text-base leading-7 text-muted-foreground sm:mt-6 sm:text-lg sm:leading-8">
            Decentralized cloud computing marketplace with ML-powered identity verification. Deploy
            workloads, manage HPC jobs, and access compute resources securely.
          </p>

          <div className="mt-8 flex flex-col items-center gap-3 sm:mt-10 sm:flex-row sm:justify-center sm:gap-4">
            <Link
              href="/connect"
              className="inline-flex w-full items-center justify-center rounded-lg bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary active:bg-primary/80 sm:w-auto"
            >
              Connect Wallet
            </Link>
            <Link
              href="/marketplace"
              className="inline-flex w-full items-center justify-center rounded-lg border border-border bg-background px-6 py-3 text-sm font-semibold shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent/80 sm:w-auto"
            >
              Browse Marketplace
            </Link>
          </div>
        </div>

        <div className="mt-12 grid w-full max-w-5xl gap-4 px-2 sm:mt-20 sm:grid-cols-2 sm:gap-6 sm:px-4 lg:grid-cols-3">
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
    </MarketingLayout>
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
      className="card-hover group rounded-xl border border-border bg-card p-5 shadow-sm transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 active:scale-[0.98] sm:p-6"
    >
      <div className="mb-3 text-2xl sm:mb-4 sm:text-3xl">{icon}</div>
      <h2 className="text-base font-semibold group-hover:text-primary sm:text-lg">{title}</h2>
      <p className="mt-1.5 text-sm text-muted-foreground sm:mt-2">{description}</p>
    </Link>
  );
}
