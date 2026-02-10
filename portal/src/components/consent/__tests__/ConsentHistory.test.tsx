import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { ConsentHistory } from '@/components/consent/ConsentHistory';
import i18n from '@/i18n';

vi.mock('@/lib/portal-adapter', () => ({
  useWallet: () => ({
    accounts: [{ address: 'virtengine1test' }],
    activeAccountIndex: 0,
  }),
}));

const historySettings = {
  dataSubject: 'virtengine1test',
  consentVersion: 2,
  lastUpdatedAt: new Date().toISOString(),
  consents: [],
  history: [
    {
      id: 'event-1',
      consentId: 'consent-1',
      dataSubject: 'virtengine1test',
      scopeId: 'veid.biometric',
      purpose: 'biometric_processing',
      eventType: 'granted',
      occurredAt: new Date().toISOString(),
      blockHeight: 12345,
      details: 'Granted',
    },
  ],
};

expectTranslations([
  'Consent history',
  'Audit trail of every consent change on your VEID profile.',
]);

describe.each(TEST_LOCALES)('ConsentHistory (%s)', (locale) => {
  beforeEach(async () => {
    await setLocale(locale);
    global.fetch = vi.fn(() =>
      Promise.resolve({
        json: () => Promise.resolve(historySettings),
      })
    ) as unknown as typeof fetch;
  });

  it('renders consent history entries', async () => {
    renderWithI18n(<ConsentHistory />);

    expect(await screen.findByText(i18n.t('Consent history', { lng: locale }))).toBeInTheDocument();
    expect(await screen.findByText('Granted')).toBeInTheDocument();
    expect(screen.getByText('veid.biometric')).toBeInTheDocument();
  });
});
