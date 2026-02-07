import dynamic from 'next/dynamic';

const JobDetailClient = dynamic(() => import('./JobDetailClient'), {
  ssr: false,
  loading: () => (
    <div className="container py-8">
      <p className="text-sm text-muted-foreground">Loading job details...</p>
    </div>
  ),
});

export function generateStaticParams() {
  return [{ jobId: '_' }];
}

export default function JobDetailPage() {
  return <JobDetailClient />;
}
