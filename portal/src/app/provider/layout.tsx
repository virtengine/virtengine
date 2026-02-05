import type { ReactNode } from 'react';
import { ProviderLayout } from '@/layouts';

export default function ProviderRoutesLayout({ children }: { children: ReactNode }) {
  return <ProviderLayout>{children}</ProviderLayout>;
}
