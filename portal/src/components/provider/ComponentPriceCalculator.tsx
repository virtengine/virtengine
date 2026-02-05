'use client';

import { useMemo, useState } from 'react';

type ResourceKey = 'cpu' | 'ram' | 'storage' | 'gpu';

const resourceLabels: Record<ResourceKey, string> = {
  cpu: 'vCPU',
  ram: 'RAM (GB)',
  storage: 'Storage (GB)',
  gpu: 'GPU',
};

export default function ComponentPriceCalculator() {
  const [units, setUnits] = useState<Record<ResourceKey, number>>({
    cpu: 4,
    ram: 8,
    storage: 100,
    gpu: 0,
  });
  const [prices, setPrices] = useState<Record<ResourceKey, number>>({
    cpu: 5,
    ram: 2,
    storage: 0.1,
    gpu: 10,
  });
  const [denom, setDenom] = useState('uve');

  const totalPerHour = useMemo(() => {
    return (Object.keys(units) as ResourceKey[]).reduce((sum, key) => {
      const unitCount = Number.isFinite(units[key]) ? units[key] : 0;
      const unitPrice = Number.isFinite(prices[key]) ? prices[key] : 0;
      return sum + unitCount * unitPrice;
    }, 0);
  }, [prices, units]);

  const totalPerDay = totalPerHour * 24;

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h2 className="text-lg font-semibold">Component Price Calculator</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        Preview total pricing based on resource units and per-unit rates.
      </p>

      <div className="mt-6 grid gap-4">
        {(Object.keys(resourceLabels) as ResourceKey[]).map((key) => (
          <div key={key} className="grid grid-cols-3 items-center gap-3">
            <span className="text-sm font-medium">{resourceLabels[key]}</span>
            <input
              type="number"
              min={0}
              step={key === 'storage' ? 1 : 0.5}
              value={units[key]}
              onChange={(event) => setUnits({ ...units, [key]: Number(event.target.value) })}
              className="rounded-lg border border-border bg-background px-3 py-2 text-sm"
            />
            <div className="flex items-center gap-2">
              <input
                type="number"
                min={0}
                step={0.01}
                value={prices[key]}
                onChange={(event) => setPrices({ ...prices, [key]: Number(event.target.value) })}
                className="flex-1 rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
              <span className="text-xs text-muted-foreground">/hr</span>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-6 flex flex-wrap items-center justify-between gap-4 rounded-lg border border-border bg-muted/20 p-4">
        <div>
          <div className="text-xs text-muted-foreground">Currency</div>
          <input
            type="text"
            value={denom}
            onChange={(event) => setDenom(event.target.value)}
            className="mt-1 w-24 rounded-lg border border-border bg-background px-3 py-2 text-sm"
          />
        </div>
        <div className="text-right">
          <div className="text-xs text-muted-foreground">Estimated total / hour</div>
          <div className="text-2xl font-semibold">
            {totalPerHour.toFixed(2)} {denom}
          </div>
          <div className="text-sm text-muted-foreground">
            {totalPerDay.toFixed(2)} {denom} / day
          </div>
        </div>
      </div>
    </div>
  );
}
