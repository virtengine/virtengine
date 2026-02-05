/**
 * Offering Detail Page Component
 * VE-703: Full offering detail view with pricing calculator
 */
import * as React from 'react';
import { useState, useCallback, useMemo } from 'react';
import type { Offering, PriceComponent } from '../../../../types/marketplace';
import { formatTokenAmount, formatDuration } from '../../../../utils/format';
import { ProviderInfo } from './ProviderInfo';

export interface OfferingDetailPageProps {
  offering: Offering;
  onBack: () => void;
  onCheckout: (durationSeconds: number) => void;
  isLoading?: boolean;
  className?: string;
}

export function OfferingDetailPage({
  offering,
  onBack,
  onCheckout,
  isLoading = false,
  className = '',
}: OfferingDetailPageProps): JSX.Element {
  const [durationHours, setDurationHours] = useState(24);
  const [durationUnit, setDurationUnit] = useState<'hours' | 'days' | 'months'>('hours');

  const durationInSeconds = useMemo(() => {
    switch (durationUnit) {
      case 'hours':
        return durationHours * 3600;
      case 'days':
        return durationHours * 86400;
      case 'months':
        return durationHours * 30 * 86400;
      default:
        return durationHours * 3600;
    }
  }, [durationHours, durationUnit]);

  const pricing = useMemo(() => {
    return calculatePricing(offering, durationInSeconds);
  }, [offering, durationInSeconds]);

  const handleDurationChange = useCallback((value: string) => {
    const num = parseInt(value, 10);
    if (!isNaN(num) && num > 0) {
      setDurationHours(num);
    }
  }, []);

  const handleCheckout = useCallback(() => {
    onCheckout(durationInSeconds);
  }, [onCheckout, durationInSeconds]);

  const isWithinDurationLimits =
    durationInSeconds >= offering.pricing.minDurationSeconds &&
    (!offering.pricing.maxDurationSeconds || durationInSeconds <= offering.pricing.maxDurationSeconds);

  const typeConfig = getOfferingTypeConfig(offering.type);

  return (
    <div className={`offering-detail-page ${className}`}>
      {/* Back Button */}
      <button
        type="button"
        className="offering-detail-page__back"
        onClick={onBack}
        aria-label="Back to offerings"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M19 12H5M12 19l-7-7 7-7" />
        </svg>
        Back to Marketplace
      </button>

      <div className="offering-detail-page__layout">
        {/* Main Content */}
        <main className="offering-detail-page__main">
          {/* Header */}
          <header className="offering-detail-page__header">
            <div className="offering-detail-page__badges">
              <span
                className="offering-detail-page__type"
                style={{ color: typeConfig.color, backgroundColor: typeConfig.bg }}
              >
                {typeConfig.label}
              </span>
              {offering.hasEncryptedDetails && (
                <span className="offering-detail-page__tee">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                  </svg>
                  TEE Enabled
                </span>
              )}
              <StatusBadge status={offering.status} />
            </div>

            <h1 className="offering-detail-page__title">{offering.title}</h1>
            <p className="offering-detail-page__description">{offering.description}</p>
          </header>

          {/* Specifications */}
          <section className="offering-detail-page__section">
            <h2 className="offering-detail-page__section-title">Resource Specifications</h2>
            <div className="offering-detail-page__specs">
              <SpecItem
                icon={<CpuIcon />}
                label="CPU Cores"
                value={`${offering.resources.cpuCores} vCPU`}
              />
              <SpecItem
                icon={<MemoryIcon />}
                label="Memory"
                value={`${offering.resources.memoryGB} GB`}
              />
              <SpecItem
                icon={<StorageIcon />}
                label="Storage"
                value={`${offering.resources.storageGB} GB`}
              />
              {offering.resources.gpuCount && (
                <SpecItem
                  icon={<GpuIcon />}
                  label="GPU"
                  value={`${offering.resources.gpuCount}x ${offering.resources.gpuModel || 'GPU'}`}
                />
              )}
              {offering.resources.bandwidthGbps && (
                <SpecItem
                  icon={<NetworkIcon />}
                  label="Bandwidth"
                  value={`${offering.resources.bandwidthGbps} Gbps`}
                />
              )}
              <SpecItem
                icon={<LocationIcon />}
                label="Region"
                value={offering.region}
              />
            </div>

            {/* Additional Attributes */}
            {Object.keys(offering.resources.attributes).length > 0 && (
              <div className="offering-detail-page__attributes">
                <h3 className="offering-detail-page__attributes-title">Additional Attributes</h3>
                <dl className="offering-detail-page__attributes-list">
                  {Object.entries(offering.resources.attributes).map(([key, value]) => (
                    <div key={key} className="offering-detail-page__attribute">
                      <dt>{formatAttributeKey(key)}</dt>
                      <dd>{value}</dd>
                    </div>
                  ))}
                </dl>
              </div>
            )}
          </section>

          {/* Pricing Breakdown */}
          {offering.pricing.components && offering.pricing.components.length > 0 && (
            <section className="offering-detail-page__section">
              <h2 className="offering-detail-page__section-title">Component Pricing</h2>
              <div className="offering-detail-page__pricing-components">
                {offering.pricing.components.map((component, index) => (
                  <PricingComponentRow key={index} component={component} denom={offering.pricing.denom} />
                ))}
              </div>
            </section>
          )}

          {/* Identity Requirements */}
          <section className="offering-detail-page__section">
            <h2 className="offering-detail-page__section-title">Requirements</h2>
            <div className="offering-detail-page__requirements">
              <div className="offering-detail-page__requirement">
                <span className="offering-detail-page__requirement-label">Identity Score</span>
                <span className="offering-detail-page__requirement-value">
                  {offering.identityRequirements.minScore > 0
                    ? `Min. ${offering.identityRequirements.minScore} required`
                    : 'No minimum'}
                </span>
              </div>
              {offering.identityRequirements.requiredScopes.length > 0 && (
                <div className="offering-detail-page__requirement">
                  <span className="offering-detail-page__requirement-label">Verification Scopes</span>
                  <div className="offering-detail-page__scopes">
                    {offering.identityRequirements.requiredScopes.map((scope) => (
                      <span key={scope} className="offering-detail-page__scope">{scope}</span>
                    ))}
                  </div>
                </div>
              )}
              <div className="offering-detail-page__requirement">
                <span className="offering-detail-page__requirement-label">MFA Required</span>
                <span className="offering-detail-page__requirement-value">
                  {offering.identityRequirements.mfaRequired ? 'Yes' : 'No'}
                </span>
              </div>
            </div>
          </section>
        </main>

        {/* Sidebar */}
        <aside className="offering-detail-page__sidebar">
          {/* Provider Info */}
          <ProviderInfo
            providerAddress={offering.providerAddress}
            providerName={offering.providerName}
            reliabilityScore={offering.reliabilityScore}
            benchmarkSummary={offering.benchmarkSummary}
            isVerified={true}
          />

          {/* Pricing Calculator */}
          <div className="pricing-calculator">
            <h2 className="pricing-calculator__title">Order Configuration</h2>

            {/* Duration Input */}
            <div className="pricing-calculator__field">
              <label className="pricing-calculator__label">Duration</label>
              <div className="pricing-calculator__duration">
                <input
                  type="number"
                  min="1"
                  className="pricing-calculator__input"
                  value={durationHours}
                  onChange={(e) => handleDurationChange(e.target.value)}
                  aria-label="Duration amount"
                />
                <select
                  className="pricing-calculator__select"
                  value={durationUnit}
                  onChange={(e) => setDurationUnit(e.target.value as 'hours' | 'days' | 'months')}
                  aria-label="Duration unit"
                >
                  <option value="hours">Hours</option>
                  <option value="days">Days</option>
                  <option value="months">Months</option>
                </select>
              </div>
              <div className="pricing-calculator__duration-info">
                <span>
                  Min: {formatDuration(offering.pricing.minDurationSeconds)}
                </span>
                {offering.pricing.maxDurationSeconds && (
                  <span>
                    Max: {formatDuration(offering.pricing.maxDurationSeconds)}
                  </span>
                )}
              </div>
              {!isWithinDurationLimits && (
                <div className="pricing-calculator__error" role="alert">
                  Duration must be between {formatDuration(offering.pricing.minDurationSeconds)}
                  {offering.pricing.maxDurationSeconds && ` and ${formatDuration(offering.pricing.maxDurationSeconds)}`}
                </div>
              )}
            </div>

            {/* Price Breakdown */}
            <div className="pricing-calculator__breakdown">
              <div className="pricing-calculator__row">
                <span>Base Price</span>
                <span>{formatTokenAmount(pricing.baseAmount, 6, offering.pricing.denom)}</span>
              </div>
              <div className="pricing-calculator__row">
                <span>Deposit</span>
                <span>{formatTokenAmount(pricing.deposit, 6, offering.pricing.denom)}</span>
              </div>
              <div className="pricing-calculator__row pricing-calculator__row--total">
                <span>Total</span>
                <span>{formatTokenAmount(pricing.total, 6, offering.pricing.denom)}</span>
              </div>
            </div>

            {/* Checkout Button */}
            <button
              type="button"
              className="pricing-calculator__checkout"
              onClick={handleCheckout}
              disabled={isLoading || !isWithinDurationLimits || offering.status !== 'active'}
            >
              {isLoading ? (
                <>
                  <LoadingSpinner /> Processing...
                </>
              ) : (
                'Proceed to Checkout'
              )}
            </button>

            {offering.status !== 'active' && (
              <p className="pricing-calculator__notice">
                This offering is not currently available for ordering.
              </p>
            )}
          </div>
        </aside>
      </div>

      <style>{offeringDetailPageStyles}</style>
    </div>
  );
}

