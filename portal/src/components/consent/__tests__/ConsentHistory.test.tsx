import { beforeEach, describe, expect, it, vi } from 'vitest';
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

const eventBase = {
  id: 'event-1',
  consentId: 'consent-1',
  dataSubject: 'virtengine1test',
  scopeId: 'veid.biometric',
  purpose: 'biometric_processing',
  eventType: 'granted',
  occurredAt: '2026-08-02T00:00:00.000Z',
  details: 'Granted',
};

function mockHistory(history: unknown[]) {
  global.fetch = vi.fn(() =>
    Promise.resolve({ json: () => Promise.resolve({ history }) })
  ) as unknown as typeof fetch;
}

expectTranslations([
  'Consent history',
  'Audit trail of every consent change on your VEID profile.',
]);

describe.each(TEST_LOCALES)('ConsentHistory (%s)', (locale) => {
  beforeEach(async () => {
    await setLocale(locale);
    mockHistory([{ ...eventBase, source: 'local' }]);
  });

  it('renders local consent history truthfully', async () => {
    renderWithI18n(<ConsentHistory />);

    expect(await screen.findByText(i18n.t('Consent history', { lng: locale }))).toBeInTheDocument();
    expect(await screen.findByText('Granted')).toBeInTheDocument();
    expect(screen.getByText('Local record')).toBeInTheDocument();
    expect(screen.getByText('Not verified on chain')).toBeInTheDocument();
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
  });
});

describe('ConsentHistory provenance rendering', () => {
  beforeEach(async () => {
    await setLocale('en');
  });

  it('links chain details only for valid projected evidence', async () => {
    mockHistory([
      {
        ...eventBase,
        source: 'chain',
        chain: {
          blockHeight: 42,
          txHash: 'ABC123',
          chainId: 'virtengine-1',
          code: 0,
          eventId: eventBase.id,
          consentId: eventBase.consentId,
        },
      },
    ]);
    renderWithI18n(<ConsentHistory />);

    expect(await screen.findByText('Validated chain')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'View block 42 in explorer' })).toHaveAttribute(
      'href',
      expect.stringContaining('/block/42')
    );
    expect(
      screen.getByRole('link', { name: 'View consent transaction in explorer' })
    ).toHaveAttribute('href', expect.stringContaining('/tx/ABC123'));
  });

  it('renders malformed legacy height data as unverified local provenance', async () => {
    mockHistory([{ ...eventBase, blockHeight: 999, txHash: 'UNVALIDATED' }]);
    renderWithI18n(<ConsentHistory />);

    expect(await screen.findByText('Local record')).toBeInTheDocument();
    expect(screen.getByText('Not verified on chain')).toBeInTheDocument();
    expect(screen.queryByText('999')).not.toBeInTheDocument();
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
  });
});
