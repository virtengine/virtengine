import dynamic from 'next/dynamic';

const CompareClient = dynamic(() => import('./CompareClient'), {
  ssr: false,
  loading: () => (
    <div className="container py-8">
      <p>Loading comparison...</p>
    </div>
  ),
});

export default function ComparePage() {
  return <CompareClient />;
}
