/**
 * Offering Card Component
 * VE-703: Display marketplace offering
 */
import * as React from 'react';
import { formatTokenAmount } from '../../utils/format';
import type { Offering } from '../../types/marketplace';

/**
 * Offering card props
 */
export interface OfferingCardProps {
  /**
   * Offering to display
   */
  offering: Offering;

  /**
   * Callback when offering is selected
   */
  onSelect?: (offering: Offering) => void;

  /**
   * Whether this card is selected
   */
  isSelected?: boolean;

  /**
   * Custom CSS class
   */
  className?: string;
}

/**
 * Get offering type label and color
 */
function getOfferingTypeConfig(type: string): { label: string; color: string; bg: string } {
  const configs: Record<string, { label: string; color: string; bg: string }> = {
    compute: { label: 'Compute', color: '#7c3aed', bg: '#ede9fe' },
    storage: { label: 'Storage', color: '#0891b2', bg: '#cffafe' },
    gpu: { label: 'GPU', color: '#dc2626', bg: '#fee2e2' },
    ai: { label: 'AI', color: '#ea580c', bg: '#ffedd5' },
    hpc: { label: 'HPC', color: '#4f46e5', bg: '#e0e7ff' },
  };
  return configs[type] || { label: type, color: '#6b7280', bg: '#f3f4f6' };
}

/**
 * Offering card component
 */
export function OfferingCard({
  offering,
  onSelect,
  isSelected = false,
  className = '',
}: OfferingCardProps): JSX.Element {
  const typeConfig = getOfferingTypeConfig(offering.offeringType);

  const handleClick = () => {
    if (onSelect) {
      onSelect(offering);
    }
  };

  return (
    <div
      className={`offering-card ${isSelected ? 'offering-card--selected' : ''} ${className}`}
      onClick={handleClick}
      role={onSelect ? 'button' : undefined}
      tabIndex={onSelect ? 0 : undefined}
    >
      {/* Header */}
      <div className="offering-card__header">
        <span
          className="offering-card__type"
          style={{ color: typeConfig.color, backgroundColor: typeConfig.bg }}
        >
          {typeConfig.label}
        </span>
        {offering.teeEnabled && (
          <span className="offering-card__tee">🔒 TEE</span>
        )}
      </div>

      {/* Title & Provider */}
      <h3 className="offering-card__title">{offering.name}</h3>
      <p className="offering-card__provider">by {offering.providerName}</p>

      {/* Description */}
      <p className="offering-card__description">{offering.description}</p>

      {/* Specs */}
      <div className="offering-card__specs">
        {offering.specs.cpu && (
          <div className="offering-card__spec">
            <span className="offering-card__spec-label">CPU</span>
            <span className="offering-card__spec-value">{offering.specs.cpu} cores</span>
          </div>
        )}
        {offering.specs.memory && (
          <div className="offering-card__spec">
            <span className="offering-card__spec-label">Memory</span>
            <span className="offering-card__spec-value">{offering.specs.memory}</span>
          </div>
        )}
        {offering.specs.storage && (
          <div className="offering-card__spec">
            <span className="offering-card__spec-label">Storage</span>
            <span className="offering-card__spec-value">{offering.specs.storage}</span>
          </div>
        )}
        {offering.specs.gpu && (
          <div className="offering-card__spec">
            <span className="offering-card__spec-label">GPU</span>
            <span className="offering-card__spec-value">{offering.specs.gpu}</span>
          </div>
        )}
      </div>

      {/* Price */}
      <div className="offering-card__price">
        <span className="offering-card__price-amount">
          {formatTokenAmount(offering.pricePerHour, 6, 'VE')}
        </span>
        <span className="offering-card__price-unit">/ hour</span>
      </div>

      {/* Availability */}
      <div className="offering-card__availability">
        <span
          className={`offering-card__status offering-card__status--${offering.availability}`}
        >
          {offering.availability === 'available' ? '● Available' :
           offering.availability === 'limited' ? '◐ Limited' : '○ Unavailable'}
        </span>
        {offering.regions && (
          <span className="offering-card__regions">
            📍 {offering.regions.join(', ')}
          </span>
        )}
      </div>

      <style>{cardStyles}</style>
    </div>
  );
}

const cardStyles = `
  .offering-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .offering-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .offering-card--selected {
    border-color: #3b82f6;
    background: #f8fafc;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
  }

  .offering-card__header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .offering-card__type {
    padding: 4px 10px;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .offering-card__tee {
    font-size: 0.75rem;
    color: #16a34a;
  }

  .offering-card__title {
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px;
  }

  .offering-card__provider {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0 0 12px;
  }

  .offering-card__description {
    font-size: 0.875rem;
    color: #374151;
    margin: 0 0 16px;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .offering-card__specs {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    margin-bottom: 16px;
  }

  .offering-card__spec {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .offering-card__spec-label {
    font-size: 0.625rem;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .offering-card__spec-value {
    font-size: 0.875rem;
    color: #111827;
    font-weight: 500;
  }

  .offering-card__price {
    display: flex;
    align-items: baseline;
    gap: 4px;
    margin-bottom: 12px;
  }

  .offering-card__price-amount {
    font-size: 1.25rem;
    font-weight: 700;
    color: #111827;
  }

  .offering-card__price-unit {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .offering-card__availability {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 12px;
    border-top: 1px solid #f3f4f6;
  }

  .offering-card__status {
    font-size: 0.75rem;
    font-weight: 500;
  }

  .offering-card__status--available {
    color: #16a34a;
  }

  .offering-card__status--limited {
    color: #ca8a04;
  }

  .offering-card__status--unavailable {
    color: #dc2626;
  }

  .offering-card__regions {
    font-size: 0.75rem;
    color: #6b7280;
  }
`;
