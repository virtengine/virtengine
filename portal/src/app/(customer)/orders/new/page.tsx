import dynamic from 'next/dynamic';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Create Order',
  description: 'Configure and deploy compute resources',
};

const NewOrderClient = dynamic(() => import('./NewOrderClient'), {
  ssr: false,
  loading: () => (
    <div className="container py-8">
      <div className="mx-auto max-w-3xl animate-pulse space-y-4">
        <div className="h-8 w-48 rounded bg-muted" />
        <div className="h-4 w-64 rounded bg-muted" />
        <div className="mt-8 h-12 rounded bg-muted" />
        <div className="h-64 rounded bg-muted" />
      </div>
    </div>
  ),
});

export default function NewOrderPage() {
  return <NewOrderClient />;
}
