/**
 * Featured Offerings Section
 * VE Portal Landing Page
 */
import * as React from "react";
import { useChain } from "../../../hooks/useChain";
import type { Offering } from "../../../types/marketplace";
import { ProviderBadge } from "../../pages/marketplace/components/ProviderInfo";

export interface FeaturedOfferingsProps {
  limit?: number;
  className?: string;
  onOrder?: (offering: Offering) => void;
  fallbackOfferings?: Offering[];
}

export function FeaturedOfferings({
  limit = 6,
  className = "",
  onOrder,
  fallbackOfferings = defaultOfferings,
}: FeaturedOfferingsProps): JSX.Element {
  const { queryClient } = useChain();
  const [offerings, setOfferings] =
    React.useState<Offering[]>(fallbackOfferings);
  const [isLoading, setIsLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    let mounted = true;
    const fetchOfferings = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const result = await queryClient.query<{ offerings?: Offering[] }>(
          "/marketplace/offerings",
          {
            limit: String(limit),
            sort_field: "reliability_score",
            sort_direction: "desc",
          },
        );

        if (mounted && result.offerings?.length) {
          setOfferings(result.offerings.slice(0, limit));
        }
      } catch (err) {
        if (mounted) {
          setError(
            "Unable to load offerings. Showing curated highlights instead.",
          );
          setOfferings(fallbackOfferings);
        }
      } finally {
        if (mounted) setIsLoading(false);
      }
    };

    fetchOfferings();
    return () => {
      mounted = false;
    };
  }, [queryClient, limit, fallbackOfferings]);

  return (
    <section
      className={`ve-offerings ${className}`}
      aria-labelledby="ve-offerings-title"
    >
      <div className="ve-offerings__header">
        <div>
          <h2 id="ve-offerings-title">Featured offerings</h2>
          <p>
            Top-performing providers with instant provisioning and transparent
            pricing.
          </p>
        </div>
        <a className="ve-offerings__link" href="/marketplace">
          View all listings
        </a>
      </div>

      <div className="ve-offerings__grid" role="list">
        {offerings.slice(0, limit).map((offering) => (
          <OfferingCard
            key={offering.id}
            offering={offering}
            onOrder={onOrder}
            isLoading={isLoading}
          />
        ))}
      </div>

      {error && (
        <div className="ve-offerings__error" role="status">
          {error}
        </div>
      )}

      <style>{offeringsStyles}</style>
    </section>
  );
}

interface OfferingCardProps {
  offering: Offering;
  onOrder?: (offering: Offering) => void;
  isLoading?: boolean;
}

function OfferingCard({
  offering,
  onOrder,
  isLoading,
}: OfferingCardProps): JSX.Element {
  const price = resolvePrice(offering);

  return (
    <article className="ve-offering-card" role="listitem">
      <div className="ve-offering-card__header">
        <span
          className={`ve-offering-card__type ve-offering-card__type--${offering.type}`}
        >
          {formatType(offering.type)}
        </span>
        <span
          className={`ve-offering-card__status ve-offering-card__status--${offering.status}`}
        >
          {formatStatus(offering.status)}
        </span>
      </div>

      <h3 className="ve-offering-card__title">{offering.title}</h3>
      <p className="ve-offering-card__description">{offering.description}</p>

      <ProviderBadge
        providerName={offering.providerName}
        providerAddress={offering.providerAddress}
        reliabilityScore={offering.reliabilityScore}
        isVerified
        className="ve-offering-card__provider"
      />

      <div className="ve-offering-card__specs">
        <SpecItem label="CPU" value={`${offering.resources.cpuCores} vCPU`} />
        <SpecItem label="Memory" value={`${offering.resources.memoryGB} GB`} />
        <SpecItem
          label="Storage"
          value={`${offering.resources.storageGB} GB`}
        />
        {offering.resources.gpuCount ? (
          <SpecItem
            label="GPU"
            value={`${offering.resources.gpuCount}x ${offering.resources.gpuModel || "GPU"}`}
          />
        ) : (
          <SpecItem label="GPU" value="Optional" />
        )}
      </div>

      <div className="ve-offering-card__footer">
        <div>
          <span className="ve-offering-card__price-label">From</span>
          <div className="ve-offering-card__price">
            {isLoading ? "--" : price}
          </div>
        </div>
        <button
          type="button"
          className="ve-offering-card__action"
          onClick={() => onOrder?.(offering)}
        >
          Quick order
        </button>
      </div>
    </article>
  );
}