interface SpecItemProps {
  icon: JSX.Element;
  label: string;
  value: string;
}

function SpecItem({ icon, label, value }: SpecItemProps): JSX.Element {
  return (
    <div className="spec-item">
      <span className="spec-item__icon" aria-hidden="true">{icon}</span>
      <div className="spec-item__content">
        <span className="spec-item__label">{label}</span>
        <span className="spec-item__value">{value}</span>
      </div>
    </div>
  );
}

interface PricingComponentRowProps {
  component: PriceComponent;
  denom: string;
}

function PricingComponentRow({ component, denom }: PricingComponentRowProps): JSX.Element {
  return (
    <div className="pricing-component">
      <span className="pricing-component__type">{formatResourceType(component.resourceType)}</span>
      <span className="pricing-component__price">
        {formatTokenAmount(component.price, 6, denom)} / {component.unit}
      </span>
    </div>
  );
}

interface StatusBadgeProps {
  status: string;
}

function StatusBadge({ status }: StatusBadgeProps): JSX.Element {
  const statusConfig: Record<string, { label: string; className: string }> = {
    active: { label: 'Active', className: 'status-badge--active' },
    paused: { label: 'Paused', className: 'status-badge--paused' },
    suspended: { label: 'Suspended', className: 'status-badge--suspended' },
    draft: { label: 'Draft', className: 'status-badge--draft' },
    pending_review: { label: 'Pending Review', className: 'status-badge--pending' },
    unlisted: { label: 'Unlisted', className: 'status-badge--unlisted' },
  };

  const config = statusConfig[status] || { label: status, className: '' };

  return (
    <span className={`status-badge ${config.className}`}>
      {config.label}
    </span>
  );
}

