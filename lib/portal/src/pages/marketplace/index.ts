/**
 * Marketplace Pages
 * VE-703: Customer marketplace browse experience
 *
 * Provides the complete marketplace UI for browsing offerings from the chain.
 */

// Main page
export { MarketplacePage } from "./MarketplacePage";
export type { MarketplacePageProps } from "./MarketplacePage";

// Hooks
export {
  useOfferings,
  OFFERING_CATEGORIES,
  REGIONS,
} from "./hooks/useOfferings";
export type {
  UseOfferingsOptions,
  OfferingsState,
  OfferingsActions,
  OfferingCategory,
  Region,
} from "./hooks/useOfferings";

// Components
export { SearchBar } from "./components/SearchBar";
export type { SearchBarProps } from "./components/SearchBar";

export { FilterPanel } from "./components/FilterPanel";
export type { FilterPanelProps } from "./components/FilterPanel";

export { CategoryNav } from "./components/CategoryNav";
export type { CategoryNavProps } from "./components/CategoryNav";

export { OfferingGrid } from "./components/OfferingGrid";
export type { OfferingGridProps } from "./components/OfferingGrid";

export { OfferingDetailPage } from "./components/OfferingDetailPage";
export type { OfferingDetailPageProps } from "./components/OfferingDetailPage";

export { ProviderInfo, ProviderBadge } from "./components/ProviderInfo";
export type {
  ProviderInfoProps,
  ProviderBadgeProps,
} from "./components/ProviderInfo";

export { MarketplaceOfferingCard } from "./components/MarketplaceOfferingCard";
export type { MarketplaceOfferingCardProps } from "./components/MarketplaceOfferingCard";
