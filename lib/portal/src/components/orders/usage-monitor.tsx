import * as React from "react";
import type {
  OrderUsageSnapshot,
  OrderUsageAlert,
} from "../../hooks/useOrderTracking";

export interface UsageMonitorProps {
  usage?: OrderUsageSnapshot;
  onAlertUpdate?: (alert: OrderUsageAlert) => void;
  className?: string;
}

const formatNumber = (value: number, unit: string) => {
  if (unit === "%") return `${Math.round(value)}%`;
  if (unit === "GB") return `${value.toFixed(1)} GB`;
  if (unit === "USD") return `$${value.toFixed(2)}`;
  return `${value.toFixed(1)} ${unit}`;
};

const buildSparkline = (history: OrderUsageSnapshot["history"]) => {
  if (!history.length) return "";
  const width = 160;
  const height = 50;
  const max = Math.max(...history.map((item) => item.cpu));
  const min = Math.min(...history.map((item) => item.cpu));
  const range = Math.max(1, max - min);

  return history
    .map((item, index) => {
      const x = (index / Math.max(1, history.length - 1)) * width;
      const y = height - ((item.cpu - min) / range) * height;
      return `${x},${y}`;
    })
    .join(" ");
};

export function UsageMonitor({
  usage,
  onAlertUpdate,
  className = "",
}: UsageMonitorProps): JSX.Element {
  const [alertThreshold, setAlertThreshold] = React.useState(80);

  if (!usage) {
    return (
      <section className={`ve-usage-monitor ${className}`}>
        <h3>Usage monitoring</h3>
        <p>Usage metrics will stream once the order is running.</p>
        <style>{usageMonitorStyles}</style>
      </section>
    );
  }

  const sparkline = buildSparkline(usage.history);

  return (
    <section className={`ve-usage-monitor ${className}`}>
      <header className="ve-usage-monitor__header">
        <div>
          <h3>Usage monitoring</h3>
          <p>Real-time resource utilization, cost, and remaining balance.</p>
        </div>
        <div className="ve-usage-monitor__cost">
          <div>
            <span>Cost accrued</span>
            <strong>
              {usage.costAccrued} {usage.costCurrency}
            </strong>
          </div>
          <div>
            <span>Remaining balance</span>
            <strong>
              {usage.remainingBalance} {usage.costCurrency}
            </strong>
          </div>
        </div>
      </header>

      <div className="ve-usage-monitor__grid">
        {usage.metrics.map((metric) => {
          const percentage = metric.limit
            ? Math.min(100, Math.round((metric.value / metric.limit) * 100))
            : Math.round(metric.value);
          return (
            <div key={metric.id} className="ve-usage-monitor__metric">
              <div>
                <strong>{metric.label}</strong>
                <span>{formatNumber(metric.value, metric.unit)}</span>
              </div>
              <div className="ve-usage-monitor__bar">
                <span style={{ width: `${percentage}%` }} />
              </div>
              {metric.limit && (
                <small>
                  {percentage}% of {metric.limit} {metric.unit}
                </small>
              )}
            </div>
          );
        })}
      </div>

      <div className="ve-usage-monitor__chart">
        <div>
          <strong>CPU trend</strong>
          <span>Last 24h</span>
        </div>
        <svg viewBox="0 0 160 50" role="img" aria-label="CPU usage trend">
          <polyline
            points={sparkline}
            fill="none"
            stroke="#0f172a"
            strokeWidth="2"
          />
        </svg>
      </div>

      <div className="ve-usage-monitor__alerts">
        <div>
          <h4>Usage alerts</h4>
          <p>Stay ahead of budget or performance limits.</p>
        </div>
        <div className="ve-usage-monitor__alert-form">
          <input
            type="number"
            min={10}
            max={100}
            value={alertThreshold}
            onChange={(event) => setAlertThreshold(Number(event.target.value))}
          />
          <button
            type="button"
            onClick={() =>
              onAlertUpdate?.({
                id: `alert-${Date.now()}`,
                type: "usage",
                label: "CPU threshold",
                threshold: alertThreshold,
                current: usage.metrics[0]?.value ?? 0,
                status:
                  alertThreshold < (usage.metrics[0]?.value ?? 0)
                    ? "warning"
                    : "ok",
                createdAt: Date.now(),
              })
            }
          >
            Add alert
          </button>
        </div>
        <div className="ve-usage-monitor__alert-list">
          {usage.alerts.length === 0 ? (
            <p>No alerts configured yet.</p>
          ) : (
            usage.alerts.map((alert) => (
              <div
                key={alert.id}
                className={`ve-usage-monitor__alert ve-alert-${alert.status}`}
              >
                <div>
                  <strong>{alert.label}</strong>
                  <span>Threshold: {alert.threshold}%</span>
                </div>
                <span>{alert.status.toUpperCase()}</span>
              </div>
            ))
          )}
        </div>
      </div>

      <style>{usageMonitorStyles}</style>
    </section>
  );
}

