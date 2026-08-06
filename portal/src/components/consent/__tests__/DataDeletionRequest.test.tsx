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
          ok: true,
          json: () => Promise.resolve({ exports: [], deletions: [] }),
        } as unknown as Response);
      }
      return Promise.resolve({
        ok: true,
        status: 202,
        json: () =>
          Promise.resolve({
            id: 'deletion-authoritative-1',
            dataSubject: 'virtengine1test',
            requestedAt: '2026-08-04T00:00:00.000Z',
            status: 'pending',
          }),
      } as unknown as Response);
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
    expect(await screen.findByText('deletion-authoritative-1')).toBeInTheDocument();
  });

  it('shows failure without claiming a deletion request', async () => {
    global.fetch = vi.fn((input: RequestInfo | URL) => {
      const url =
        typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      if (url.includes('/api/consent/requests')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ exports: [], deletions: [] }),
        } as Response);
      }
      return Promise.resolve({ ok: false, status: 503 } as Response);
    }) as typeof fetch;
    renderWithI18n(<DataExportStatus dataSubject="virtengine1test" />);
    await screen.findByText(i18n.t('Your data rights', { lng: locale }));
    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Request deletion', { lng: locale }) })
    );

    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(screen.queryByText(/deletion-authoritative/)).not.toBeInTheDocument();
  });

  it('replaces an accepted pending record with newer server status', async () => {
    let requestLoads = 0;
    global.fetch = vi.fn((input: RequestInfo | URL) => {
      const url =
        typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      if (url.includes('/api/consent/requests')) {
        requestLoads += 1;
        return Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              exports: [],
              deletions:
                requestLoads > 1
                  ? [
                      {
                        id: 'deletion-authoritative-1',
                        dataSubject: 'virtengine1test',
                        requestedAt: '2026-08-04T00:00:00.000Z',
                        status: 'processing',
                      },
                    ]
                  : [],
            }),
        } as Response);
      }
      return Promise.resolve({
        ok: true,
        status: 202,
        json: () =>
          Promise.resolve({
            id: 'deletion-authoritative-1',
            dataSubject: 'virtengine1test',
            requestedAt: '2026-08-04T00:00:00.000Z',
            status: 'pending',
          }),
      } as Response);
    }) as typeof fetch;
    renderWithI18n(<DataExportStatus dataSubject="virtengine1test" />);
    await screen.findByText(i18n.t('Your data rights', { lng: locale }));
    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Request deletion', { lng: locale }) })
    );
    expect(await screen.findByText('pending')).toBeInTheDocument();

    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Request JSON export', { lng: locale }) })
    );
    await waitFor(() => expect(screen.getByText('processing')).toBeInTheDocument());
  });

  it('re-enables deletion after a pending request is aborted by subject change', async () => {
    let resolveDeletion!: (value: Response) => void;
    global.fetch = vi.fn((input: RequestInfo | URL) => {
      const url =
        typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;
      if (url.includes('/api/consent/requests')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ exports: [], deletions: [] }),
        } as Response);
      }
      return new Promise<Response>((resolve) => (resolveDeletion = resolve));
    }) as typeof fetch;
    const view = renderWithI18n(<DataExportStatus dataSubject="virtengine1first" />);
    await screen.findByText(i18n.t('Your data rights', { lng: locale }));
    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Request deletion', { lng: locale }) })
    );
    expect(screen.getByRole('button', { name: /Requesting deletion/i })).toBeDisabled();

    view.rerender(<DataExportStatus dataSubject="virtengine1second" />);
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: i18n.t('Request deletion', { lng: locale }) })
      ).toBeEnabled()
    );
    resolveDeletion({ ok: true, json: () => Promise.resolve({}) } as Response);
  });
});
