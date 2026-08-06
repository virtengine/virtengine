import type { Metadata } from 'next';
import Link from 'next/link';
import { OrganizationUnavailable } from '../OrganizationUnavailable';

export const metadata: Metadata = {
  title: 'Organization Details',
  description: 'Organization support availability',
};

export default async function OrganizationDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await params;

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link
          href="/organizations"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to organizations
        </Link>
        <h1 className="text-2xl font-bold">Organization</h1>
      </div>
      <OrganizationUnavailable />
    </div>
  );
}
