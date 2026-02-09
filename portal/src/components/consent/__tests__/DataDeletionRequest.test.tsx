import { describe, it, expect, beforeEach, vi } from 'vitest';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { DataExportStatus } from '@/components/consent/DataExportStatus';
import i18n from '@/i18n';

expectTranslations(['Your data rights', 'Request deletion']);

describe.each(TEST_LOCALES)('DataExportStatus deletion (%s)', (locale) => {
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
      return Promise.resolve({ json: () => Promise.resolve({}) } as unknown as Response);
    }) as unknown as typeof fetch;
  });

  it('submits a deletion request', async () => {
    renderWithI18n(<DataExportStatus dataSubject="virtengine1test" />);

    await screen.findByText(i18n.t('Your data rights', { lng: locale }));
    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Request deletion', { lng: locale }) })
    );

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/consent/deletion',
        expect.objectContaining({ method: 'POST' })
      );
    });
  });
});
