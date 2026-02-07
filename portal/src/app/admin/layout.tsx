import type { ReactNode } from 'react';
import { AdminLayout } from '@/layouts';

export default function AdminRoutesLayout({ children }: { children: ReactNode }) {
  return <AdminLayout>{children}</AdminLayout>;
}
