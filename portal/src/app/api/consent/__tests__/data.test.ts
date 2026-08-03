import { describe, expect, it } from 'vitest';
import { getConsentSettings, grantConsent, withdrawConsent } from '@/app/api/consent/data';

describe('consent API data provenance', () => {
  it('exposes demo fixtures explicitly as local without chain fields', () => {
    const settings = getConsentSettings('virtengine1demo');

    expect(settings.history.length).toBeGreaterThan(0);
    expect(settings.history.every((event) => event.source === 'local')).toBe(true);
    expect(settings.history.every((event) => !('chain' in event))).toBe(true);
    expect(settings.history.some((event) => event.details?.includes('Demo fixture'))).toBe(true);
  });

  it('keeps privacy-center grants and withdrawals local', () => {
    const dataSubject = `virtengine1test-${Date.now()}`;
    const record = grantConsent({
      dataSubject,
      scopeId: 'veid.analytics',
      purpose: 'analytics',
      consentText: 'analytics consent',
      acknowledgement: 'local-acknowledgement',
    });
    expect(getConsentSettings(dataSubject).history[0]).toMatchObject({ source: 'local' });

    withdrawConsent({ dataSubject, consentId: record.id });
    const history = getConsentSettings(dataSubject).history;
    expect(history).toHaveLength(2);
    expect(history.every((event) => event.source === 'local' && !('chain' in event))).toBe(true);
  });
});
