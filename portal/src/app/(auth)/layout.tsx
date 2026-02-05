import type { ReactNode } from 'react';
import { AuthLayout } from '@/layouts';

export default function AuthRoutesLayout({ children }: { children: ReactNode }) {
  return <AuthLayout>{children}</AuthLayout>;
}