function LoadingSpinner(): JSX.Element {
  return (
    <svg className="loading-spinner" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10" opacity="0.25" />
      <path d="M12 2a10 10 0 0 1 10 10" />
    </svg>
  );
}

// Icons
function CpuIcon(): JSX.Element {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="4" y="4" width="16" height="16" rx="2" />
      <rect x="9" y="9" width="6" height="6" />
      <path d="M9 1v3M15 1v3M9 20v3M15 20v3M20 9h3M20 14h3M1 9h3M1 14h3" />
    </svg>
  );
}

function MemoryIcon(): JSX.Element {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M6 19v-2M10 19v-2M14 19v-2M18 19v-2M6 5v2M10 5v2M14 5v2M18 5v2" />
      <rect x="2" y="7" width="20" height="10" rx="1" />
    </svg>
  );
}

function StorageIcon(): JSX.Element {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <ellipse cx="12" cy="5" rx="9" ry="3" />
      <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
      <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
    </svg>
  );
}

function GpuIcon(): JSX.Element {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="6" width="20" height="12" rx="2" />
      <path d="M6 10v4M10 10v4M14 10v4M18 10v4" />
    </svg>
  );
}

function NetworkIcon(): JSX.Element {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M4 10h16M4 14h16M9 4L5 20M19 4l-4 16" />
    </svg>
  );
}

function LocationIcon(): JSX.Element {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
      <circle cx="12" cy="10" r="3" />
    </svg>
  );
}

