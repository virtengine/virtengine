import { chainConfig } from "./chain";

export type MarketListing = {
  id: string;
  title: string;
  price: string;
  providerAddress: string;
  providerName?: string;
  category?: string;
  description?: string;
  state?: string;
};

type RawRecord = Record<string, unknown>;

const OFFERING_ENDPOINTS = [
  "/virtengine/market/v1/offerings",
  "/virtengine/market/v1beta5/offerings",
  "/marketplace/offerings",
];

const PROVIDER_ENDPOINTS = [
  "/virtengine/provider/v1/providers",
  "/virtengine/provider/v1beta4/providers",
];

function coerceString(value: unknown, fallback = ""): string {
  if (typeof value === "string") return value;
  if (typeof value === "number") return value.toString();
  return fallback;
}

function asRecord(value: unknown): RawRecord | null {
  return value && typeof value === "object" ? (value as RawRecord) : null;
}

async function fetchJson(path: string): Promise<RawRecord> {
  const response = await fetch(`${chainConfig.rest}${path}`);
  if (!response.ok) {
    throw new Error(`Request failed (${response.status}) for ${path}`);
  }
  return (await response.json()) as RawRecord;
}

async function fetchFirstAvailable(paths: string[]): Promise<RawRecord> {
  let lastError: Error | undefined;
  for (const path of paths) {
    try {
      return await fetchJson(path);
    } catch (error) {
      lastError = error instanceof Error ? error : new Error("Request failed");
    }
  }

  throw lastError ?? new Error("No market endpoint responded");
}

function extractProviderAddress(record: RawRecord): string {
  const id = asRecord(record.id);
  return (
    coerceString(record.provider_address) ||
    coerceString(record.provider) ||
    coerceString(id?.providerAddress) ||
    coerceString(id?.provider_address)
  );
}

function extractSequence(record: RawRecord): string {
  const id = asRecord(record.id);
  return (
    coerceString(record.sequence) ||
    coerceString(record.seq) ||
    coerceString(id?.sequence) ||
    coerceString(id?.seq)
  );
}

function extractPrice(record: RawRecord): string {
  const directPrice = asRecord(record.price);
  if (directPrice) {
    const amount = coerceString(directPrice.amount);
    const denom = coerceString(directPrice.denom);
    if (amount) return `${amount} ${denom || "uve"}`;
  }

  const pricing = asRecord(record.pricing);
  if (pricing) {
    const amount = coerceString(pricing.base_price) || coerceString(pricing.basePrice);
    if (amount) return `${amount} uve`;
  }

  const prices = Array.isArray(record.prices) ? record.prices : [];
  for (const entry of prices) {
    const rawEntry = asRecord(entry);
    const rawPrice = asRecord(rawEntry?.price);
    const amount = coerceString(rawPrice?.amount);
    if (amount) {
      return `${amount} ${coerceString(rawPrice?.denom, "uve")}`;
    }
  }

  return "Unavailable";
}

function normalizeProviderDirectory(payload: RawRecord): Map<string, string> {
  const records = Array.isArray(payload.providers) ? payload.providers : [];
  const result = new Map<string, string>();
  for (const entry of records) {
    const provider = asRecord(entry);
    if (!provider) continue;
    const address =
      coerceString(provider.address) ||
      coerceString(provider.provider_address) ||
      coerceString(provider.owner);
    if (!address) continue;

    const metadata = asRecord(provider.metadata);
    const profile = asRecord(provider.profile);
    const name =
      coerceString(provider.display_name) ||
      coerceString(provider.name) ||
      coerceString(metadata?.display_name) ||
      coerceString(profile?.display_name);
    if (name) {
      result.set(address, name);
    }
  }
  return result;
}

function normalizeListing(record: RawRecord, providerNames: Map<string, string>): MarketListing {
  const providerAddress = extractProviderAddress(record);
  const sequence = extractSequence(record) || "0";
  const title =
    coerceString(record.name) || `${providerAddress || "unknown-provider"} offering ${sequence}`;
  return {
    id: `${providerAddress}/${sequence}`,
    title,
    price: extractPrice(record),
    providerAddress,
    providerName: providerNames.get(providerAddress),
    category: coerceString(record.category) || coerceString(record.resource_type),
    description: coerceString(record.description),
    state: coerceString(record.state) || coerceString(record.status),
  };
}

export async function fetchMarketListings(): Promise<MarketListing[]> {
  const offeringsPayload = await fetchFirstAvailable(
    OFFERING_ENDPOINTS.map((path) => `${path}?pagination.limit=25`)
  );

  let providerNames = new Map<string, string>();
  try {
    const providersPayload = await fetchFirstAvailable(
      PROVIDER_ENDPOINTS.map((path) => `${path}?pagination.limit=200`)
    );
    providerNames = normalizeProviderDirectory(providersPayload);
  } catch {
    providerNames = new Map<string, string>();
  }

  const offerings = Array.isArray(offeringsPayload.offerings) ? offeringsPayload.offerings : [];
  return offerings
    .map((record) => asRecord(record))
    .filter((record): record is RawRecord => Boolean(record))
    .map((record) => normalizeListing(record, providerNames));
}
