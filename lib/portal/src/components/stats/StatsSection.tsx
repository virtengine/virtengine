/**
 * Network Statistics Section
 * VE Portal Landing Page
 */
import * as React from "react";
import { useChain } from "../../../hooks/useChain";

interface NetworkStats {
  providers: number;
  activeLeases: number;
  totalCpuCores: number;
  totalClusters: number;
  transactions: number;
  tokenSupply: number;
}

export interface StatsSectionProps {
  refreshMs?: number;
  tokenDenom?: string;
  displayDenom?: string;
  className?: string;
}

const defaultStats: NetworkStats = {
  providers: 0,
  activeLeases: 0,
  totalCpuCores: 0,
  totalClusters: 0,
  transactions: 0,
  tokenSupply: 0,
};

export function StatsSection({
  refreshMs = 30000,
  tokenDenom = "uvirt",
  displayDenom = "VIRT",
  className = "",
}: StatsSectionProps): JSX.Element {
  const { queryClient } = useChain();
  const [stats, setStats] = React.useState<NetworkStats>(defaultStats);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = React.useState<number | null>(null);

  const fetchStats = React.useCallback(async () => {
    setIsLoading(true);
    setError(null);

    const providerPromise = queryClient.query<{
      providers?: unknown[];
      pagination?: { total?: string };
    }>("/virtengine/provider/v1beta4/providers", {
      "pagination.limit": "1",
      "pagination.count_total": "true",
    });

    const leasesPromise = queryClient.query<{
      leases?: unknown[];
      pagination?: { total?: string };
    }>("/virtengine/market/v1beta5/leases/list", {
      "pagination.limit": "1",
      "pagination.count_total": "true",
      "filters.state": "active",
    });

    const clustersPromise = queryClient.query<{
      clusters?: Array<{
        cluster_metadata?: { total_cpu_cores?: string | number };
      }>;
      pagination?: { total?: string };
    }>("/virtengine/hpc/v1/clusters", {
      "pagination.limit": "500",
      "pagination.count_total": "true",
    });

    const txPromise = queryClient.query<{
      pagination?: { total?: string };
    }>("/cosmos/tx/v1beta1/txs", {
      "pagination.limit": "1",
      "pagination.count_total": "true",
    });

    const supplyPromise = queryClient.query<{
      amount?: { amount?: string };
    }>("/cosmos/bank/v1beta1/supply/by_denom", {
      denom: tokenDenom,
    });

    const results = await Promise.allSettled([
      providerPromise,
      leasesPromise,
      clustersPromise,
      txPromise,
      supplyPromise,
    ]);

    const nextStats: Partial<NetworkStats> = {};
    let hasAnyData = false;

    const providerResult = results[0];
    if (providerResult.status === "fulfilled") {
      const totalProviders = parseInt(
        providerResult.value.pagination?.total || "0",
        10,
      );
      nextStats.providers = Number.isFinite(totalProviders)
        ? totalProviders
        : 0;
      hasAnyData = true;
    }

    const leasesResult = results[1];
    if (leasesResult.status === "fulfilled") {
      const totalLeases = parseInt(
        leasesResult.value.pagination?.total || "0",
        10,
      );
      nextStats.activeLeases = Number.isFinite(totalLeases) ? totalLeases : 0;
      hasAnyData = true;
    }

    const clustersResult = results[2];
    if (clustersResult.status === "fulfilled") {
      const clusters = clustersResult.value.clusters || [];
      const totalCpu = clusters.reduce((acc, cluster) => {
        const raw = cluster.cluster_metadata?.total_cpu_cores ?? 0;
        const value = typeof raw === "string" ? parseInt(raw, 10) : raw;
        return acc + (Number.isFinite(value) ? value : 0);
      }, 0);
      nextStats.totalCpuCores = totalCpu;
      nextStats.totalClusters = clusters.length;
      hasAnyData = true;
    }

    const txResult = results[3];
    if (txResult.status === "fulfilled") {
      const totalTx = parseInt(txResult.value.pagination?.total || "0", 10);
      nextStats.transactions = Number.isFinite(totalTx) ? totalTx : 0;
      hasAnyData = true;
    }

    const supplyResult = results[4];
    if (supplyResult.status === "fulfilled") {
      const rawSupply = parseFloat(supplyResult.value.amount?.amount || "0");
      nextStats.tokenSupply = Number.isFinite(rawSupply)
        ? rawSupply / 1_000_000
        : 0;
      hasAnyData = true;
    }

    if (!hasAnyData) {
      setError("Unable to load network stats.");
    }

    setStats((prev) => ({ ...prev, ...nextStats }));
    setLastUpdated(Date.now());
    setIsLoading(false);
  }, [queryClient, tokenDenom]);

  React.useEffect(() => {
    let mounted = true;
    const run = async () => {
      if (!mounted) return;
      await fetchStats();
    };

    run();
    const interval = window.setInterval(run, refreshMs);
    return () => {
      mounted = false;
      window.clearInterval(interval);
    };
  }, [fetchStats, refreshMs]);

  return (
    <section
      className={`ve-stats ${className}`}
      aria-labelledby="ve-stats-title"
    >
      <div className="ve-stats__header">
        <div>
          <h2 id="ve-stats-title">Network momentum</h2>
          <p>Real-time network telemetry from the VirtEngine chain.</p>
        </div>
        <div className="ve-stats__meta" aria-live="polite">
          {formatStatusLabel(isLoading, error, lastUpdated)}
        </div>
      </div>

      <div className="ve-stats__grid" role="list">
        <StatCard
          label="Total providers"
          value={stats.providers}
          isLoading={isLoading}
        />
        <StatCard
          label="Active leases"
          value={stats.activeLeases}
          isLoading={isLoading}
        />
        <StatCard
          label="Compute capacity"
          value={stats.totalCpuCores}
          suffix=" CPU cores"
          helper={`${stats.totalClusters} clusters online`}
          isLoading={isLoading}
        />
        <StatCard
          label="Transactions processed"
          value={stats.transactions}
          isLoading={isLoading}
        />
        <StatCard
          label={`${displayDenom} supply`}
          value={stats.tokenSupply}
          suffix={` ${displayDenom}`}
          isLoading={isLoading}
        />
      </div>

      {error && (
        <div className="ve-stats__error" role="alert">
          {error}
        </div>
      )}

      <style>{statsStyles}</style>
    </section>
  );
}

