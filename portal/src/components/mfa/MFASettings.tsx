'use client';

import { useState } from 'react';
import { useMFAStore } from '@/features/mfa';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import { FactorList } from './FactorList';
import { TOTPSetup } from './TOTPSetup';
import { WebAuthnSetup } from './WebAuthnSetup';
import { SMSSetup } from './SMSSetup';
import { EmailSetup } from './EmailSetup';
import { BackupCodes } from './BackupCodes';
import { RecoveryFlow } from './RecoveryFlow';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Separator } from '@/components/ui/Separator';

interface MFASettingsProps {
  className?: string;
}

type SetupView = 'none' | 'totp' | 'webauthn' | 'sms' | 'email' | 'backup' | 'recovery';

/**
 * MFA Settings Component
 * Full management page for MFA configuration, enrolled factors, trusted browsers, and audit.
 */
export function MFASettings({ className }: MFASettingsProps) {
  const {
    isLoading,
    isEnabled,
    factors,
    trustedBrowsers,
    auditLog = [],
    revokeTrustedBrowser,
  } = useMFAStore();
  const [setupView, setSetupView] = useState<SetupView>('none');
  const safeFactors = factors ?? [];
  const safeTrustedBrowsers = trustedBrowsers ?? [];
  const safeAuditLog = auditLog ?? [];

  if (isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-64 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  // If user is in a setup flow, show the setup component full-width
  if (setupView !== 'none') {
    const handleComplete = () => setSetupView('none');
    const handleCancel = () => setSetupView('none');

    switch (setupView) {
      case 'totp':
        return (
          <TOTPSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />
        );
      case 'webauthn':
        return (
          <WebAuthnSetup
            onComplete={handleComplete}
            onCancel={handleCancel}
            className={className}
          />
        );
      case 'sms':
        return (
          <SMSSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />
        );
      case 'email':
        return (
          <EmailSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />
        );
      case 'backup':
        return (
          <BackupCodes onComplete={handleComplete} onCancel={handleCancel} className={className} />
        );
      case 'recovery':
        return (
          <RecoveryFlow onComplete={handleComplete} onCancel={handleCancel} className={className} />
        );
    }
  }

  return (
    <div className={cn('space-y-6', className)}>
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Two-Factor Authentication</CardTitle>
              <CardDescription className="mt-1">
                Manage your MFA factors, trusted devices, and security audit log
              </CardDescription>
            </div>
            <Badge variant={isEnabled ? 'success' : 'warning'} dot>
              {isEnabled ? 'Enabled' : 'Disabled'}
            </Badge>
          </div>
        </CardHeader>
      </Card>

      <Tabs defaultValue="factors">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="factors">Factors ({safeFactors.length})</TabsTrigger>
          <TabsTrigger value="enroll">Add Factor</TabsTrigger>
          <TabsTrigger value="browsers">Trusted ({safeTrustedBrowsers.length})</TabsTrigger>
          <TabsTrigger value="audit">Audit Log</TabsTrigger>
        </TabsList>

        {/* Enrolled factors */}
        <TabsContent value="factors" className="mt-4">
          <FactorList onAddFactor={() => {}} />
        </TabsContent>

        {/* Enrollment options */}
        <TabsContent value="enroll" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Add a Security Factor</CardTitle>
              <CardDescription>
                Choose a method to add to your account. We recommend using an authenticator app or
                security key for the strongest protection.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <FactorOption
                icon="📱"
                title="Authenticator App (TOTP)"
                description="Use Google Authenticator, Authy, or similar app to generate time-based codes"
                security="Recommended"
                onClick={() => setSetupView('totp')}
              />
              <FactorOption
                icon="🔐"
                title="Security Key (FIDO2/WebAuthn)"
                description="Use a hardware security key or device biometrics (Touch ID, Face ID, Windows Hello)"
                security="Strongest"
                onClick={() => setSetupView('webauthn')}
              />
              <FactorOption
                icon="💬"
                title="SMS Verification"
                description="Receive a one-time code via text message"
                security="Backup"
                onClick={() => setSetupView('sms')}
              />
              <FactorOption
                icon="📧"
                title="Email Verification"
                description="Use your email address to receive verification codes"
                security="Backup"
                onClick={() => setSetupView('email')}
              />
              <FactorOption
                icon="📋"
                title="Backup Codes"
                description="Generate one-time recovery codes in case you lose access to your primary factor"
                security="Recovery"
                onClick={() => setSetupView('backup')}
              />

              <Separator className="my-4" />

              <button
                type="button"
                onClick={() => setSetupView('recovery')}
                className="w-full text-left text-sm text-muted-foreground hover:text-foreground"
              >
                🔄 Lost your device? Start account recovery →
              </button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Trusted browsers */}
        <TabsContent value="browsers" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Trusted Browsers</CardTitle>
              <CardDescription>
                Browsers you&apos;ve marked as trusted will skip MFA for a configured period.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {safeTrustedBrowsers.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No trusted browsers. When you verify with MFA and choose &quot;Trust this
                  browser&quot;, it will appear here.
                </p>
              ) : (
                <div className="space-y-3">
                  {safeTrustedBrowsers.map((browser) => (
                    <div
                      key={browser.id}
                      className="flex items-center justify-between rounded-lg border bg-card p-3"
                    >
                      <div>
                        <p className="text-sm font-medium">
                          {browser.deviceName || browser.browserName}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          Trusted {new Date(browser.trustedAt).toLocaleDateString()} · Expires{' '}
                          {new Date(browser.expiresAt).toLocaleDateString()}
                          {browser.region && ` · ${browser.region}`}
                        </p>
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-destructive"
                        onClick={() => revokeTrustedBrowser(browser.id)}
                      >
                        Revoke
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Audit log */}
        <TabsContent value="audit" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Security Audit Log</CardTitle>
              <CardDescription>Recent MFA verification activity on your account.</CardDescription>
            </CardHeader>
            <CardContent>
              {safeAuditLog.length === 0 ? (
                <p className="text-sm text-muted-foreground">No MFA activity recorded yet.</p>
              ) : (
                <div className="space-y-2">
                  {safeAuditLog.map((entry) => (
                    <div
                      key={entry.id}
                      className="flex items-center justify-between border-b py-2 last:border-0"
                    >
                      <div>
                        <p className="text-sm">
                          <Badge
                            variant={entry.success ? 'success' : 'destructive'}
                            size="sm"
                            className="mr-2"
                          >
                            {entry.success ? 'Success' : 'Failed'}
                          </Badge>
                          <span className="text-muted-foreground">
                            {entry.transactionType.replace(/_/g, ' ')}
                          </span>
                        </p>
                        <p className="text-xs text-muted-foreground">
                          via {entry.factorType}
                          {entry.region && ` · ${entry.region}`}
                          {entry.txHash && ` · tx: ${entry.txHash.slice(0, 12)}…`}
                        </p>
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {new Date(entry.timestamp).toLocaleString()}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

/** Factor option button for enrollment selection */
function FactorOption({
  icon,
  title,
  description,
  security,
  onClick,
}: {
  icon: string;
  title: string;
  description: string;
  security: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-4 rounded-lg border bg-card p-4 text-left transition-colors hover:border-primary/50 hover:bg-accent"
    >
      <span className="text-2xl" aria-hidden="true">
        {icon}
      </span>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="font-medium">{title}</span>
          <Badge variant="outline" size="sm">
            {security}
          </Badge>
        </div>
        <p className="mt-0.5 text-sm text-muted-foreground">{description}</p>
      </div>
      <span className="text-muted-foreground" aria-hidden="true">
        →
      </span>
    </button>
  );
}