function getOfferingTypeConfig(type: string): { label: string; color: string; bg: string } {
  const configs: Record<string, { label: string; color: string; bg: string }> = {
    compute: { label: 'Compute', color: '#7c3aed', bg: '#ede9fe' },
    storage: { label: 'Storage', color: '#0891b2', bg: '#cffafe' },
    gpu: { label: 'GPU', color: '#dc2626', bg: '#fee2e2' },
    kubernetes: { label: 'Kubernetes', color: '#2563eb', bg: '#dbeafe' },
    slurm: { label: 'SLURM', color: '#4f46e5', bg: '#e0e7ff' },
    custom: { label: 'Custom', color: '#6b7280', bg: '#f3f4f6' },
  };
  return configs[type] || { label: type, color: '#6b7280', bg: '#f3f4f6' };
}

function formatAttributeKey(key: string): string {
  return key
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (l) => l.toUpperCase());
}

function formatResourceType(type: string): string {
  const types: Record<string, string> = {
    cpu: 'CPU',
    ram: 'Memory',
    storage: 'Storage',
    gpu: 'GPU',
    network: 'Network',
  };
  return types[type] || type;
}

interface PricingResult {
  baseAmount: string;
  deposit: string;
  total: string;
}

function calculatePricing(offering: Offering, durationSeconds: number): PricingResult {
  const basePrice = parseInt(offering.pricing.basePrice || '0', 10);
  const deposit = parseInt(offering.pricing.depositRequired || '0', 10);

  // Calculate based on duration and price unit
  let baseAmount = 0;
  const hours = durationSeconds / 3600;

  switch (offering.pricing.unit) {
    case 'per_hour':
      baseAmount = basePrice * hours;
      break;
    case 'per_day':
      baseAmount = basePrice * (hours / 24);
      break;
    case 'per_month':
      baseAmount = basePrice * (hours / (24 * 30));
      break;
    default:
      baseAmount = basePrice * hours;
  }

  const total = baseAmount + deposit;

  return {
    baseAmount: String(Math.ceil(baseAmount)),
    deposit: String(deposit),
    total: String(Math.ceil(total)),
  };
}

