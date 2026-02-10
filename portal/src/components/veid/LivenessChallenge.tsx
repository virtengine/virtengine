/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Liveness Challenge Component
 * Displays liveness challenge instructions and tracks completion.
 */

'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import { Badge } from '@/components/ui/Badge';
import { Progress } from '@/components/ui/Progress';

type ChallengeType = 'blink' | 'turn-left' | 'turn-right' | 'smile' | 'nod';

interface Challenge {
  type: ChallengeType;
  instruction: string;
  icon: string;
}

const CHALLENGES: Challenge[] = [
  { type: 'blink', instruction: 'Blink your eyes slowly', icon: '👁' },
  { type: 'turn-left', instruction: 'Turn your head slightly to the left', icon: '👈' },
  { type: 'turn-right', instruction: 'Turn your head slightly to the right', icon: '👉' },
  { type: 'smile', instruction: 'Smile naturally', icon: '😊' },
  { type: 'nod', instruction: 'Nod your head gently', icon: '👍' },
];

interface LivenessChallengeProps {
  /** Number of challenges to present (1-5) */
  challengeCount?: number;
  /** Time limit per challenge in seconds */
  timeLimitSeconds?: number;
  /** Callback when all challenges are completed */
  onComplete: () => void;
  /** Callback on failure */
  onFail?: () => void;
  /** Callback to cancel */
  onCancel?: () => void;
  className?: string;
}

export function LivenessChallenge({
  challengeCount = 3,
  timeLimitSeconds = 10,
  onComplete,
  onFail,
  onCancel,
  className,
}: LivenessChallengeProps) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const [timeRemaining, setTimeRemaining] = useState(timeLimitSeconds);
  const [selectedChallenges] = useState<Challenge[]>(() => {
    const shuffled = [...CHALLENGES].sort(() => Math.random() - 0.5);
    return shuffled.slice(0, Math.min(challengeCount, CHALLENGES.length));
  });
  const [isActive, setIsActive] = useState(false);
  const [failed, setFailed] = useState(false);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const totalChallenges = selectedChallenges.length;
  const currentChallenge = selectedChallenges[currentIndex];
  const progress = Math.round((currentIndex / totalChallenges) * 100);

  useEffect(() => {
    if (!isActive || !currentChallenge) return;

    timerRef.current = setInterval(() => {
      setTimeRemaining((prev) => {
        if (prev <= 1) {
          // Time's up - simulate auto-detection pass for demo
          // In production, this would use ML liveness detection from capture lib
          handleChallengePass();
          return timeLimitSeconds;
        }
        return prev - 1;
      });
    }, 1000);

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isActive, currentIndex]);

  const handleStart = useCallback(() => {
    setIsActive(true);
    setTimeRemaining(timeLimitSeconds);
  }, [timeLimitSeconds]);

  const handleChallengePass = useCallback(() => {
    if (timerRef.current) clearInterval(timerRef.current);

    if (currentIndex + 1 >= totalChallenges) {
      setIsActive(false);
      onComplete();
    } else {
      setCurrentIndex((prev) => prev + 1);
      setTimeRemaining(timeLimitSeconds);
    }
  }, [currentIndex, totalChallenges, timeLimitSeconds, onComplete]);

  const handleSkipChallenge = useCallback(() => {
    // Simulate passing (in production, the ML model would detect the action)
    handleChallengePass();
  }, [handleChallengePass]);

  const handleRetry = useCallback(() => {
    setFailed(false);
    setCurrentIndex(0);
    setTimeRemaining(timeLimitSeconds);
    setIsActive(false);
  }, [timeLimitSeconds]);

  if (failed) {
    return (
      <Card className={cn(className)}>
        <CardHeader>
          <CardTitle className="text-destructive">Liveness Check Failed</CardTitle>
          <CardDescription>We could not verify your liveness. Please try again.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Alert variant="destructive">
            <AlertDescription>
              The liveness check did not complete successfully. This may happen due to poor lighting
              or camera issues.
            </AlertDescription>
          </Alert>
          <div className="flex gap-2">
            <Button onClick={handleRetry}>Try Again</Button>
            {onFail && (
              <Button variant="outline" onClick={onFail}>
                Skip
              </Button>
            )}
            {onCancel && (
              <Button variant="outline" onClick={onCancel}>
                Cancel
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Liveness Check</CardTitle>
            <CardDescription>
              Follow the instructions to verify you are a real person
            </CardDescription>
          </div>
          <Badge variant="secondary">
            {currentIndex + 1} / {totalChallenges}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        <Progress value={progress} className="h-2" />

        {!isActive ? (
          <div className="space-y-6 text-center">
            <div className="mx-auto flex h-24 w-24 items-center justify-center rounded-full bg-primary/10">
              <span className="text-4xl" role="img" aria-label="Liveness check">
                🎯
              </span>
            </div>
            <div>
              <h3 className="text-lg font-semibold">Ready for Liveness Check?</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                You will be asked to perform {totalChallenges} simple actions. Make sure your face
                is clearly visible in the camera.
              </p>
            </div>
            <div className="flex justify-center gap-2">
              <Button size="lg" onClick={handleStart}>
                Begin Check
              </Button>
              {onCancel && (
                <Button variant="outline" size="lg" onClick={onCancel}>
                  Cancel
                </Button>
              )}
            </div>
          </div>
        ) : currentChallenge ? (
          <div className="space-y-6 text-center">
            <div className="mx-auto flex h-32 w-32 items-center justify-center rounded-full border-4 border-primary bg-primary/5">
              <span className="text-5xl" role="img" aria-label={currentChallenge.instruction}>
                {currentChallenge.icon}
              </span>
            </div>

            <div>
              <h3 className="text-xl font-bold">{currentChallenge.instruction}</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                Time remaining:{' '}
                <span
                  className={cn(
                    'font-mono font-bold',
                    timeRemaining <= 3 ? 'text-destructive' : 'text-foreground'
                  )}
                >
                  {timeRemaining}s
                </span>
              </p>
            </div>

            {/* Timer ring */}
            <div className="mx-auto w-16">
              <svg viewBox="0 0 100 100" className="-rotate-90">
                <circle
                  cx="50"
                  cy="50"
                  r="45"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="6"
                  className="text-muted"
                />
                <circle
                  cx="50"
                  cy="50"
                  r="45"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="6"
                  strokeLinecap="round"
                  className="text-primary transition-all duration-1000"
                  strokeDasharray={`${(timeRemaining / timeLimitSeconds) * 283} 283`}
                />
              </svg>
            </div>

            <Button variant="outline" size="sm" onClick={handleSkipChallenge}>
              I&apos;ve completed this action
            </Button>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
