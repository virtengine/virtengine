'use client';

import Link from 'next/link';
import { useIdentity } from '@/lib/portal-adapter';
import { TierDisplay } from '@/components/veid';
import { SCOPE_DISPLAY, FEATURE_THRESHOLDS } from '@/features/veid';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import { cn } from '@/lib/utils';

export default function IdentityPage() {
  const { state } = useIdentity();

  const currentScore = state.score?.value ?? 0;
  const tier = state.score?.tier ?? 'unverified';

  // Build scope statuses from identity state
  const allScopes: {
    type: keyof typeof SCOPE_DISPLAY;
    status: 'completed' | 'in-progress' | 'pending';
  }[] = [
    {
      type: 'email',
      status: state.completedScopes.some((s) => s.type === 'email' && s.completed)
        ? 'completed'
        : 'pending',
    },
    {
      type: 'id_document',
      status:
        state.status === 'processing' &&
        !state.completedScopes.some((s) => s.type === 'id_document' && s.completed)
          ? 'in-progress'
          : state.completedScopes.some((s) => s.type === 'id_document' && s.completed)
            ? 'completed'
            : 'pending',
    },
    {
      type: 'selfie',
      status: state.completedScopes.some((s) => s.type === 'selfie' && s.completed)
        ? 'completed'
        : 'pending',
    },
    {
      type: 'domain',
      status: state.completedScopes.some((s) => s.type === 'domain' && s.completed)
        ? 'completed'
        : 'pending',
    },
    {
      type: 'sso',
      status: state.completedScopes.some((s) => s.type === 'sso' && s.completed)
        ? 'completed'
        : 'pending',
    },
    {
      type: 'biometric',
      status: state.completedScopes.some((s) => s.type === 'biometric' && s.completed)
        ? 'completed'
        : 'pending',
    },
  ];

  return (
    <div className="container py-8">
      <div className="mx-auto max-w-3xl">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold">Identity Verification</h1>
          <p className="mt-2 text-muted-foreground">
            Complete your VEID verification to access premium features
          </p>
        </div>

        {/* Identity Score Display */}
        <div className="mb-8">
          <Card>
            <CardContent className="flex flex-col items-center gap-6 p-8 sm:flex-row sm:justify-between">
              <div>
                <h2 className="text-lg font-semibold">Your Identity Score</h2>
                <p className="text-sm text-muted-foreground">Based on completed verifications</p>
              </div>
              <TierDisplay score={currentScore} tier={tier} size="lg" />
            </CardContent>
          </Card>
        </div>

        {/* Verification Steps */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Verification Steps</h2>

          {allScopes.map((scope) => {
            const info = SCOPE_DISPLAY[scope.type];
            return (
              <VerificationStep
                key={scope.type}
                title={info.label}
                description={info.description}
                status={scope.status}
                points={info.points}
                icon={info.icon}
              />
            );
          })}
        </div>

        {/* Feature access thresholds */}
        <div className="mt-8 rounded-lg border border-border bg-muted/50 p-6">
          <h3 className="font-semibold">Unlock More Features</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Increase your identity score to unlock additional features:
          </p>
          <ul className="mt-4 space-y-2 text-sm">
            {FEATURE_THRESHOLDS.map((threshold) => {
              const met = currentScore >= threshold.minScore;
              return (
                <li key={threshold.action} className="flex items-center gap-2">
                  <span
                    className={met ? 'text-green-600 dark:text-green-400' : 'text-muted-foreground'}
                  >
                    {met ? '✓' : '○'}
                  </span>
                  <span className={cn(!met && 'text-muted-foreground')}>
                    {threshold.label} ({threshold.minScore}+ score)
                  </span>
                </li>
              );
            })}
          </ul>
        </div>

        {/* CTA */}
        {state.status === 'unknown' && (
          <div className="mt-8 text-center">
            <Link href="/verify">
              <Button size="lg">Start Verification</Button>
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}

interface VerificationStepProps {
  title: string;
  description: string;
  status: 'completed' | 'in-progress' | 'pending';
  points: number;
  icon?: string;
}

function VerificationStep({ title, description, status, points, icon }: VerificationStepProps) {
  const statusConfig = {
    completed: {
      bg: 'bg-green-50 dark:bg-green-950/20',
      text: 'text-green-600 dark:text-green-400',
      label: 'Completed',
      statusIcon: '✓',
    },
    'in-progress': {
      bg: 'bg-amber-50 dark:bg-amber-950/20',
      text: 'text-amber-600 dark:text-amber-400',
      label: 'In Progress',
      statusIcon: '◐',
    },
    pending: {
      bg: 'bg-muted',
      text: 'text-muted-foreground',
      label: 'Pending',
      statusIcon: '○',
    },
  };

  const config = statusConfig[status];

  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-4">
        <div className={cn('flex h-10 w-10 items-center justify-center rounded-full', config.bg)}>
          <span className={config.text}>{icon ?? config.statusIcon}</span>
        </div>
        <div>
          <h3 className="font-medium">{title}</h3>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-muted-foreground">+{points} pts</span>
        {status === 'pending' && (
          <Link href="/verify">
            <Button size="sm">Start</Button>
          </Link>
        )}
        {status === 'in-progress' && (
          <Link href="/verify">
            <Button size="sm">Continue</Button>
          </Link>
        )}
        {status === 'completed' && (
          <span className={cn('text-sm', config.text)}>{config.label}</span>
        )}
      </div>
    </div>
  );
}