function SpecItem({
  label,
  value,
}: {
  label: string;
  value: string;
}): JSX.Element {
  return (
    <div className="ve-offering-card__spec">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatType(type: Offering["type"]): string {
  const map: Record<Offering["type"], string> = {
    compute: "Compute",
    gpu: "GPU",
    storage: "Storage",
    kubernetes: "Kubernetes",
    slurm: "HPC",
    custom: "Custom",
  };
  return map[type] || "Compute";
}

function formatStatus(status: Offering["status"]): string {
  const map: Record<Offering["status"], string> = {
    active: "Active",
    paused: "Paused",
    unlisted: "Unlisted",
    suspended: "Suspended",
    draft: "Draft",
    pending_review: "Review",
  };
  return map[status] || "Active";
}

function resolvePrice(offering: Offering): string {
  const price = offering.pricing.basePrice
    ? parseFloat(offering.pricing.basePrice)
    : offering.pricing.components?.length
      ? parseFloat(offering.pricing.components[0].price)
      : 0;

  const unit = offering.pricing.unit || "per_hour";
  const formatted = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 2,
  }).format(price || 0.18);

  return `${formatted}/${unit.replace("per_", "")}`;
}

const defaultOfferings: Offering[] = [
  {
    id: "ve-offer-001",
    providerAddress: "ve1alpha",
    providerName: "Nimbus Array",
    title: "Elastic GPU Cloud",
    description:
      "H100 and A100 clusters tuned for AI training with private fabric connectivity.",
    type: "gpu",
    status: "active",
    region: "North America",
    resources: {
      cpuCores: 64,
      memoryGB: 512,
      storageGB: 2000,
      gpuCount: 8,
      gpuModel: "NVIDIA H100",
      bandwidthGbps: 100,
      attributes: {},
    },
    pricing: {
      basePrice: "2.75",
      unit: "per_hour",
      denom: "uvirt",
      depositRequired: "0",
      minDurationSeconds: 3600,
      maxDurationSeconds: 604800,
    },
    identityRequirements: {
      minScore: 72,
      requiredScopes: ["basic", "enterprise"],
      mfaRequired: true,
    },
    reliabilityScore: 96,
    benchmarkSummary: {
      cpuScore: 92,
      memoryScore: 95,
      storageScore: 90,
      networkScore: 94,
      gpuScore: 98,
      overallScore: 95,
      lastBenchmarkAt: Date.now(),
      suiteVersion: "1.9.0",
    },
    createdAt: Date.now(),
    updatedAt: Date.now(),
    hasEncryptedDetails: true,
  },
  {
    id: "ve-offer-002",
    providerAddress: "ve1beta",
    providerName: "Aria Fabric",
    title: "Latency-Optimized Compute",
    description:
      "Bare metal pools with SLA-backed latency and transparent hardware lineage.",
    type: "compute",
    status: "active",
    region: "Europe",
    resources: {
      cpuCores: 32,
      memoryGB: 256,
      storageGB: 1000,
      bandwidthGbps: 50,
      attributes: {},
    },
    pricing: {
      basePrice: "0.68",
      unit: "per_hour",
      denom: "uvirt",
      depositRequired: "0",
      minDurationSeconds: 3600,
    },
    identityRequirements: {
      minScore: 48,
      requiredScopes: ["basic"],
      mfaRequired: false,
    },
    reliabilityScore: 92,
    benchmarkSummary: {
      cpuScore: 89,
      memoryScore: 91,
      storageScore: 88,
      networkScore: 93,
      overallScore: 90,
      lastBenchmarkAt: Date.now(),
      suiteVersion: "1.9.0",
    },
    createdAt: Date.now(),
    updatedAt: Date.now(),
    hasEncryptedDetails: false,
  },
  {
    id: "ve-offer-003",
    providerAddress: "ve1gamma",
    providerName: "Solstice Data",
    title: "Secure Kubernetes Mesh",
    description:
      "Managed clusters with confidential nodes and policy-ready compliance controls.",
    type: "kubernetes",
    status: "active",
    region: "Asia Pacific",
    resources: {
      cpuCores: 24,
      memoryGB: 192,
      storageGB: 800,
      bandwidthGbps: 40,
      attributes: {},
    },
    pricing: {
      basePrice: "0.42",
      unit: "per_hour",
      denom: "uvirt",
      depositRequired: "0",
      minDurationSeconds: 3600,
    },
    identityRequirements: {
      minScore: 60,
      requiredScopes: ["basic", "compliance"],
      mfaRequired: true,
    },
    reliabilityScore: 89,
    benchmarkSummary: {
      cpuScore: 86,
      memoryScore: 88,
      storageScore: 90,
      networkScore: 87,
      overallScore: 88,
      lastBenchmarkAt: Date.now(),
      suiteVersion: "1.9.0",
    },
    createdAt: Date.now(),
    updatedAt: Date.now(),
    hasEncryptedDetails: true,
  },
];