const usageMonitorStyles = `
  .ve-usage-monitor {
    background: #fff;
    border-radius: 22px;
    padding: 20px;
    border: 1px solid #e2e8f0;
    display: grid;
    gap: 18px;
  }

  .ve-usage-monitor__header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
  }

  .ve-usage-monitor__header h3 {
    margin: 0;
    font-size: 1.2rem;
    font-family: 'Space Grotesk', 'DM Sans', sans-serif;
  }

  .ve-usage-monitor__header p {
    margin: 6px 0 0;
    color: #64748b;
  }

  .ve-usage-monitor__cost {
    display: grid;
    gap: 6px;
    background: #0f172a;
    color: #fff;
    padding: 12px 16px;
    border-radius: 16px;
  }

  .ve-usage-monitor__cost span {
    font-size: 0.75rem;
    color: #cbd5f5;
  }

  .ve-usage-monitor__cost strong {
    display: block;
    font-size: 1rem;
  }

  .ve-usage-monitor__grid {
    display: grid;
    gap: 12px;
  }

  .ve-usage-monitor__metric {
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    padding: 12px;
    display: grid;
    gap: 6px;
    background: #f8fafc;
  }

  .ve-usage-monitor__metric strong {
    display: block;
    font-size: 0.9rem;
  }

  .ve-usage-monitor__metric span {
    color: #475569;
    font-size: 0.85rem;
  }

  .ve-usage-monitor__bar {
    height: 6px;
    background: #e2e8f0;
    border-radius: 999px;
    overflow: hidden;
  }

  .ve-usage-monitor__bar span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, #0f172a, #38bdf8);
  }

  .ve-usage-monitor__chart {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: center;
    padding: 12px 16px;
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    background: #f8fafc;
  }

  .ve-usage-monitor__chart span {
    display: block;
    font-size: 0.75rem;
    color: #64748b;
  }

  .ve-usage-monitor__alerts {
    display: grid;
    gap: 12px;
  }

  .ve-usage-monitor__alerts h4 {
    margin: 0;
  }

  .ve-usage-monitor__alert-form {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .ve-usage-monitor__alert-form input {
    width: 90px;
    padding: 6px 10px;
    border-radius: 10px;
    border: 1px solid #e2e8f0;
  }

  .ve-usage-monitor__alert-form button {
    padding: 6px 12px;
    border-radius: 999px;
    border: none;
    background: #0f172a;
    color: #fff;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .ve-usage-monitor__alert-list {
    display: grid;
    gap: 8px;
  }

  .ve-usage-monitor__alert {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-radius: 12px;
    padding: 10px 12px;
    border: 1px solid #e2e8f0;
    background: #f8fafc;
    font-size: 0.85rem;
  }

  .ve-alert-warning {
    border-color: #f59e0b;
    color: #92400e;
    background: #fffbeb;
  }

  .ve-alert-critical {
    border-color: #ef4444;
    color: #991b1b;
    background: #fef2f2;
  }

  @media (max-width: 960px) {
    .ve-usage-monitor__header {
      flex-direction: column;
    }

    .ve-usage-monitor__chart {
      flex-direction: column;
      align-items: flex-start;
    }
  }
`;
