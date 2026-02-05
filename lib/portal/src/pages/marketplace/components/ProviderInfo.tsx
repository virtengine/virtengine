/**
 * Provider Info Component
 * VE-703: Display provider information and ratings
 */
import * as React from 'react';
import { formatAddress } from '../../../../utils/format';

export interface ProviderInfoProps {
  providerAddress: string;
  providerName: string;
  reliabilityScore: number;
  benchmarkSummary?: {
    cpuScore: number;
    memoryScore: number;
    storageScore: number;
    networkScore: number;
    gpuScore?: number;
    overallScore: number;
    lastBenchmarkAt: number;
  };
  totalOfferings?: number;
  totalOrders?: number;
  isVerified?: boolean;
  className?: string;
}

export function ProviderInfo({
  providerAddress,
  providerName,
  reliabilityScore,
  benchmarkSummary,
  totalOfferings,
  totalOrders,
  isVerified = false,
  className = '',
}: ProviderInfoProps): JSX.Element {
  const reliabilityColor = getScoreColor(reliabilityScore);

  return (
    <div className={`provider-info ${className}`}>
      <div className="provider-info__header">
        <div className="provider-info__identity">
          <ProviderAvatar name={providerName} address={providerAddress} />
          <div className="provider-info__details">
            <div className="provider-info__name-row">
              <h3 className="provider-info__name">{providerName}</h3>
              {isVerified && (
                <span
                  className="provider-info__verified"
                  title="Verified provider"
                  aria-label="Verified provider"
                >
                  <VerifiedBadge />
                </span>
              )}
            </div>
            <a
              href={`/providers/${providerAddress}`}
              className="provider-info__address"
              title={providerAddress}
            >
              {formatAddress(providerAddress, 8)}
            </a>
          </div>
        </div>

        <div className="provider-info__reliability" role="meter" aria-valuenow={reliabilityScore} aria-valuemin={0} aria-valuemax={100} aria-label="Reliability score">
          <div className="provider-info__score" style={{ color: reliabilityColor }}>
            {reliabilityScore}
          </div>
          <div className="provider-info__score-label">Reliability</div>
        </div>
      </div>

      {/* Stats */}
      {(totalOfferings !== undefined || totalOrders !== undefined) && (
        <div className="provider-info__stats">
          {totalOfferings !== undefined && (
            <div className="provider-info__stat">
              <span className="provider-info__stat-value">{totalOfferings}</span>
              <span className="provider-info__stat-label">Offerings</span>
            </div>
          )}
          {totalOrders !== undefined && (
            <div className="provider-info__stat">
              <span className="provider-info__stat-value">{totalOrders}</span>
              <span className="provider-info__stat-label">Orders</span>
            </div>
          )}
        </div>
      )}

      {/* Benchmark Summary */}
      {benchmarkSummary && (
        <div className="provider-info__benchmarks">
          <h4 className="provider-info__benchmarks-title">Performance Benchmarks</h4>
          <div className="provider-info__benchmark-grid">
            <BenchmarkItem label="CPU" score={benchmarkSummary.cpuScore} />
            <BenchmarkItem label="Memory" score={benchmarkSummary.memoryScore} />
            <BenchmarkItem label="Storage" score={benchmarkSummary.storageScore} />
            <BenchmarkItem label="Network" score={benchmarkSummary.networkScore} />
            {benchmarkSummary.gpuScore !== undefined && (
              <BenchmarkItem label="GPU" score={benchmarkSummary.gpuScore} />
            )}
          </div>
          <div className="provider-info__benchmark-overall">
            <span className="provider-info__benchmark-overall-label">Overall Score</span>
            <span className="provider-info__benchmark-overall-value" style={{ color: getScoreColor(benchmarkSummary.overallScore) }}>
              {benchmarkSummary.overallScore}
            </span>
          </div>
        </div>
      )}

      <style>{providerInfoStyles}</style>
    </div>
  );
}

interface ProviderAvatarProps {
  name: string;
  address: string;
}

function ProviderAvatar({ name, address }: ProviderAvatarProps): JSX.Element {
  const initials = name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  const hue = hashStringToNumber(address) % 360;
  const backgroundColor = `hsl(${hue}, 60%, 90%)`;
  const color = `hsl(${hue}, 60%, 35%)`;

  return (
    <div
      className="provider-avatar"
      style={{ backgroundColor, color }}
      aria-hidden="true"
    >
      {initials}
    </div>
  );
}

function hashStringToNumber(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }
  return Math.abs(hash);
}

interface BenchmarkItemProps {
  label: string;
  score: number;
}

