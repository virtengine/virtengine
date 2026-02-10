// Server component wrapper for static export compatibility
import dynamic from 'next/dynamic';

const OfferingDetailClient = dynamic(() => import('./OfferingDetailClient'), {
  ssr: false,
  loading: () => (
    <div className="container py-8">
      <p>Loading...</p>
    </div>
  ),
});

export function generateStaticParams() {
  return [{ provider: '_', sequence: '_' }];
}

export default function OfferingDetailPage() {
  return <OfferingDetailClient />;
}
