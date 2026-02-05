/**
 * Marketplace Offering Card Component
 * VE-703: Display marketplace offering in browse grid
 */
import * as React from "react";
import { formatTokenAmount } from "../../../../utils/format";
import type { Offering } from "../../../../types/marketplace";

export interface MarketplaceOfferingCardProps {
  offering: Offering;
  onSelect?: (offering: Offering) => void;
  isSelected?: boolean;
  className?: string;
}

function getOfferingTypeConfig(type: string): {
  label: string;
  color: string;
  bg: string;
} {
  const configs: Record<string, { label: string; color: string; bg: string }> =
    {
      compute: { label: "Compute", color: "#7c3aed", bg: "#ede9fe" },
      storage: { label: "Storage", color: "#0891b2", bg: "#cffafe" },
      gpu: { label: "GPU", color: "#dc2626", bg: "#fee2e2" },
      kubernetes: { label: "Kubernetes", color: "#2563eb", bg: "#dbeafe" },
      slurm: { label: "SLURM", color: "#4f46e5", bg: "#e0e7ff" },
      custom: { label: "Custom", color: "#6b7280", bg: "#f3f4f6" },
    };
  return configs[type] || { label: type, color: "#6b7280", bg: "#f3f4f6" };
}

export function MarketplaceOfferingCard({
  offering,
  onSelect,
  isSelected = false,
  className = "",
}: MarketplaceOfferingCardProps): JSX.Element {
  const typeConfig = getOfferingTypeConfig(offering.type);

  const handleClick = () => {
    if (onSelect) {
      onSelect(offering);
    }
  };

  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (onSelect && (event.key === "Enter" || event.key === " ")) {
      event.preventDefault();
      onSelect(offering);
    }
  };

  const cardId = `offering-${offering.id}`;
  const titleId = `${cardId}-title`;
  const descId = `${cardId}-desc`;
  const priceId = `${cardId}-price`;

  const availabilityText =
    offering.status === "active"
      ? "Available"
      : offering.status === "paused"
        ? "Limited availability"
        : "Unavailable";

  const availabilityClass =
    offering.status === "active"
      ? "available"
      : offering.status === "paused"
        ? "limited"
        : "unavailable";

  return (
    <div
      className={`mkt-offering-card ${isSelected ? "mkt-offering-card--selected" : ""} ${className}`}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role={onSelect ? "button" : "article"}
      tabIndex={onSelect ? 0 : undefined}
      aria-pressed={onSelect ? isSelected : undefined}
      aria-labelledby={titleId}
      aria-describedby={`${descId} ${priceId}`}
    >
      {/* Header */}
      <div className="mkt-offering-card__header">
        <span
          className="mkt-offering-card__type"
          style={{ color: typeConfig.color, backgroundColor: typeConfig.bg }}
          aria-label={`Type: ${typeConfig.label}`}
        >
          {typeConfig.label}
        </span>
        {offering.hasEncryptedDetails && (
          <span
            className="mkt-offering-card__tee"
            aria-label="Trusted Execution Environment enabled"
          >
            <span aria-hidden="true">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0 1 10 0v4" />
              </svg>
            </span>{" "}
            TEE
          </span>
        )}
      </div>

      {/* Title & Provider */}
      <h3 className="mkt-offering-card__title" id={titleId}>
        {offering.title}
      </h3>
      <p className="mkt-offering-card__provider">by {offering.providerName}</p>

      {/* Description */}
      <p className="mkt-offering-card__description" id={descId}>
        {offering.description}
      </p>

      {/* Specs */}
      <dl className="mkt-offering-card__specs" aria-label="Specifications">
        <div className="mkt-offering-card__spec">
          <dt className="mkt-offering-card__spec-label">CPU</dt>
          <dd className="mkt-offering-card__spec-value">
            {offering.resources.cpuCores} cores
          </dd>
        </div>
        <div className="mkt-offering-card__spec">
          <dt className="mkt-offering-card__spec-label">Memory</dt>
          <dd className="mkt-offering-card__spec-value">
            {offering.resources.memoryGB} GB
          </dd>
        </div>
        <div className="mkt-offering-card__spec">
          <dt className="mkt-offering-card__spec-label">Storage</dt>
          <dd className="mkt-offering-card__spec-value">
            {offering.resources.storageGB} GB
          </dd>
        </div>
        {offering.resources.gpuCount && (
          <div className="mkt-offering-card__spec">
            <dt className="mkt-offering-card__spec-label">GPU</dt>
            <dd className="mkt-offering-card__spec-value">
              {offering.resources.gpuCount}x{" "}
              {offering.resources.gpuModel || "GPU"}
            </dd>
          </div>
        )}
      </dl>

      {/* Price */}
      <div className="mkt-offering-card__price" id={priceId}>
        <span className="mkt-offering-card__price-amount">
          {formatTokenAmount(
            offering.pricing.basePrice || "0",
            6,
            offering.pricing.denom,
          )}
        </span>
        <span className="mkt-offering-card__price-unit">
          / {formatPriceUnit(offering.pricing.unit)}
        </span>
      </div>

      {/* Availability & Reliability */}
      <div className="mkt-offering-card__footer">
        <span
          className={`mkt-offering-card__status mkt-offering-card__status--${availabilityClass}`}
          role="status"
          aria-label={availabilityText}
        >
          <span aria-hidden="true">
            {offering.status === "active"
              ? "●"
              : offering.status === "paused"
                ? "◐"
                : "○"}
          </span>{" "}
          {availabilityText}
        </span>
        <span
          className="mkt-offering-card__reliability"
          title="Reliability score"
        >
          <span aria-hidden="true">★</span> {offering.reliabilityScore}%
        </span>
      </div>

      <style>{cardStyles}</style>
    </div>
  );
}

