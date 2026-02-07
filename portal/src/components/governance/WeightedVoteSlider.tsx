/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Input } from '@/components/ui/Input';
import { Progress } from '@/components/ui/Progress';

interface WeightedVoteSliderProps {
  label: string;
  value: number;
  onChange: (value: number) => void;
  colorClass: string;
}

export function WeightedVoteSlider({
  label,
  value,
  onChange,
  colorClass,
}: WeightedVoteSliderProps) {
  const clampValue = (nextValue: number) => {
    if (Number.isNaN(nextValue)) return 0;
    return Math.min(100, Math.max(0, Math.round(nextValue)));
  };

  return (
    <div className="space-y-2 rounded-lg border border-border bg-background p-3">
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="text-xs text-muted-foreground">{value}%</span>
      </div>
      <input
        type="range"
        min={0}
        max={100}
        step={1}
        value={value}
        onChange={(event) => onChange(clampValue(Number(event.target.value)))}
        className="w-full accent-current"
        aria-label={`${label} weight`}
      />
      <Progress value={value} className="h-2" indicatorClassName={colorClass} />
      <Input
        type="number"
        min={0}
        max={100}
        value={value}
        onChange={(event) => onChange(clampValue(Number(event.target.value)))}
      />
    </div>
  );
}
