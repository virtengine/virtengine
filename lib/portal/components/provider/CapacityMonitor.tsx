/**
 * Capacity Monitor Component
 * VE-705: Monitor provider resource capacity
 */
import * as React from 'react';

export interface CapacityMetrics {
  cpu: { used: number; total: number };
  memory: { used: number; total: number };
  storage: { used: number; total: number };
  gpu?: { used: number; total: number };
}

export interface CapacityMonitorProps {
  metrics: CapacityMetrics;
  refreshInterval?: number;
  onRefresh?: () => void;
  showGpu?: boolean;
  className?: string;
}

function ProgressBar({ used, total, label }: { used: number; total: number; label: string }): JSX.Element {
  const percentage = total > 0 ? Math.round((used / total) * 100) : 0;
  const color = percentage > 90 ? '#dc2626' : percentage > 70 ? '#f59e0b' : '#16a34a';

  return (
    <div style={{ marginBottom: '16px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
        <span style={{ fontSize: '14px', fontWeight: 500 }}>{label}</span>
        <span style={{ fontSize: '14px', color: '#666' }}>{used} / {total} ({percentage}%)</span>
      </div>
      <div style={{ height: '8px', backgroundColor: '#e5e7eb', borderRadius: '4px', overflow: 'hidden' }}>
        <div
          style={{
            height: '100%',
            width: `${percentage}%`,
            backgroundColor: color,
            transition: 'width 0.3s',
          }}
        />
      </div>
    </div>
  );
}

export function CapacityMonitor({
  metrics,
  refreshInterval,
  onRefresh,
  showGpu = true,
  className,
}: CapacityMonitorProps): JSX.Element {
  React.useEffect(() => {
    if (refreshInterval && onRefresh) {
      const interval = setInterval(onRefresh, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [refreshInterval, onRefresh]);

  return (
    <div className={className} style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h3 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>
          Capacity Monitor
        </h3>
        {onRefresh && (
          <button
            onClick={onRefresh}
            style={{
              padding: '8px 16px',
              fontSize: '14px',
              color: '#374151',
              backgroundColor: '#f3f4f6',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Refresh
          </button>
        )}
      </div>

      <ProgressBar used={metrics.cpu.used} total={metrics.cpu.total} label="CPU (cores)" />
      <ProgressBar used={metrics.memory.used} total={metrics.memory.total} label="Memory (GB)" />
      <ProgressBar used={metrics.storage.used} total={metrics.storage.total} label="Storage (GB)" />
      {showGpu && metrics.gpu && (
        <ProgressBar used={metrics.gpu.used} total={metrics.gpu.total} label="GPU (units)" />
      )}
    </div>
  );
}