function formatPriceUnit(unit?: string): string {
  switch (unit) {
    case "per_hour":
      return "hour";
    case "per_day":
      return "day";
    case "per_month":
      return "month";
    case "per_cpu_hour":
      return "CPU-hour";
    case "per_gpu_hour":
      return "GPU-hour";
    case "per_gb_hour":
      return "GB-hour";
    default:
      return "hour";
  }
}

const cardStyles = `
  .mkt-offering-card {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 20px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .mkt-offering-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .mkt-offering-card:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
    border-color: #3b82f6;
  }

  .mkt-offering-card--selected {
    border-color: #3b82f6;
    background: #f8fafc;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
  }

  @media (prefers-reduced-motion: reduce) {
    .mkt-offering-card {
      transition: none;
    }
    .mkt-offering-card:hover {
      transform: none;
    }
  }

  .mkt-offering-card__header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .mkt-offering-card__type {
    padding: 4px 10px;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .mkt-offering-card__tee {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    font-size: 0.75rem;
    color: #16a34a;
  }

  .mkt-offering-card__title {
    font-size: 1.125rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 4px;
  }

  .mkt-offering-card__provider {
    font-size: 0.75rem;
    color: #6b7280;
    margin: 0 0 12px;
  }

  .mkt-offering-card__description {
    font-size: 0.875rem;
    color: #374151;
    margin: 0 0 16px;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .mkt-offering-card__specs {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    margin: 0 0 16px;
  }

  .mkt-offering-card__spec {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .mkt-offering-card__spec-label {
    font-size: 0.625rem;
    color: #9ca3af;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .mkt-offering-card__spec-value {
    font-size: 0.875rem;
    color: #111827;
    font-weight: 500;
    margin: 0;
  }

  .mkt-offering-card__price {
    display: flex;
    align-items: baseline;
    gap: 4px;
    margin-bottom: 12px;
  }

  .mkt-offering-card__price-amount {
    font-size: 1.25rem;
    font-weight: 700;
    color: #111827;
  }

  .mkt-offering-card__price-unit {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .mkt-offering-card__footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-top: 12px;
    border-top: 1px solid #f3f4f6;
  }

  .mkt-offering-card__status {
    font-size: 0.75rem;
    font-weight: 500;
  }

  .mkt-offering-card__status--available {
    color: #16a34a;
  }

  .mkt-offering-card__status--limited {
    color: #ca8a04;
  }

  .mkt-offering-card__status--unavailable {
    color: #dc2626;
  }

  .mkt-offering-card__reliability {
    font-size: 0.75rem;
    font-weight: 500;
    color: #f59e0b;
  }
`;
