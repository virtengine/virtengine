/**
 * Filter Panel Component
 * VE-703: Marketplace offering filter controls
 */
import * as React from "react";
import { useState, useCallback } from "react";
import type {
  OfferingFilter,
  OfferingType,
} from "../../../../types/marketplace";
import { OFFERING_CATEGORIES, REGIONS } from "../hooks/useOfferings";

export interface FilterPanelProps {
  filter: OfferingFilter;
  onChange: (filter: OfferingFilter) => void;
  onApply: () => void;
  onReset: () => void;
  isCollapsible?: boolean;
  className?: string;
}

export function FilterPanel({
  filter,
  onChange,
  onApply,
  onReset,
  isCollapsible = true,
  className = "",
}: FilterPanelProps): JSX.Element {
  const [expandedSections, setExpandedSections] = useState<Set<string>>(
    new Set(["type", "region", "resources"]),
  );

  const toggleSection = useCallback((section: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      if (next.has(section)) {
        next.delete(section);
      } else {
        next.add(section);
      }
      return next;
    });
  }, []);

  const handleTypeChange = useCallback(
    (type: OfferingType, checked: boolean) => {
      const types = filter.types || [];
      const newTypes = checked
        ? [...types, type]
        : types.filter((t) => t !== type);
      onChange({
        ...filter,
        types: newTypes.length > 0 ? newTypes : undefined,
      });
    },
    [filter, onChange],
  );

  const handleRegionChange = useCallback(
    (region: string, checked: boolean) => {
      const regions = filter.regions || [];
      const newRegions = checked
        ? [...regions, region]
        : regions.filter((r) => r !== region);
      onChange({
        ...filter,
        regions: newRegions.length > 0 ? newRegions : undefined,
      });
    },
    [filter, onChange],
  );

  const handleResourceChange = useCallback(
    (field: keyof OfferingFilter, value: string) => {
      const numValue = value ? parseInt(value, 10) : undefined;
      onChange({ ...filter, [field]: numValue });
    },
    [filter, onChange],
  );

  const handlePriceChange = useCallback(
    (value: string) => {
      onChange({ ...filter, maxPricePerHour: value || undefined });
    },
    [filter, onChange],
  );

  const handleReliabilityChange = useCallback(
    (value: string) => {
      const numValue = value ? parseInt(value, 10) : undefined;
      onChange({ ...filter, minReliabilityScore: numValue });
    },
    [filter, onChange],
  );

  const handleGpuToggle = useCallback(
    (checked: boolean) => {
      onChange({ ...filter, requireGpu: checked || undefined });
    },
    [filter, onChange],
  );

  const handleEligibleToggle = useCallback(
    (checked: boolean) => {
      onChange({ ...filter, onlyEligible: checked || undefined });
    },
    [filter, onChange],
  );

  const activeFilterCount = [
    filter.types?.length,
    filter.regions?.length,
    filter.minCpuCores,
    filter.minMemoryGB,
    filter.minStorageGB,
    filter.requireGpu,
    filter.maxPricePerHour,
    filter.minReliabilityScore,
    filter.onlyEligible,
  ].filter(Boolean).length;

  const offeringTypes: { type: OfferingType; label: string }[] = [
    { type: "compute", label: "Compute" },
    { type: "gpu", label: "GPU" },
    { type: "storage", label: "Storage" },
    { type: "kubernetes", label: "Kubernetes" },
    { type: "slurm", label: "SLURM/HPC" },
    { type: "custom", label: "Custom" },
  ];

  return (
    <aside
      className={`filter-panel ${className}`}
      aria-label="Filter offerings"
    >
      <div className="filter-panel__header">
        <h2 className="filter-panel__title">
          Filters
          {activeFilterCount > 0 && (
            <span
              className="filter-panel__count"
              aria-label={`${activeFilterCount} active filters`}
            >
              {activeFilterCount}
            </span>
          )}
        </h2>
        {activeFilterCount > 0 && (
          <button
            type="button"
            className="filter-panel__reset"
            onClick={onReset}
            aria-label="Reset all filters"
          >
            Reset
          </button>
        )}
      </div>

      {/* Type Filter */}
      <FilterSection
        id="type"
        title="Offering Type"
        isExpanded={expandedSections.has("type")}
        onToggle={() => toggleSection("type")}
        isCollapsible={isCollapsible}
      >
        <div className="filter-panel__checkboxes">
          {offeringTypes.map(({ type, label }) => (
            <label key={type} className="filter-panel__checkbox">
              <input
                type="checkbox"
                checked={filter.types?.includes(type) || false}
                onChange={(e) => handleTypeChange(type, e.target.checked)}
              />
              <span className="filter-panel__checkbox-label">{label}</span>
            </label>
          ))}
        </div>
      </FilterSection>

      {/* Region Filter */}
      <FilterSection
        id="region"
        title="Region"
        isExpanded={expandedSections.has("region")}
        onToggle={() => toggleSection("region")}
        isCollapsible={isCollapsible}
      >
        <div className="filter-panel__checkboxes">
          {REGIONS.map((region) => (
            <label key={region.id} className="filter-panel__checkbox">
              <input
                type="checkbox"
                checked={filter.regions?.includes(region.id) || false}
                onChange={(e) =>
                  handleRegionChange(region.id, e.target.checked)
                }
              />
              <span className="filter-panel__checkbox-label">
                <span aria-hidden="true">{region.flag}</span> {region.name}
              </span>
            </label>
          ))}
        </div>
      </FilterSection>

      {/* Resource Requirements */}
      <FilterSection
        id="resources"
        title="Resource Requirements"
        isExpanded={expandedSections.has("resources")}
        onToggle={() => toggleSection("resources")}
        isCollapsible={isCollapsible}
      >
        <div className="filter-panel__fields">
          <div className="filter-panel__field">
            <label htmlFor="filter-cpu" className="filter-panel__label">
              Min CPU Cores
            </label>
            <input
              id="filter-cpu"
              type="number"
              min="0"
              step="1"
              className="filter-panel__input"
              value={filter.minCpuCores || ""}
              onChange={(e) =>
                handleResourceChange("minCpuCores", e.target.value)
              }
              placeholder="Any"
            />
          </div>

          <div className="filter-panel__field">
            <label htmlFor="filter-memory" className="filter-panel__label">
              Min Memory (GB)
            </label>
            <input
              id="filter-memory"
              type="number"
              min="0"
              step="1"
              className="filter-panel__input"
              value={filter.minMemoryGB || ""}
              onChange={(e) =>
                handleResourceChange("minMemoryGB", e.target.value)
              }
              placeholder="Any"
            />
          </div>

          <div className="filter-panel__field">
            <label htmlFor="filter-storage" className="filter-panel__label">
              Min Storage (GB)
            </label>
            <input
              id="filter-storage"
              type="number"
              min="0"
              step="1"
              className="filter-panel__input"
              value={filter.minStorageGB || ""}
              onChange={(e) =>
                handleResourceChange("minStorageGB", e.target.value)
              }
              placeholder="Any"
            />
          </div>

          <label className="filter-panel__checkbox filter-panel__checkbox--standalone">
            <input
              type="checkbox"
              checked={filter.requireGpu || false}
              onChange={(e) => handleGpuToggle(e.target.checked)}
            />
            <span className="filter-panel__checkbox-label">Requires GPU</span>
          </label>
        </div>
      </FilterSection>

      {/* Price Filter */}
      <FilterSection
        id="price"
        title="Price"
        isExpanded={expandedSections.has("price")}
        onToggle={() => toggleSection("price")}
        isCollapsible={isCollapsible}
      >
        <div className="filter-panel__fields">
          <div className="filter-panel__field">
            <label htmlFor="filter-max-price" className="filter-panel__label">
              Max Price per Hour (VE)
            </label>
            <input
              id="filter-max-price"
              type="text"
              className="filter-panel__input"
              value={filter.maxPricePerHour || ""}
              onChange={(e) => handlePriceChange(e.target.value)}
              placeholder="No limit"
            />
          </div>
        </div>
      </FilterSection>

      {/* Reliability Filter */}
      <FilterSection
        id="reliability"
        title="Provider Quality"
        isExpanded={expandedSections.has("reliability")}
        onToggle={() => toggleSection("reliability")}
        isCollapsible={isCollapsible}
      >
        <div className="filter-panel__fields">
          <div className="filter-panel__field">
            <label htmlFor="filter-reliability" className="filter-panel__label">
              Min Reliability Score
            </label>
            <input
              id="filter-reliability"
              type="range"
              min="0"
              max="100"
              step="10"
              className="filter-panel__range"
              value={filter.minReliabilityScore || 0}
              onChange={(e) => handleReliabilityChange(e.target.value)}
            />
            <div className="filter-panel__range-labels">
              <span>Any</span>
              <span>{filter.minReliabilityScore || 0}%</span>
            </div>
          </div>

          <label className="filter-panel__checkbox filter-panel__checkbox--standalone">
            <input
              type="checkbox"
              checked={filter.onlyEligible || false}
              onChange={(e) => handleEligibleToggle(e.target.checked)}
            />
            <span className="filter-panel__checkbox-label">
              Only show offerings I can order
            </span>
          </label>
        </div>
      </FilterSection>

      {/* Apply Button */}
      <div className="filter-panel__actions">
        <button type="button" className="filter-panel__apply" onClick={onApply}>
          Apply Filters
        </button>
      </div>

      <style>{filterPanelStyles}</style>
    </aside>
  );
}

