// @ts-nocheck
/**
 * Identity Status Card Component
 * VE-701: Display identity status with actions
 */
import * as React from 'react';
import { useIdentity } from '../../hooks/useIdentity';
import { IdentityScoreDisplay } from './IdentityScoreDisplay';
import { formatAddress, formatRelativeTime } from '../../utils/format';
import type { VerificationScope } from '../../types/identity';

/**
 * Identity status card props
 */
export interface IdentityStatusCardProps {
  /**
   * Callback when user wants to start verification
   */
  onStartVerification?: () => void;

  /**
   * Callback when user wants to upgrade scope
   */
  onUpgradeScope?: (scope: VerificationScope) => void;

  /**
   * Custom CSS class
   */
  className?: string;
}

/**
 * Status badge component
 * A11Y: Uses role="status" for screen reader announcement
 */
interface StatusBadgeProps {
  status: string;
}

function StatusBadge({ status }: StatusBadgeProps): JSX.Element {
  // WCAG 2.1 AA compliant color combinations
  const statusConfig: Record<string, { color: string; bg: string; label: string }> = {
    verified: { color: '#166534', bg: '#dcfce7', label: 'Status: Verified' },
    pending: { color: '#854d0e', bg: '#fef9c3', label: 'Status: Pending verification' },
    expired: { color: '#991b1b', bg: '#fee2e2', label: 'Status: Verification expired' },
    failed: { color: '#991b1b', bg: '#fee2e2', label: 'Status: Verification failed' },
    none: { color: '#4b5563', bg: '#f3f4f6', label: 'Status: Not verified' },
  };

  const config = statusConfig[status] || statusConfig.none;
  const displayText = status.charAt(0).toUpperCase() + status.slice(1);

  return (
    <span
      className="status-badge"
      style={{ color: config.color, backgroundColor: config.bg }}
      role="status"
      aria-label={config.label}
    >
      {displayText}
      <style>{`
        .status-badge {
          display: inline-block;
          padding: 4px 12px;
          border-radius: 9999px;
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: capitalize;
        }
      `}</style>
    </span>
  );
}

/**
 * Identity status card component
 */
