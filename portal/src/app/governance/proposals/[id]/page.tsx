// Server component wrapper for static export compatibility
import dynamic from 'next/dynamic';

const ProposalDetailClient = dynamic(() => import('./ProposalDetailClient'), {
  ssr: false,
  loading: () => (
    <div className="container py-8">
      <p>Loading...</p>
    </div>
  ),
});

export function generateStaticParams() {
  return [{ id: '_' }];
}

export default function ProposalDetailPage() {
  return <ProposalDetailClient />;
}
