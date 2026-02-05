// Layout for nested dynamic route
export async function generateStaticParams() {
  return [{ provider: '_' }];
}

export default function ProviderLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
