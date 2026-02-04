'use client';

import { cn } from '@/lib/utils';
import { Progress } from '@/components/ui/progress';
import { CheckCircle2, Circle, Loader2 } from 'lucide-react';

type CaptureStep = 'document-front' | 'document-back' | 'selfie' | 'processing' | 'complete';

interface CaptureProgressProps {
  currentStep: CaptureStep;
  className?: string;
}

const steps: { key: CaptureStep; label: string }[] = [
  { key: 'document-front', label: 'Front of ID' },
  { key: 'document-back', label: 'Back of ID' },
  { key: 'selfie', label: 'Selfie' },
  { key: 'processing', label: 'Processing' },
  { key: 'complete', label: 'Complete' },
];

/**
 * Capture Progress Component
 * Shows progress through the capture flow
 */
export function CaptureProgress({ currentStep, className }: CaptureProgressProps) {
  const currentIndex = steps.findIndex((s) => s.key === currentStep);
  const progressPercent = ((currentIndex + 1) / steps.length) * 100;

  return (
    <div className={cn('space-y-4', className)}>
      <Progress value={progressPercent} className="h-2" />
      
      <div className="flex justify-between">
        {steps.map((step, index) => {
          const isComplete = index < currentIndex;
          const isCurrent = index === currentIndex;
          const isPending = index > currentIndex;

          return (
            <div
              key={step.key}
              className={cn(
                'flex flex-col items-center gap-1',
                isComplete && 'text-green-600 dark:text-green-400',
                isCurrent && 'text-primary',
                isPending && 'text-muted-foreground'
              )}
            >
              {isComplete ? (
                <CheckCircle2 className="h-5 w-5" />
              ) : isCurrent ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : (
                <Circle className="h-5 w-5" />
              )}
              <span className="text-xs font-medium">{step.label}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
