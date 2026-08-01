'use client';

import { cn } from '@/lib/utils';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/Alert';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';

interface LivenessChallengeProps {
  onCancel?: () => void;
  className?: string;
}

export function LivenessChallenge({ onCancel, className }: LivenessChallengeProps) {
  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Liveness Check Unavailable</CardTitle>
        <CardDescription>Liveness must be verified during secure selfie capture.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Alert variant="destructive">
          <AlertTitle>Provider required</AlertTitle>
          <AlertDescription>
            This standalone challenge cannot produce verified liveness evidence.
          </AlertDescription>
        </Alert>
        {onCancel && (
          <Button variant="outline" onClick={onCancel}>
            Go Back
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