export function IdentityStatusCard({
  onStartVerification,
  onUpgradeScope,
  className = '',
}: IdentityStatusCardProps): JSX.Element {
  const { state, getScopeRequirements } = useIdentity();
  const { identity, status, isLoading, error } = state;

  if (isLoading) {
    return (
      <div 
        className={`identity-card identity-card--loading ${className}`}
        role="status"
        aria-label="Loading identity information"
        aria-busy="true"
      >
        <div className="identity-card__skeleton" aria-hidden="true" />
        <span className="sr-only">Loading identity information...</span>
        <style>{cardStyles}</style>
      </div>
    );
  }

  if (error) {
    return (
      <div 
        className={`identity-card identity-card--error ${className}`}
        role="alert"
        aria-live="polite"
      >
        <p className="identity-card__error">
          <span aria-hidden="true">⚠ </span>
          Failed to load identity: {error.message}
        </p>
        <style>{cardStyles}</style>
      </div>
    );
  }

  if (!identity) {
    return (
      <div className={`identity-card identity-card--empty ${className}`}>
        <div className="identity-card__content">
          <h3 className="identity-card__title" id="identity-card-title">Verify Your Identity</h3>
          <p className="identity-card__description" id="identity-card-desc">
            Complete identity verification to unlock full platform features.
          </p>
          {onStartVerification && (
            <button
              className="identity-card__button identity-card__button--primary"
              onClick={onStartVerification}
              aria-describedby="identity-card-desc"
            >
              Start Verification
            </button>
          )}
        </div>
        <style>{cardStyles}</style>
      </div>
    );
  }

  const scopes: VerificationScope[] = ['basic', 'standard', 'enhanced', 'provider'];
  const currentScopeIndex = scopes.indexOf(identity.scope);

  return (
    <div 
      className={`identity-card ${className}`}
      role="region"
      aria-labelledby="identity-status-heading"
    >
      <div className="identity-card__header">
        <div className="identity-card__header-left">
          <h3 className="identity-card__title" id="identity-status-heading">Identity Status</h3>
          <StatusBadge status={status} />
        </div>
        <IdentityScoreDisplay score={identity.score} size="md" />
      </div>

      <dl className="identity-card__details" aria-label="Identity details">
        <div className="identity-card__detail">
          <dt className="identity-card__detail-label">VEID</dt>
          <dd className="identity-card__detail-value">
            <span aria-label={`VEID: ${identity.veid}`}>
              {formatAddress(identity.veid, 12)}
            </span>
          </dd>
        </div>
        <div className="identity-card__detail">
          <dt className="identity-card__detail-label">Scope</dt>
          <dd className="identity-card__detail-value identity-card__detail-value--scope">
            {identity.scope.charAt(0).toUpperCase() + identity.scope.slice(1)}
          </dd>
        </div>
        <div className="identity-card__detail">
          <dt className="identity-card__detail-label">Verified</dt>
          <dd className="identity-card__detail-value">
            <time dateTime={new Date(identity.verifiedAt).toISOString()}>
              {formatRelativeTime(identity.verifiedAt)}
            </time>
          </dd>
        </div>
        {identity.expiresAt && (
          <div className="identity-card__detail">
            <dt className="identity-card__detail-label">Expires</dt>
            <dd className="identity-card__detail-value">
              <time dateTime={new Date(identity.expiresAt).toISOString()}>
                {formatRelativeTime(identity.expiresAt)}
              </time>
            </dd>
          </div>
        )}
      </dl>

      {onUpgradeScope && currentScopeIndex < scopes.length - 1 && (
        <div className="identity-card__upgrade" role="group" aria-label="Upgrade options">
          <p className="identity-card__upgrade-text" id="upgrade-description">
            Upgrade to unlock more features
          </p>
          <div className="identity-card__upgrade-options" role="list">
            {scopes.slice(currentScopeIndex + 1).map((scope) => {
              const requirements = getScopeRequirements(scope);
              const scopeLabel = scope.charAt(0).toUpperCase() + scope.slice(1);
              const buttonId = `upgrade-${scope}`;
              const descId = requirements.minimumScore ? `${buttonId}-desc` : undefined;
              return (
                <button
                  key={scope}
                  id={buttonId}
                  className="identity-card__button identity-card__button--outline"
                  onClick={() => onUpgradeScope(scope)}
                  aria-describedby={descId}
                  role="listitem"
                >
                  Upgrade to {scopeLabel}
                  {requirements.minimumScore && (
                    <span className="identity-card__button-hint" id={descId}>
                      Requires {requirements.minimumScore}+ score
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      )}

      <style>{cardStyles}</style>
    </div>
  );
}

const cardStyles = `
  .identity-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .identity-card--loading,
  .identity-card--error {
    min-height: 200px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .identity-card__skeleton {
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, #f3f4f6 25%, #e5e7eb 50%, #f3f4f6 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    border-radius: 8px;
  }

  @keyframes shimmer {
    0% { background-position: 200% 0; }
    100% { background-position: -200% 0; }
  }

  .identity-card__error {
    color: #ef4444;
    font-size: 0.875rem;
  }

  .identity-card--empty .identity-card__content {
    text-align: center;
  }

  .identity-card__header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;
  }

  .identity-card__header-left {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .identity-card__title {
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
  }

  .identity-card__description {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 8px 0 16px;
  }

  .identity-card__details {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
    padding: 16px 0;
    border-top: 1px solid #e5e7eb;
  }

  .identity-card__detail {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .identity-card__detail-label {
    font-size: 0.75rem;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .identity-card__detail-value {
    font-size: 0.875rem;
    color: #111827;
    font-family: monospace;
  }

  .identity-card__detail-value--scope {
    font-family: inherit;
    font-weight: 600;
  }

  .identity-card__upgrade {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
  }

  .identity-card__upgrade-text {
    font-size: 0.875rem;
    color: #6b7280;
    margin: 0 0 12px;
  }

  .identity-card__upgrade-options {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .identity-card__button {
    padding: 10px 20px;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }

  .identity-card__button--primary {
    background: #3b82f6;
    color: white;
  }

  .identity-card__button--primary:hover {
    background: #2563eb;
  }

  .identity-card__button--outline {
    background: white;
    border: 1px solid #e5e7eb;
    color: #374151;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .identity-card__button--outline:hover {
    border-color: #3b82f6;
    color: #3b82f6;
  }

  .identity-card__button-hint {
    font-size: 0.625rem;
    font-weight: 400;
    color: #9ca3af;
  }
`;
