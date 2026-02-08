import crypto from 'crypto';
import type {
  ConsentEvent,
  ConsentRecord,
  ConsentSettingsResponse,
  ConsentPurpose,
  DataExportRequest,
  DeletionRequest,
} from '@/types/consent';

const DEFAULT_SUBJECT = 'virtengine1demo';

const now = new Date();

let consentVersion = 3;

const consentRecords = new Map<string, ConsentRecord[]>([
  [
    DEFAULT_SUBJECT,
    [
      {
        id: 'consent-biometric-001',
        dataSubject: DEFAULT_SUBJECT,
        scopeId: 'veid.biometric',
        purpose: 'biometric_processing',
        status: 'active',
        policyVersion: '1.0',
        consentVersion,
        grantedAt: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000).toISOString(),
        consentHash: hashString('biometric-processing'),
        signatureHash: hashString('sig-biometric'),
      },
      {
        id: 'consent-retention-001',
        dataSubject: DEFAULT_SUBJECT,
        scopeId: 'veid.data_retention',
        purpose: 'data_retention',
        status: 'active',
        policyVersion: '1.0',
        consentVersion,
        grantedAt: new Date(now.getTime() - 8 * 24 * 60 * 60 * 1000).toISOString(),
        consentHash: hashString('retention'),
        signatureHash: hashString('sig-retention'),
      },
      {
        id: 'consent-marketing-001',
        dataSubject: DEFAULT_SUBJECT,
        scopeId: 'veid.marketing',
        purpose: 'marketing',
        status: 'withdrawn',
        policyVersion: '1.0',
        consentVersion: 2,
        grantedAt: new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        withdrawnAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString(),
        consentHash: hashString('marketing'),
        signatureHash: hashString('sig-marketing'),
      },
    ],
  ],
]);

const consentEvents = new Map<string, ConsentEvent[]>([
  [
    DEFAULT_SUBJECT,
    [
      {
        id: 'event-001',
        consentId: 'consent-biometric-001',
        dataSubject: DEFAULT_SUBJECT,
        scopeId: 'veid.biometric',
        purpose: 'biometric_processing',
        eventType: 'granted',
        occurredAt: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000).toISOString(),
        blockHeight: 182032,
        details: 'Initial biometric consent granted',
      },
      {
        id: 'event-002',
        consentId: 'consent-retention-001',
        dataSubject: DEFAULT_SUBJECT,
        scopeId: 'veid.data_retention',
        purpose: 'data_retention',
        eventType: 'granted',
        occurredAt: new Date(now.getTime() - 8 * 24 * 60 * 60 * 1000).toISOString(),
        blockHeight: 182112,
        details: 'Retention consent granted',
      },
      {
        id: 'event-003',
        consentId: 'consent-marketing-001',
        dataSubject: DEFAULT_SUBJECT,
        scopeId: 'veid.marketing',
        purpose: 'marketing',
        eventType: 'revoked',
        occurredAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString(),
        blockHeight: 182900,
        details: 'Marketing consent withdrawn',
      },
    ],
  ],
]);

let exportRequests: DataExportRequest[] = [];
let deletionRequests: DeletionRequest[] = [];

export function getConsentSettings(dataSubject: string): ConsentSettingsResponse {
  const subject = dataSubject || DEFAULT_SUBJECT;
  const consents = consentRecords.get(subject) ?? [];
  const history = consentEvents.get(subject) ?? [];
  const lastUpdatedAt = consents
    .map((consent) => consent.grantedAt)
    .sort()
    .slice(-1)[0];

  return {
    dataSubject: subject,
    consentVersion,
    lastUpdatedAt: lastUpdatedAt ?? new Date().toISOString(),
    consents,
    history,
  };
}

export function grantConsent(params: {
  dataSubject: string;
  scopeId: string;
  purpose: ConsentPurpose;
  consentText: string;
  signature: string;
}): ConsentRecord {
  const subject = params.dataSubject || DEFAULT_SUBJECT;
  const consents = consentRecords.get(subject) ?? [];
  const existing = consents.find((consent) => consent.scopeId === params.scopeId);
  const now = new Date().toISOString();

  consentVersion += 1;

  const record: ConsentRecord = {
    id: existing?.id ?? `consent-${params.scopeId}-${Date.now()}`,
    dataSubject: subject,
    scopeId: params.scopeId,
    purpose: params.purpose,
    status: 'active',
    policyVersion: '1.0',
    consentVersion,
    grantedAt: now,
    consentHash: hashString(params.consentText),
    signatureHash: hashString(params.signature),
  };

  const next = consents.filter((consent) => consent.scopeId !== params.scopeId).concat(record);
  consentRecords.set(subject, next);

  appendEvent(subject, {
    id: `event-${Date.now()}`,
    consentId: record.id,
    dataSubject: subject,
    scopeId: params.scopeId,
    purpose: params.purpose,
    eventType: 'granted',
    occurredAt: now,
    blockHeight: randomHeight(),
    details: 'Consent granted via privacy center',
  });

  return record;
}

export function withdrawConsent(params: {
  dataSubject: string;
  consentId: string;
}): ConsentRecord | null {
  const subject = params.dataSubject || DEFAULT_SUBJECT;
  const consents = consentRecords.get(subject) ?? [];
  const now = new Date().toISOString();

  const updated = consents.map((consent) =>
    consent.id === params.consentId
      ? {
          ...consent,
          status: 'withdrawn' as const,
          withdrawnAt: now,
          consentVersion: ++consentVersion,
        }
      : consent
  );

  const record = updated.find((consent) => consent.id === params.consentId) ?? null;
  consentRecords.set(subject, updated);

  if (record) {
    appendEvent(subject, {
      id: `event-${Date.now()}`,
      consentId: record.id,
      dataSubject: subject,
      scopeId: record.scopeId,
      purpose: record.purpose,
      eventType: 'revoked',
      occurredAt: now,
      blockHeight: randomHeight(),
      details: 'Consent withdrawn via privacy center',
    });
  }

  return record;
}

export function requestExport(dataSubject: string, format: 'json' | 'csv'): DataExportRequest {
  const req: DataExportRequest = {
    id: `export-${Date.now()}`,
    dataSubject,
    requestedAt: new Date().toISOString(),
    status: 'processing',
    format,
  };
  exportRequests = [req, ...exportRequests];
  return req;
}

export function requestDeletion(dataSubject: string): DeletionRequest {
  const req: DeletionRequest = {
    id: `deletion-${Date.now()}`,
    dataSubject,
    requestedAt: new Date().toISOString(),
    status: 'pending',
  };
  deletionRequests = [req, ...deletionRequests];
  return req;
}

export function listRequests(dataSubject: string) {
  return {
    exports: exportRequests.filter((req) => req.dataSubject === dataSubject),
    deletions: deletionRequests.filter((req) => req.dataSubject === dataSubject),
  };
}

function appendEvent(subject: string, event: ConsentEvent) {
  const events = consentEvents.get(subject) ?? [];
  consentEvents.set(subject, [event, ...events]);
}

function hashString(value: string) {
  return crypto.createHash('sha256').update(value).digest('hex');
}

function randomHeight() {
  return 182000 + Math.floor(Math.random() * 1200);
}