const offeringDetailPageStyles = `
  .offering-detail-page {
    max-width: 1200px;
    margin: 0 auto;
    padding: 24px;
  }

  .offering-detail-page__back {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 0;
    background: transparent;
    border: none;
    color: #3b82f6;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    margin-bottom: 24px;
  }

  .offering-detail-page__back:hover {
    text-decoration: underline;
  }

  .offering-detail-page__layout {
    display: grid;
    grid-template-columns: 1fr 380px;
    gap: 24px;
    align-items: start;
  }

  .offering-detail-page__main {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .offering-detail-page__header {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 24px;
  }

  .offering-detail-page__badges {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
    flex-wrap: wrap;
  }

  .offering-detail-page__type {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .offering-detail-page__tee {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    background: #dcfce7;
    color: #16a34a;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .status-badge {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .status-badge--active {
    background: #dcfce7;
    color: #16a34a;
  }

  .status-badge--paused {
    background: #fef3c7;
    color: #d97706;
  }

  .status-badge--suspended {
    background: #fee2e2;
    color: #dc2626;
  }

  .status-badge--draft,
  .status-badge--pending,
  .status-badge--unlisted {
    background: #f3f4f6;
    color: #6b7280;
  }

  .offering-detail-page__title {
    margin: 0 0 8px;
    font-size: 1.75rem;
    font-weight: 700;
    color: #111827;
  }

  .offering-detail-page__description {
    margin: 0;
    font-size: 1rem;
    color: #4b5563;
    line-height: 1.6;
  }

  .offering-detail-page__section {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 24px;
  }

  .offering-detail-page__section-title {
    margin: 0 0 16px;
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
  }

  .offering-detail-page__specs {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }

  .spec-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px;
    background: #f9fafb;
    border-radius: 8px;
  }

  .spec-item__icon {
    color: #6b7280;
    flex-shrink: 0;
  }

  .spec-item__content {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .spec-item__label {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .spec-item__value {
    font-size: 0.875rem;
    font-weight: 600;
    color: #111827;
  }

  .offering-detail-page__attributes {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #f3f4f6;
  }

  .offering-detail-page__attributes-title {
    margin: 0 0 12px;
    font-size: 0.875rem;
    font-weight: 600;
    color: #374151;
  }

  .offering-detail-page__attributes-list {
    margin: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 8px 16px;
  }

  .offering-detail-page__attribute {
    display: flex;
    gap: 8px;
    font-size: 0.875rem;
  }

  .offering-detail-page__attribute dt {
    color: #6b7280;
  }

  .offering-detail-page__attribute dd {
    margin: 0;
    font-weight: 500;
    color: #111827;
  }

  .offering-detail-page__pricing-components {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .pricing-component {
    display: flex;
    justify-content: space-between;
    padding: 12px;
    background: #f9fafb;
    border-radius: 8px;
  }

  .pricing-component__type {
    font-weight: 500;
    color: #374151;
  }

  .pricing-component__price {
    font-weight: 600;
    color: #111827;
  }

  .offering-detail-page__requirements {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .offering-detail-page__requirement {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px;
    background: #f9fafb;
    border-radius: 8px;
  }

  .offering-detail-page__requirement-label {
    font-size: 0.875rem;
    color: #6b7280;
  }

  .offering-detail-page__requirement-value {
    font-size: 0.875rem;
    font-weight: 500;
    color: #111827;
  }

  .offering-detail-page__scopes {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .offering-detail-page__scope {
    padding: 4px 8px;
    background: #eff6ff;
    color: #3b82f6;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .offering-detail-page__sidebar {
    display: flex;
    flex-direction: column;
    gap: 16px;
    position: sticky;
    top: 24px;
  }

  /* Pricing Calculator */
  .pricing-calculator {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
  }

  .pricing-calculator__title {
    margin: 0 0 16px;
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
  }

  .pricing-calculator__field {
    margin-bottom: 16px;
  }

  .pricing-calculator__label {
    display: block;
    margin-bottom: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .pricing-calculator__duration {
    display: flex;
    gap: 8px;
  }

  .pricing-calculator__input {
    flex: 1;
    padding: 10px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.875rem;
    color: #111827;
  }

  .pricing-calculator__input:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .pricing-calculator__select {
    padding: 10px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.875rem;
    color: #111827;
    background: white;
    cursor: pointer;
  }

  .pricing-calculator__select:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .pricing-calculator__duration-info {
    display: flex;
    justify-content: space-between;
    margin-top: 6px;
    font-size: 0.75rem;
    color: #6b7280;
  }

  .pricing-calculator__error {
    margin-top: 6px;
    padding: 8px;
    background: #fee2e2;
    color: #dc2626;
    border-radius: 6px;
    font-size: 0.75rem;
  }

  .pricing-calculator__breakdown {
    margin-bottom: 16px;
    padding: 16px;
    background: #f9fafb;
    border-radius: 8px;
  }

  .pricing-calculator__row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    font-size: 0.875rem;
    color: #6b7280;
    border-bottom: 1px solid #e5e7eb;
  }

  .pricing-calculator__row:last-child {
    border-bottom: none;
  }

  .pricing-calculator__row--total {
    padding-top: 12px;
    margin-top: 4px;
    border-top: 2px solid #e5e7eb;
    border-bottom: none;
    font-weight: 600;
    color: #111827;
    font-size: 1rem;
  }

  .pricing-calculator__checkout {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px 20px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .pricing-calculator__checkout:hover:not(:disabled) {
    background: #2563eb;
  }

  .pricing-calculator__checkout:disabled {
    background: #9ca3af;
    cursor: not-allowed;
  }

  .pricing-calculator__notice {
    margin: 12px 0 0;
    padding: 12px;
    background: #fef3c7;
    color: #92400e;
    border-radius: 8px;
    font-size: 0.875rem;
    text-align: center;
  }

  .loading-spinner {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  @media (prefers-reduced-motion: reduce) {
    .pricing-calculator__checkout {
      transition: none;
    }
    .loading-spinner {
      animation: none;
    }
  }

  @media (max-width: 1024px) {
    .offering-detail-page__layout {
      grid-template-columns: 1fr;
    }

    .offering-detail-page__sidebar {
      position: static;
    }
  }

  @media (max-width: 640px) {
    .offering-detail-page {
      padding: 16px;
    }

    .offering-detail-page__specs {
      grid-template-columns: 1fr;
    }
  }
`;