function BenchmarkItem({ label, score }: BenchmarkItemProps): JSX.Element {
  const color = getScoreColor(score);
  const percentage = Math.min(100, Math.max(0, score));

  return (
    <div className="benchmark-item">
      <div className="benchmark-item__header">
        <span className="benchmark-item__label">{label}</span>
        <span className="benchmark-item__score" style={{ color }}>{score}</span>
      </div>
      <div className="benchmark-item__bar" role="progressbar" aria-valuenow={score} aria-valuemin={0} aria-valuemax={100} aria-label={`${label} score`}>
        <div
          className="benchmark-item__fill"
          style={{ width: `${percentage}%`, backgroundColor: color }}
        />
      </div>
    </div>
  );
}

function getScoreColor(score: number): string {
  if (score >= 80) return '#16a34a'; // Green
  if (score >= 60) return '#ca8a04'; // Yellow
  if (score >= 40) return '#ea580c'; // Orange
  return '#dc2626'; // Red
}

function VerifiedBadge(): JSX.Element {
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="currentColor"
      className="verified-badge"
    >
      <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z" />
    </svg>
  );
}

/**
 * Compact Provider Badge for use in cards
 */
export interface ProviderBadgeProps {
  providerName: string;
  providerAddress: string;
  reliabilityScore: number;
  isVerified?: boolean;
  className?: string;
}

export function ProviderBadge({
  providerName,
  providerAddress,
  reliabilityScore,
  isVerified = false,
  className = '',
}: ProviderBadgeProps): JSX.Element {
  const reliabilityColor = getScoreColor(reliabilityScore);

  return (
    <div className={`provider-badge ${className}`}>
      <ProviderAvatar name={providerName} address={providerAddress} />
      <div className="provider-badge__info">
        <span className="provider-badge__name">
          {providerName}
          {isVerified && (
            <span className="provider-badge__verified" aria-label="Verified">
              <VerifiedBadge />
            </span>
          )}
        </span>
        <span className="provider-badge__score" style={{ color: reliabilityColor }}>
          {reliabilityScore}% reliability
        </span>
      </div>
      <style>{providerBadgeStyles}</style>
    </div>
  );
}

const providerInfoStyles = `
  .provider-info {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
  }

  .provider-info__header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 16px;
  }

  .provider-info__identity {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .provider-avatar {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1rem;
    font-weight: 600;
    flex-shrink: 0;
  }

  .provider-info__details {
    min-width: 0;
  }

  .provider-info__name-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .provider-info__name {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .provider-info__verified {
    color: #3b82f6;
    display: flex;
  }

  .provider-info__address {
    font-size: 0.75rem;
    color: #6b7280;
    font-family: monospace;
    text-decoration: none;
  }

  .provider-info__address:hover {
    color: #3b82f6;
    text-decoration: underline;
  }

  .provider-info__reliability {
    text-align: center;
  }

  .provider-info__score {
    font-size: 1.5rem;
    font-weight: 700;
  }

  .provider-info__score-label {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .provider-info__stats {
    display: flex;
    gap: 24px;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #f3f4f6;
  }

  .provider-info__stat {
    display: flex;
    flex-direction: column;
  }

  .provider-info__stat-value {
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
  }

  .provider-info__stat-label {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .provider-info__benchmarks {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #f3f4f6;
  }

  .provider-info__benchmarks-title {
    margin: 0 0 12px;
    font-size: 0.875rem;
    font-weight: 600;
    color: #374151;
  }

  .provider-info__benchmark-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .benchmark-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .benchmark-item__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .benchmark-item__label {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .benchmark-item__score {
    font-size: 0.75rem;
    font-weight: 600;
  }

  .benchmark-item__bar {
    height: 4px;
    background: #f3f4f6;
    border-radius: 2px;
    overflow: hidden;
  }

  .benchmark-item__fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.3s ease;
  }

  .provider-info__benchmark-overall {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f3f4f6;
  }

  .provider-info__benchmark-overall-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .provider-info__benchmark-overall-value {
    font-size: 1.25rem;
    font-weight: 700;
  }

  @media (prefers-reduced-motion: reduce) {
    .benchmark-item__fill {
      transition: none;
    }
  }
`;

const providerBadgeStyles = `
  .provider-badge {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .provider-badge .provider-avatar {
    width: 32px;
    height: 32px;
    font-size: 0.75rem;
    border-radius: 8px;
  }

  .provider-badge__info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .provider-badge__name {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
  }

  .provider-badge__verified {
    color: #3b82f6;
    display: flex;
  }

  .provider-badge__verified svg {
    width: 14px;
    height: 14px;
  }

  .provider-badge__score {
    font-size: 0.75rem;
    font-weight: 500;
  }
`;
