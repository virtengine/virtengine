import type { Metadata } from 'next';
import { ConsentManager } from '@/components/consent/ConsentManager';
import { ConsentHistory } from '@/components/consent/ConsentHistory';

export const metadata: Metadata = {
  title: 'Privacy Center',
  description: 'Manage GDPR consent, exports, and data deletion preferences',
};

export default function PrivacyCenterPage() {
  return (
    <div className="container py-10">
      <div className="mb-8 space-y-2">
        <h1 className="text-3xl font-semibold">Privacy Center</h1>
        <p className="text-sm text-muted-foreground">
          Review consent history, export your data, or request deletion under GDPR.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
        <ConsentManager />
        <ConsentHistory />
      </div>
    </div>
  );
}
