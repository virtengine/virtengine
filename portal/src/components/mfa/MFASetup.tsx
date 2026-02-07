'use client';

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { TOTPSetup } from './TOTPSetup';
import { WebAuthnSetup } from './WebAuthnSetup';
import { SMSSetup } from './SMSSetup';
import { EmailSetup } from './EmailSetup';
import { BackupCodes } from './BackupCodes';

interface MFASetupProps {
  className?: string;
  onComplete?: () => void;
  onCancel?: () => void;
}

type FactorChoice = 'none' | 'totp' | 'webauthn' | 'sms' | 'email' | 'backup';

/**
 * MFA Setup Component
 * Guides users through MFA enrollment with factor type selection.
 */
export function MFASetup({ className, onComplete, onCancel }: MFASetupProps) {
  const [choice, setChoice] = useState<FactorChoice>('none');

  const handleComplete = () => {
    setChoice('none');
    onComplete?.();
  };

  const handleCancel = () => {
    setChoice('none');
    onCancel?.();
  };

  if (choice === 'totp') {
    return <TOTPSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />;
  }
  if (choice === 'webauthn') {
    return (
      <WebAuthnSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />
    );
  }
  if (choice === 'sms') {
    return <SMSSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />;
  }
  if (choice === 'email') {
    return <EmailSetup onComplete={handleComplete} onCancel={handleCancel} className={className} />;
  }
  if (choice === 'backup') {
    return (
      <BackupCodes onComplete={handleComplete} onCancel={handleCancel} className={className} />
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Set Up Two-Factor Authentication</CardTitle>
        <CardDescription>Add an extra layer of security to your account</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <FactorOption
          icon="📱"
          title="Authenticator App"
          description="Use Google Authenticator, Authy, or similar"
          badge="Recommended"
          onClick={() => setChoice('totp')}
        />
        <FactorOption
          icon="🔐"
          title="Security Key"
          description="FIDO2/WebAuthn hardware key or biometrics"
          badge="Strongest"
          onClick={() => setChoice('webauthn')}
        />
        <FactorOption
          icon="💬"
          title="SMS Verification"
          description="Receive a one-time code via text message"
          badge="Backup"
          onClick={() => setChoice('sms')}
        />
        <FactorOption
          icon="📧"
          title="Email Verification"
          description="Receive one-time codes at your email address"
          badge="Backup"
          onClick={() => setChoice('email')}
        />
        <FactorOption
          icon="📋"
          title="Backup Codes"
          description="Generate one-time recovery codes"
          badge="Recovery"
          onClick={() => setChoice('backup')}
        />
      </CardContent>
    </Card>
  );
}

function FactorOption({
  icon,
  title,
  description,
  badge,
  onClick,
}: {
  icon: string;
  title: string;
  description: string;
  badge: string;
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
            {badge}
          </Badge>
        </div>
        <p className="mt-0.5 text-sm text-muted-foreground">{description}</p>
      </div>
    </button>
  );
}
