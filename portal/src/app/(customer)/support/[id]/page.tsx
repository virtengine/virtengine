import dynamic from 'next/dynamic';

const SupportTicketDetailClient = dynamic(() => import('./SupportTicketDetailClient'), {
  ssr: false,
  loading: () => (
    <div className="container py-10">
      <p>Loading...</p>
    </div>
  ),
});

export function generateStaticParams() {
  return [{ id: '_' }];
}

export default function SupportTicketDetailPage() {
  return <SupportTicketDetailClient />;
}
