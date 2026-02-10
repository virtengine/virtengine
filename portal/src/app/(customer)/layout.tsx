import type { ReactNode } from 'react';
import { CustomerLayout } from '@/layouts';

export default function CustomerRoutesLayout({ children }: { children: ReactNode }) {
  return <CustomerLayout>{children}</CustomerLayout>;
}
