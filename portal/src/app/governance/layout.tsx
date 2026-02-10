import type { ReactNode } from 'react';
import { CustomerLayout } from '@/layouts';

export default function GovernanceRoutesLayout({ children }: { children: ReactNode }) {
  return <CustomerLayout>{children}</CustomerLayout>;
}
