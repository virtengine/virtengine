/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Time-series chart using native SVG. No external charting library required.
 * Displays metric history with configurable time ranges.
 */

'use client';

import { useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import type { MetricSeries, Granularity } from '@virtengine/portal/types/metrics';
import { formatTimestamp } from '@virtengine/portal/types/metrics';

interface TimeSeriesChartProps {
  title: string;
  series: MetricSeries | undefined;
  granularity: Granularity;
  height?: number;
  isLoading?: boolean;
  className?: string;
}

export function TimeSeriesChart({
  title,
  series,
  granularity,
  height = 200,
  isLoading,
  className,
}: TimeSeriesChartProps) {
  const chartData = useMemo(() => {
    if (!series?.data.length) return null;

    const data = series.data;
    const minVal = Math.min(...data.map((p) => p.value));
    const maxVal = Math.max(...data.map((p) => p.value));
    const range = maxVal - minVal || 1;
    const padding = 16;
    const chartWidth = 100;
    const chartHeight = height - padding * 2;

    const points = data.map((p, i) => ({
      x: padding + (i / (data.length - 1 || 1)) * (chartWidth - padding * 2),
      y: padding + (1 - (p.value - minVal) / range) * chartHeight,
      value: p.value,
      timestamp: p.timestamp,
    }));

    const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ');

    const lastPoint = points[points.length - 1];
    const firstPoint = points[0];
    if (!lastPoint || !firstPoint) return null;

    const areaPath =
      linePath +
      ` L ${lastPoint.x} ${padding + chartHeight}` +
      ` L ${firstPoint.x} ${padding + chartHeight} Z`;

    // Pick ~5 labels evenly distributed
    const labelCount = Math.min(5, data.length);
    const labels = Array.from({ length: labelCount }, (_, i) => {
      const idx = Math.round((i / (labelCount - 1 || 1)) * (data.length - 1));
      const pt = points[idx];
      const dp = data[idx];
      return {
        x: pt?.x ?? 0,
        text: dp ? formatTimestamp(dp.timestamp, granularity) : '',
      };
    });

    return { points, linePath, areaPath, minVal, maxVal, labels };
  }, [series, height, granularity]);

  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <Skeleton className="w-full" style={{ height }} />
        </CardContent>
      </Card>
    );
  }

  if (!chartData || !series) {
    return (
      <Card className={className}>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium">{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <div
            className="flex items-center justify-center text-sm text-muted-foreground"
            style={{ height }}
          >
            No data available
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <span className="text-xs text-muted-foreground">{series.unit}</span>
      </CardHeader>
      <CardContent>
        <svg
          viewBox={`0 0 100 ${height}`}
          preserveAspectRatio="none"
          className="w-full"
          style={{ height }}
          role="img"
          aria-label={`${title} chart`}
        >
          {/* Grid lines */}
          {[0.25, 0.5, 0.75].map((frac) => (
            <line
              key={frac}
              x1="16"
              x2="84"
              y1={16 + (1 - frac) * (height - 32)}
              y2={16 + (1 - frac) * (height - 32)}
              className="stroke-muted"
              strokeWidth="0.2"
              strokeDasharray="1 1"
            />
          ))}

          {/* Area fill */}
          <path d={chartData.areaPath} fill="hsl(var(--primary) / 0.1)" />

          {/* Line */}
          <path
            d={chartData.linePath}
            fill="none"
            stroke="hsl(var(--primary))"
            strokeWidth="0.5"
            strokeLinejoin="round"
          />

          {/* X-axis labels */}
          {chartData.labels.map((label, i) => (
            <text
              // eslint-disable-next-line react/no-array-index-key
              key={i}
              x={label.x}
              y={height - 2}
              textAnchor="middle"
              className="fill-muted-foreground"
              fontSize="3"
            >
              {label.text}
            </text>
          ))}

          {/* Y-axis min/max */}
          <text x="14" y="18" textAnchor="end" className="fill-muted-foreground" fontSize="3">
            {chartData.maxVal.toFixed(1)}
          </text>
          <text
            x="14"
            y={height - 18}
            textAnchor="end"
            className="fill-muted-foreground"
            fontSize="3"
          >
            {chartData.minVal.toFixed(1)}
          </text>
        </svg>
      </CardContent>
    </Card>
  );
}
