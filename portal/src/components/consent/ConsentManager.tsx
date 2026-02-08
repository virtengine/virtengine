'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { useWallet } from '@/lib/portal-adapter';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import type { ConsentPurpose, ConsentSettingsResponse } from '@/types/consent';
import { DataExportStatus } from '@/components/consent/DataExportStatus';

const CONSENT_PURPOSES: Array<{
  key: ConsentPurpose;
  scopeId: string;
  title: string;
  description: string;
  required: boolean;
}> = [
  {
    key: 'biometric_processing',
    scopeId: 'veid.biometric',
    title: 'Biometric processing',
    description: 'Allow facial, liveness, and biometric processing to verify your VEID identity.',
    required: true,
  },
  {
    key: 'data_retention',
    scopeId: 'veid.data_retention',
    title: 'Data retention',
    description: 'Retain verification artifacts for the legally required period.',
    required: true,
  },
  {
    key: 'analytics',
    scopeId: 'veid.analytics',
    title: 'Product analytics',
    description: 'Share anonymized usage insights to improve verification accuracy.',
    required: false,
  },
  {
    key: 'marketing',
    scopeId: 'veid.marketing',
    title: 'Marketing updates',
    description: 'Receive updates about new features, integrations, and compliance releases.',
    required: false,
  },
];

function buildConsentText(purpose: ConsentPurpose) {
  const summary = CONSENT_PURPOSES.find((item) => item.key === purpose);
  return `Consent for ${summary?.title ?? purpose}. I agree to the VirtEngine privacy policy.`;
}

export function ConsentManager() {
  const wallet = useWallet();
  const account = wallet.accounts[wallet.activeAccountIndex];
  const address = account?.address ?? 'virtengine1demo';

  const [settings, setSettings] = useState<ConsentSettingsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadSettings = useCallback(async () => {
    try {
      setLoading(true);
      const res = await fetch(`/api/consent/${address}`);
      const data = (await res.json()) as ConsentSettingsResponse;
      setSettings(data);
      setError(null);
    } catch (err) {
      setError('Unable to load consent settings.');
    } finally {
      setLoading(false);
    }
  }, [address]);

  useEffect(() => {
    void loadSettings();
  }, [loadSettings]);

  const activeConsent = useCallback(
    (scopeId: string) => settings?.consents.find((consent) => consent.scopeId === scopeId),
    [settings?.consents]
  );

  const handleToggle = async (entry: (typeof CONSENT_PURPOSES)[number]) => {
    if (!settings) return;
    const current = activeConsent(entry.scopeId);
    const shouldEnable = !current || current.status !== 'active';

    setUpdating(entry.scopeId);
    try {
      if (shouldEnable) {
        await fetch('/api/consent/grant', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            dataSubject: address,
            scopeId: entry.scopeId,
            purpose: entry.key,
            consentText: buildConsentText(entry.key),
            signature: `sig-${Date.now()}`,
          }),
        });
      } else if (current) {
        await fetch('/api/consent/withdraw', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            dataSubject: address,
            consentId: current.id,
          }),
        });
      }
      await loadSettings();
    } catch (err) {
      setError('Unable to update consent right now.');
    } finally {
      setUpdating(null);
    }
  };

  const withdrawnRequired = useMemo(() => {
    return settings?.consents.some(
      (consent) =>
        consent.status === 'withdrawn' &&
        ['biometric_processing', 'data_retention'].includes(consent.purpose)
    );
  }, [settings?.consents]);

  if (loading) {
    return <p className="text-sm text-muted-foreground">Loading privacy controls…</p>;
  }

  if (!settings) {
    return <p className="text-sm text-muted-foreground">No privacy data available.</p>;
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Consent controls</CardTitle>
          <CardDescription>
            Manage how VirtEngine processes your identity data across the VEID network.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="grid gap-3 sm:grid-cols-3">
            <div className="rounded-lg border border-border bg-muted/30 p-3">
              <p className="text-xs text-muted-foreground">Policy version</p>
              <p className="text-sm font-semibold">v1.0</p>
            </div>
            <div className="rounded-lg border border-border bg-muted/30 p-3">
              <p className="text-xs text-muted-foreground">Consent version</p>
              <p className="text-sm font-semibold">#{settings.consentVersion}</p>
            </div>
            <div className="rounded-lg border border-border bg-muted/30 p-3">
              <p className="text-xs text-muted-foreground">Last updated</p>
              <p className="text-sm font-semibold">
                {new Date(settings.lastUpdatedAt).toLocaleDateString()}
              </p>
            </div>
          </div>

          <div className="space-y-3">
            {CONSENT_PURPOSES.map((entry) => {
              const current = activeConsent(entry.scopeId);
              const isActive = current?.status === 'active';
              const isDisabled = updating === entry.scopeId;
              const statusLabel = current?.status ?? 'inactive';

              return (
                <div
                  key={entry.scopeId}
                  className="flex flex-col gap-4 rounded-lg border border-border bg-muted/20 p-4 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-semibold">{entry.title}</p>
                      {entry.required && (
                        <Badge variant="secondary" className="uppercase tracking-wide">
                          Required
                        </Badge>
                      )}
                      <Badge
                        variant={isActive ? 'default' : 'outline'}
                        className={isActive ? 'bg-success/15 text-success' : undefined}
                      >
                        {statusLabel}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted-foreground">{entry.description}</p>
                  </div>
                  <Button
                    variant={isActive ? 'secondary' : 'default'}
                    className="min-w-[120px]"
                    onClick={() => handleToggle(entry)}
                    aria-pressed={isActive}
                    disabled={isDisabled}
                  >
                    {isActive ? 'Withdraw' : 'Grant'}
                  </Button>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {withdrawnRequired && (
        <Alert variant="warning">
          <AlertTitle>Required consent withdrawn</AlertTitle>
          <AlertDescription>
            Some required consents are currently withdrawn. This may pause VEID verification and
            restrict marketplace access until re-enabled.
          </AlertDescription>
        </Alert>
      )}

      {error && (
        <Alert variant="destructive">
          <AlertTitle>Update failed</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <DataExportStatus dataSubject={address} />
    </div>
  );
}