const offeringsStyles = `
  .ve-offerings {
    padding: 96px 0;
    background: #0f172a;
    color: #f8fafc;
  }

  .ve-offerings__header {
    width: min(1120px, 90vw);
    margin: 0 auto 36px;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 16px;
  }

  .ve-offerings__header h2 {
    font-family: "Space Grotesk", "Manrope", "Segoe UI", sans-serif;
    font-size: clamp(1.9rem, 2.4vw, 2.6rem);
    margin: 0 0 8px;
  }

  .ve-offerings__header p {
    margin: 0;
    color: rgba(226, 232, 240, 0.7);
    max-width: 460px;
  }

  .ve-offerings__link {
    color: #38bdf8;
    text-decoration: none;
    font-weight: 600;
  }

  .ve-offerings__grid {
    width: min(1120px, 90vw);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 20px;
  }

  .ve-offering-card {
    background: rgba(15, 23, 42, 0.9);
    border: 1px solid rgba(148, 163, 184, 0.2);
    border-radius: 20px;
    padding: 22px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    box-shadow: 0 24px 50px rgba(2, 6, 23, 0.45);
  }

  .ve-offering-card__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.12em;
  }

  .ve-offering-card__type {
    padding: 4px 10px;
    border-radius: 999px;
    background: rgba(56, 189, 248, 0.18);
    color: #7dd3fc;
    font-weight: 600;
  }

  .ve-offering-card__status {
    color: #cbd5f5;
  }

  .ve-offering-card__title {
    margin: 0;
    font-size: 1.2rem;
  }

  .ve-offering-card__description {
    margin: 0;
    color: rgba(226, 232, 240, 0.7);
    min-height: 48px;
  }

  .ve-offering-card__provider {
    margin-top: 4px;
  }

  .ve-offering-card__specs {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .ve-offering-card__spec {
    background: rgba(30, 41, 59, 0.6);
    border-radius: 12px;
    padding: 10px 12px;
    font-size: 0.8rem;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .ve-offering-card__spec span {
    color: rgba(226, 232, 240, 0.6);
  }

  .ve-offering-card__spec strong {
    font-size: 0.95rem;
  }

  .ve-offering-card__footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .ve-offering-card__price-label {
    font-size: 0.75rem;
    color: rgba(226, 232, 240, 0.6);
    text-transform: uppercase;
    letter-spacing: 0.18em;
  }

  .ve-offering-card__price {
    font-size: 1.15rem;
    font-weight: 700;
  }

  .ve-offering-card__action {
    background: #38bdf8;
    color: #0f172a;
    border: none;
    border-radius: 999px;
    padding: 10px 16px;
    font-weight: 600;
    cursor: pointer;
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .ve-offering-card__action:hover {
    transform: translateY(-1px);
    box-shadow: 0 12px 24px rgba(56, 189, 248, 0.4);
  }

  .ve-offerings__error {
    width: min(1120px, 90vw);
    margin: 16px auto 0;
    color: rgba(148, 163, 184, 0.9);
    font-size: 0.85rem;
  }

  @media (max-width: 860px) {
    .ve-offerings__header {
      flex-direction: column;
      align-items: flex-start;
    }

    .ve-offering-card__footer {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
  }
`;
