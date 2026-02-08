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
  signatureHash: string;
  ipAddressHash?: string;
  detailedRecordRef?: string;
};

export type ConsentEventType = 'granted' | 'revoked' | 'updated' | 'expired';

export type ConsentEvent = {
  id: string;
  consentId: string;
  dataSubject: string;
  scopeId: string;
  purpose: ConsentPurpose;
  eventType: ConsentEventType;
  occurredAt: string;
  blockHeight: number;
  details?: string;
};

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
