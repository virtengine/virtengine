export type ConsentPurpose =
  | 'biometric_processing'
  | 'data_retention'
  | 'third_party_sharing'
  | 'marketing'
  | 'analytics';

export type ConsentStatus = 'active' | 'withdrawn' | 'expired';

export type ConsentRecord = {
  id: string;
  dataSubject: string;
  scopeId: string;
  purpose: ConsentPurpose;
  status: ConsentStatus;
  policyVersion: string;
  consentVersion: number;
  grantedAt: string;
  expiresAt?: string;
  withdrawnAt?: string;
  consentHash: string;
  acknowledgementHash: string;
  ipAddressHash?: string;
  detailedRecordRef?: string;
};

export type ConsentEventType = 'granted' | 'revoked' | 'updated' | 'expired';

export type ConsentEventBase = {
  id: string;
  consentId: string;
  dataSubject: string;
  scopeId: string;
  purpose: ConsentPurpose;
  eventType: ConsentEventType;
  occurredAt: string;
  details?: string;
};

export type ConsentChainEvidence = {
  blockHeight: number;
  txHash: string;
  chainId: string;
  code: 0;
  eventId: string;
  consentId: string;
};

export type LocalConsentEvent = ConsentEventBase & {
  source: 'local';
  chain?: never;
};

export type ChainConsentEvent = ConsentEventBase & {
  source: 'chain';
  chain: ConsentChainEvidence;
};

export type ConsentEvent = LocalConsentEvent | ChainConsentEvent;

const CONSENT_PURPOSES: ConsentPurpose[] = [
  'biometric_processing',
  'data_retention',
  'third_party_sharing',
  'marketing',
  'analytics',
];

const CONSENT_EVENT_TYPES: ConsentEventType[] = ['granted', 'revoked', 'updated', 'expired'];

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function parseEventBase(value: unknown): ConsentEventBase | null {
  if (!isRecord(value)) return null;
  const requiredStrings = ['id', 'consentId', 'dataSubject', 'scopeId', 'occurredAt'] as const;
  if (requiredStrings.some((key) => typeof value[key] !== 'string' || value[key] === '')) {
    return null;
  }
  if (!CONSENT_PURPOSES.includes(value.purpose as ConsentPurpose)) return null;
  if (!CONSENT_EVENT_TYPES.includes(value.eventType as ConsentEventType)) return null;
  if (value.details !== undefined && typeof value.details !== 'string') return null;

  return {
    id: value.id as string,
    consentId: value.consentId as string,
    dataSubject: value.dataSubject as string,
    scopeId: value.scopeId as string,
    purpose: value.purpose as ConsentPurpose,
    eventType: value.eventType as ConsentEventType,
    occurredAt: value.occurredAt as string,
    ...(typeof value.details === 'string' ? { details: value.details } : {}),
  };
}

export function projectConsentEventToChain(
  event: ConsentEventBase,
  evidence: unknown
): ChainConsentEvent | null {
  if (!isRecord(evidence)) return null;
  if (
    !Number.isInteger(evidence.blockHeight) ||
    (evidence.blockHeight as number) <= 0 ||
    typeof evidence.txHash !== 'string' ||
    evidence.txHash.trim() === '' ||
    typeof evidence.chainId !== 'string' ||
    evidence.chainId.trim() === '' ||
    evidence.code !== 0 ||
    evidence.eventId !== event.id ||
    evidence.consentId !== event.consentId
  ) {
    return null;
  }

  return {
    ...event,
    source: 'chain',
    chain: {
      blockHeight: evidence.blockHeight as number,
      txHash: evidence.txHash,
      chainId: evidence.chainId,
      code: 0,
      eventId: evidence.eventId,
      consentId: evidence.consentId,
    },
  };
}

export function normalizeConsentEvent(value: unknown): ConsentEvent | null {
  const event = parseEventBase(value);
  if (!event) return null;
  if (isRecord(value) && value.source === 'chain') {
    return projectConsentEventToChain(event, value.chain) ?? { ...event, source: 'local' };
  }
  return { ...event, source: 'local' };
}

export type ConsentSettingsResponse = {
  dataSubject: string;
  consentVersion: number;
  lastUpdatedAt: string;
  consents: ConsentRecord[];
  history: ConsentEvent[];
};

export type ExportStatus = 'pending' | 'processing' | 'ready' | 'failed' | 'expired';

export type DeletionStatus = 'pending' | 'blocked' | 'processing' | 'complete' | 'failed';

export type DataExportRequest = {
  id: string;
  dataSubject: string;
  requestedAt: string;
  status: ExportStatus;
  format: 'json' | 'csv';
  downloadUrl?: string;
  expiresAt?: string;
  error?: string;
};

export type DeletionRequest = {
  id: string;
  dataSubject: string;
  requestedAt: string;
  status: DeletionStatus;
  blockers?: string[];
  error?: string;
  completedAt?: string;
};
