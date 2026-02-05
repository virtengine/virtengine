import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Identity Verification',
  description: 'Complete your identity verification with VEID',
};

export default function IdentityPage() {
  return (
    <div className="container py-8">
      <div className="mx-auto max-w-3xl">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold">Identity Verification</h1>
          <p className="mt-2 text-muted-foreground">
            Complete your VEID verification to access premium features
          </p>
        </div>

        {/* Identity Score Card */}
        <div className="mb-8 rounded-xl border border-border bg-card p-8">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">Your Identity Score</h2>
              <p className="text-sm text-muted-foreground">Based on completed verifications</p>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-right">
                <div className="text-4xl font-bold text-primary">72</div>
                <div className="text-sm text-muted-foreground">/ 100</div>
              </div>
              <div className="relative h-20 w-20 rounded-full border-4 border-primary/20">
                <svg className="h-full w-full -rotate-90" viewBox="0 0 100 100">
                  <circle
                    cx="50"
                    cy="50"
                    r="45"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="10"
                    className="text-primary"
                    strokeDasharray={`${72 * 2.83} ${283 - 72 * 2.83}`}
                    strokeLinecap="round"
                  />
                </svg>
              </div>
            </div>
          </div>
        </div>

        {/* Verification Steps */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Verification Steps</h2>

          <VerificationStep
            title="Email Verification"
            description="Verify your email address"
            status="completed"
            points={10}
          />
          <VerificationStep
            title="Phone Verification"
            description="Verify your phone number"
            status="completed"
            points={15}
          />
          <VerificationStep
            title="Document Verification"
            description="Upload a government-issued ID"
            status="in-progress"
            points={25}
          />
          <VerificationStep
            title="Selfie Verification"
            description="Take a selfie for liveness check"
            status="pending"
            points={20}
          />
          <VerificationStep
            title="Address Verification"
            description="Verify your address with a utility bill"
            status="pending"
            points={15}
          />
          <VerificationStep
            title="MFA Setup"
            description="Enable multi-factor authentication"
            status="completed"
            points={15}
          />
        </div>

        {/* Required Scopes Info */}
        <div className="mt-8 rounded-lg border border-border bg-muted/50 p-6">
          <h3 className="font-semibold">Unlock More Features</h3>
          <p className="mt-2 text-sm text-muted-foreground">
            Increase your identity score to unlock additional features:
          </p>
          <ul className="mt-4 space-y-2 text-sm">
            <li className="flex items-center gap-2">
              <span className="text-success">✓</span>
              <span>Basic Marketplace Access (50+ score)</span>
            </li>
            <li className="flex items-center gap-2">
              <span className="text-success">✓</span>
              <span>Order Creation (60+ score)</span>
            </li>
            <li className="flex items-center gap-2">
              <span className="text-muted-foreground">○</span>
              <span className="text-muted-foreground">HPC Job Submission (75+ score)</span>
            </li>
            <li className="flex items-center gap-2">
              <span className="text-muted-foreground">○</span>
              <span className="text-muted-foreground">Provider Registration (85+ score)</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}

interface VerificationStepProps {
  title: string;
  description: string;
  status: 'completed' | 'in-progress' | 'pending';
  points: number;
}

function VerificationStep({ title, description, status, points }: VerificationStepProps) {
  const statusConfig = {
    completed: {
      bg: 'bg-success/10',
      text: 'text-success',
      label: 'Completed',
      icon: '✓',
    },
    'in-progress': {
      bg: 'bg-warning/10',
      text: 'text-warning',
      label: 'In Progress',
      icon: '◐',
    },
    pending: {
      bg: 'bg-muted',
      text: 'text-muted-foreground',
      label: 'Pending',
      icon: '○',
    },
  };

  const config = statusConfig[status];

  return (
    <div className="flex items-center justify-between rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-4">
        <div className={`flex h-10 w-10 items-center justify-center rounded-full ${config.bg}`}>
          <span className={config.text}>{config.icon}</span>
        </div>
        <div>
          <h3 className="font-medium">{title}</h3>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <span className="text-sm text-muted-foreground">+{points} pts</span>
        {status === 'pending' && (
          <button
            type="button"
            className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Start
          </button>
        )}
        {status === 'in-progress' && (
          <button
            type="button"
            className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Continue
          </button>
        )}
        {status === 'completed' && <span className={`text-sm ${config.text}`}>{config.label}</span>}
      </div>
    </div>
  );
}
