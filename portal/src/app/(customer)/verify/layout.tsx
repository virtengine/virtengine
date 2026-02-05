import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Identity Verification',
  description: 'Verify your identity to unlock full platform features',
};

export default function VerifyLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
