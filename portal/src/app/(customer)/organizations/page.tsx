import type { Metadata } from 'next';
import { OrganizationUnavailable } from './OrganizationUnavailable';

export const metadata: Metadata = {
  title: 'Organizations',
  description: 'Organization support availability',
};

export default function OrganizationsPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Organizations</h1>
      </div>
      <OrganizationUnavailable />
    </div>
  );
}
