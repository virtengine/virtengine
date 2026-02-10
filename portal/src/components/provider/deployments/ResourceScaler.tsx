'use client';

interface ResourceScalerProps {
  label: string;
  value: number;
  unit: string;
  min: number;
  max: number;
  step?: number;
  onChange: (value: number) => void;
}

export function ResourceScaler({
  label,
  value,
  unit,
  min,
  max,
  step = 1,
  onChange,
}: ResourceScalerProps) {
  const handleChange = (next: number) => {
    const clamped = Math.min(Math.max(next, min), max);
    onChange(clamped);
  };

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium">{label}</p>
          <p className="text-xs text-muted-foreground">
            {min}-{max} {unit}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => handleChange(value - step)}
            className="flex h-8 w-8 items-center justify-center rounded-full border border-border text-sm hover:bg-accent"
          >
            -
          </button>
          <div className="min-w-[72px] text-center text-sm font-semibold">
            {value} {unit}
          </div>
          <button
            type="button"
            onClick={() => handleChange(value + step)}
            className="flex h-8 w-8 items-center justify-center rounded-full border border-border text-sm hover:bg-accent"
          >
            +
          </button>
        </div>
      </div>
      <input
        type="range"
        min={min}
        max={max}
        value={value}
        step={step}
        onChange={(event) => handleChange(Number(event.target.value))}
        className="mt-4 w-full"
      />
    </div>
  );
}
