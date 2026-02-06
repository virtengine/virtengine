import type { Metadata } from 'next';
import ProviderDashboardClient from './ProviderDashboardClient';

export const metadata: Metadata = {
  title: 'Provider Dashboard',
  description: 'Manage your provider infrastructure and offerings',
};

export default function ProviderDashboardPage() {
  return <ProviderDashboardClient />;
}