interface StatCardProps {
  label: string;
  value: number;
  suffix?: string;
  helper?: string;
  isLoading?: boolean;
}

function StatCard({
  label,
  value,
  suffix,
  helper,
  isLoading,
}: StatCardProps): JSX.Element {
  return (
    <div className="ve-stat-card" role="listitem">
      <span className="ve-stat-card__label">{label}</span>
      {isLoading ? (
        <div className="ve-stat-card__skeleton" aria-hidden="true" />
      ) : (
        <div className="ve-stat-card__value">
          <AnimatedNumber value={value} />
          {suffix && <span className="ve-stat-card__suffix">{suffix}</span>}
        </div>
      )}
      {helper && <span className="ve-stat-card__helper">{helper}</span>}
    </div>
  );
}

interface AnimatedNumberProps {
  value: number;
  durationMs?: number;
}

function AnimatedNumber({
  value,
  durationMs = 1200,
}: AnimatedNumberProps): JSX.Element {
  const [display, setDisplay] = React.useState(0);

  React.useEffect(() => {
    const prefersReduced = window.matchMedia?.(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    if (prefersReduced) {
      setDisplay(value);
      return;
    }

    let start: number | null = null;
    const initial = display;
    const delta = value - initial;

    const step = (timestamp: number) => {
      if (start === null) start = timestamp;
      const progress = Math.min((timestamp - start) / durationMs, 1);
      setDisplay(initial + delta * easeOutCubic(progress));
      if (progress < 1) {
        requestAnimationFrame(step);
      }
    };

    const animation = requestAnimationFrame(step);
    return () => cancelAnimationFrame(animation);
  }, [value]);

  return <span>{formatCompact(display)}</span>;
}

function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3);
}

function formatCompact(value: number): string {
  if (!Number.isFinite(value)) return "0";
  const formatter = new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: value < 1000 ? 0 : 2,
  });
  return formatter.format(Math.round(value));
}

function formatStatusLabel(
  isLoading: boolean,
  error: string | null,
  lastUpdated: number | null,
): string {
  if (isLoading) return "Syncing...";
  if (error) return "Live feed unavailable";
  if (!lastUpdated) return "Updated just now";

  const seconds = Math.max(0, Math.floor((Date.now() - lastUpdated) / 1000));
  if (seconds < 10) return "Updated just now";
  if (seconds < 60) return `Updated ${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  return `Updated ${minutes}m ago`;
}

const statsStyles = `
  .ve-stats {
    padding: 80px 0;
    background: #f8fafc;
    color: #0f172a;
  }

  .ve-stats__header {
    width: min(1120px, 90vw);
    margin: 0 auto 36px;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 16px;
  }

  .ve-stats__header h2 {
    font-family: "Space Grotesk", "Manrope", "Segoe UI", sans-serif;
    font-size: clamp(1.8rem, 2.4vw, 2.4rem);
    margin: 0 0 8px;
  }

  .ve-stats__header p {
    margin: 0;
    color: #475569;
    max-width: 480px;
  }

  .ve-stats__meta {
    font-size: 0.85rem;
    color: #0f172a;
    background: #e2e8f0;
    padding: 6px 12px;
    border-radius: 999px;
    font-weight: 600;
  }

  .ve-stats__grid {
    width: min(1120px, 90vw);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 18px;
  }

  .ve-stat-card {
    background: #ffffff;
    border-radius: 18px;
    padding: 20px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 14px 34px rgba(15, 23, 42, 0.08);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .ve-stat-card__label {
    font-size: 0.85rem;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.12em;
  }

  .ve-stat-card__value {
    font-size: 1.8rem;
    font-weight: 700;
    color: #0f172a;
    display: flex;
    align-items: baseline;
    gap: 6px;
  }

  .ve-stat-card__suffix {
    font-size: 0.85rem;
    color: #475569;
    font-weight: 500;
  }

  .ve-stat-card__helper {
    font-size: 0.85rem;
    color: #475569;
  }

  .ve-stat-card__skeleton {
    height: 32px;
    border-radius: 10px;
    background: linear-gradient(90deg, #e2e8f0, #f8fafc, #e2e8f0);
    background-size: 200% 100%;
    animation: statsLoading 1.4s ease infinite;
  }

  .ve-stats__error {
    margin: 20px auto 0;
    width: min(1120px, 90vw);
    background: #fee2e2;
    color: #b91c1c;
    border-radius: 12px;
    padding: 12px 16px;
    font-size: 0.9rem;
  }

  @keyframes statsLoading {
    0% { background-position: 0% 50%; }
    100% { background-position: 100% 50%; }
  }

  @media (max-width: 860px) {
    .ve-stats__header {
      flex-direction: column;
      align-items: flex-start;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .ve-stat-card__skeleton {
      animation: none;
    }
  }
`;
