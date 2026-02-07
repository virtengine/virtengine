'use client';

import Link from 'next/link';
import { MarketingLayout } from '@/layouts';
import { useTranslation } from 'react-i18next';

export default function HomePage() {
  const { t } = useTranslation();

  return (
    <MarketingLayout>
      <section className="relative flex flex-1 flex-col items-center justify-center px-4 py-20">
        <div className="absolute inset-0 -z-10 bg-gradient-to-b from-primary/5 via-transparent to-transparent" />

        <div className="text-center">
          <h1 className="text-balance text-4xl font-bold tracking-tight sm:text-6xl">
            <span className="gradient-text">{t('VirtEngine')}</span> {t('Portal')}
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-muted-foreground">
            {t(
              'Decentralized cloud computing marketplace with ML-powered identity verification. Deploy workloads, manage HPC jobs, and access compute resources securely.'
            )}
          </p>

          <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
            <Link
              href="/connect"
              className="inline-flex items-center justify-center rounded-lg bg-primary px-6 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
            >
              {t('Connect Wallet')}
            </Link>
            <Link
              href="/marketplace"
              className="inline-flex items-center justify-center rounded-lg border border-border bg-background px-6 py-3 text-sm font-semibold shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              {t('Browse Marketplace')}
            </Link>
          </div>
        </div>

        <div className="mt-20 grid max-w-5xl gap-6 px-4 sm:grid-cols-2 lg:grid-cols-3">
          <FeatureCard
            title={t('Marketplace')}
            description={t('Browse and purchase compute resources from providers worldwide.')}
            href="/marketplace"
            icon="🛒"
          />
          <FeatureCard
            title={t('Identity (VEID)')}
            description={t('Complete identity verification with ML-powered scoring.')}
            href="/identity"
            icon="🔐"
          />
          <FeatureCard
            title={t('HPC Jobs')}
            description={t('Submit and manage high-performance computing workloads.')}
            href="/hpc/jobs"
            icon="⚡"
          />
          <FeatureCard
            title={t('Provider Console')}
            description={t('Manage your offerings and infrastructure as a provider.')}
            href="/provider/dashboard"
            icon="🖥️"
          />
          <FeatureCard
            title={t('Governance')}
            description={t('Participate in protocol governance and voting.')}
            href="/governance/proposals"
            icon="🗳️"
          />
          <FeatureCard
            title={t('Orders')}
            description={t('Track your orders and manage deployments.')}
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
      className="card-hover group rounded-xl border border-border bg-card p-6 shadow-sm transition-all focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
    >
      <div className="mb-4 text-3xl" aria-hidden="true">
        {icon}
      </div>
      <h2 className="text-lg font-semibold group-hover:text-primary">{title}</h2>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
    </Link>
  );
}