interface FilterSectionProps {
  id: string;
  title: string;
  isExpanded: boolean;
  onToggle: () => void;
  isCollapsible: boolean;
  children: React.ReactNode;
}

function FilterSection({
  id,
  title,
  isExpanded,
  onToggle,
  isCollapsible,
  children,
}: FilterSectionProps): JSX.Element {
  const contentId = `filter-section-${id}-content`;

  return (
    <div className="filter-section">
      {isCollapsible ? (
        <button
          type="button"
          className="filter-section__header"
          onClick={onToggle}
          aria-expanded={isExpanded}
          aria-controls={contentId}
        >
          <span className="filter-section__title">{title}</span>
          <span className="filter-section__icon" aria-hidden="true">
            {isExpanded ? "−" : "+"}
          </span>
        </button>
      ) : (
        <div className="filter-section__header filter-section__header--static">
          <span className="filter-section__title">{title}</span>
        </div>
      )}

      {(!isCollapsible || isExpanded) && (
        <div id={contentId} className="filter-section__content">
          {children}
        </div>
      )}
    </div>
  );
}

const filterPanelStyles = `
  .filter-panel {
    background: white;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    padding: 16px;
    width: 280px;
    flex-shrink: 0;
  }

  .filter-panel__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid #e5e7eb;
  }

  .filter-panel__title {
    font-size: 1rem;
    font-weight: 600;
    color: #111827;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .filter-panel__count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 20px;
    height: 20px;
    padding: 0 6px;
    background: #3b82f6;
    color: white;
    border-radius: 10px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .filter-panel__reset {
    padding: 4px 8px;
    background: transparent;
    border: none;
    color: #3b82f6;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    border-radius: 4px;
    transition: background-color 0.2s;
  }

  .filter-panel__reset:hover {
    background: #eff6ff;
  }

  .filter-section {
    margin-bottom: 12px;
  }

  .filter-section__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 8px 0;
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: left;
  }

  .filter-section__header--static {
    cursor: default;
  }

  .filter-section__title {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
  }

  .filter-section__icon {
    color: #9ca3af;
    font-size: 1rem;
    font-weight: 500;
  }

  .filter-section__content {
    padding-top: 8px;
  }

  .filter-panel__checkboxes {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .filter-panel__checkbox {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  .filter-panel__checkbox input[type="checkbox"] {
    width: 16px;
    height: 16px;
    accent-color: #3b82f6;
    cursor: pointer;
  }

  .filter-panel__checkbox-label {
    font-size: 0.875rem;
    color: #374151;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .filter-panel__checkbox--standalone {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f3f4f6;
  }

  .filter-panel__fields {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .filter-panel__field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .filter-panel__label {
    font-size: 0.75rem;
    font-weight: 500;
    color: #6b7280;
    text-transform: uppercase;
    letter-spacing: 0.025em;
  }

  .filter-panel__input {
    padding: 8px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 0.875rem;
    color: #111827;
    transition: border-color 0.2s;
  }

  .filter-panel__input:focus {
    outline: none;
    border-color: #3b82f6;
  }

  .filter-panel__input::placeholder {
    color: #9ca3af;
  }

  .filter-panel__range {
    width: 100%;
    accent-color: #3b82f6;
    cursor: pointer;
  }

  .filter-panel__range-labels {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    color: #6b7280;
  }

  .filter-panel__actions {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #e5e7eb;
  }

  .filter-panel__apply {
    width: 100%;
    padding: 10px 16px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .filter-panel__apply:hover {
    background: #2563eb;
  }

  .filter-panel__apply:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  @media (prefers-reduced-motion: reduce) {
    .filter-panel__reset,
    .filter-panel__input,
    .filter-panel__apply {
      transition: none;
    }
  }

  @media (max-width: 768px) {
    .filter-panel {
      width: 100%;
    }
  }
`;
