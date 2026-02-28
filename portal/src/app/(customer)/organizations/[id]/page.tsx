import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Organization Details',
  description: 'View and manage your organization',
};

export default async function OrganizationDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link
          href="/organizations"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to organizations
        </Link>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
              <span className="text-xl font-semibold text-primary">O</span>
            </div>
            <div>
              <h1 className="text-2xl font-bold">Organization</h1>
              <p className="text-sm text-muted-foreground">ID: {id}</p>
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold">Live organization data unavailable</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          Membership, billing, and organization settings for this route are not exposed by the
          live portal APIs configured for this deployment. This page intentionally avoids synthetic
          member lists or spend totals.
        </p>
        <div className="mt-6 grid gap-4 sm:grid-cols-3">
          <div className="rounded-lg border border-border bg-muted/40 p-4">
            <p className="text-sm text-muted-foreground">Organization ID</p>
            <p className="mt-1 font-mono text-sm">{id}</p>
          </div>
          <div className="rounded-lg border border-border bg-muted/40 p-4">
            <p className="text-sm text-muted-foreground">Members</p>
            <p className="mt-1 text-sm">Query not available from live API</p>
          </div>
          <div className="rounded-lg border border-border bg-muted/40 p-4">
            <p className="text-sm text-muted-foreground">Billing</p>
            <p className="mt-1 text-sm">Query not available from live API</p>
          </div>
        </div>
        <div className="mt-6 flex flex-wrap gap-3">
          <Link
            href="/dashboard"
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Return to Dashboard
          </Link>
          <Link
            href="/billing/escrow"
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
          >
            View Billing
          </Link>
          <Link
            href="/account/settings"
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
          >
            Account Settings
          </Link>
        </div>
      </div>
    </div>
  );
}
