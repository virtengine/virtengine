import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { ConsentManager } from '@/components/consent/ConsentManager';
import i18n from '@/i18n';

vi.mock('@/lib/portal-adapter', () => ({
  useWallet: () => ({
    accounts: [{ address: 'virtengine1test' }],
    activeAccountIndex: 0,
  }),
}));

const baseSettings = {
  dataSubject: 'virtengine1test',
  consentVersion: 3,
  lastUpdatedAt: new Date('2024-01-01T00:00:00Z').toISOString(),
  consents: [
    {
      id: 'consent-biometric-001',
      dataSubject: 'virtengine1test',
      scopeId: 'veid.biometric',
      purpose: 'biometric_processing',
      status: 'active',
      policyVersion: '1.0',
      consentVersion: 3,
      grantedAt: new Date().toISOString(),
      consentHash: 'hash',
      signatureHash: 'sig',
    },
    {
      id: 'consent-retention-001',
      dataSubject: 'virtengine1test',
      scopeId: 'veid.data_retention',
      purpose: 'data_retention',
      status: 'withdrawn',
      policyVersion: '1.0',
      consentVersion: 3,
      grantedAt: new Date().toISOString(),
      withdrawnAt: new Date().toISOString(),
      consentHash: 'hash',
      signatureHash: 'sig',
    },
  ],
  history: [],
};

expectTranslations([
  'Consent controls',
  'Manage how VirtEngine processes your identity data across the VEID network.',
  'Required consent withdrawn',
  'Unable to load consent settings.',
]);

describe.each(TEST_LOCALES)('ConsentManager (%s)', (locale) => {
  beforeEach(async () => {
    await setLocale(locale);
    global.fetch = vi.fn((input: RequestInfo | URL) => {
      const url =
        typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      if (url.includes('/api/consent/requests')) {
        return Promise.resolve({
          json: () => Promise.resolve({ exports: [], deletions: [] }),
        } as unknown as Response);
      }
      if (url.includes('/api/consent/grant') || url.includes('/api/consent/withdraw')) {
        return Promise.resolve({ json: () => Promise.resolve({}) } as unknown as Response);
      }
      return Promise.resolve({
        json: () => Promise.resolve(baseSettings),
      } as unknown as Response);
    }) as unknown as typeof fetch;
  });

  it('renders consent status and required alert', async () => {
    renderWithI18n(<ConsentManager />);

    expect(
      await screen.findByText(i18n.t('Consent controls', { lng: locale }))
    ).toBeInTheDocument();
    expect(
      screen.getByText(i18n.t('Required consent withdrawn', { lng: locale }))
    ).toBeInTheDocument();
  });

  it('grants inactive consent and refreshes settings', async () => {
    renderWithI18n(<ConsentManager />);

    await screen.findByText(i18n.t('Consent controls', { lng: locale }));
    fireEvent.click(screen.getAllByRole('button', { name: i18n.t('Grant', { lng: locale }) })[0]);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/consent/grant',
        expect.objectContaining({ method: 'POST' })
      );
    });
  });
});
