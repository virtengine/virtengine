import dynamic from 'next/dynamic';

const OrderCreateClient = dynamic(() => import('./OrderCreateClient'), {
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

export default function OrderCreatePage() {
  return <OrderCreateClient />;
}
