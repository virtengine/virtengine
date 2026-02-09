import { describe, it, expect, beforeEach, vi } from 'vitest';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { NotificationPreferencesPanel } from '@/components/notifications/NotificationPreferences';
import i18n from '@/i18n';

const prefs = {
  userAddress: 'virtengine1demo',
  channels: {
    veid_status: ['in_app'],
    order_update: ['email'],
    escrow_deposit: [],
    security_alert: ['push'],
    provider_alert: ['in_app'],
  },
  frequencies: {
    veid_status: 'immediate',
    order_update: 'digest',
    escrow_deposit: 'immediate',
    security_alert: 'immediate',
    provider_alert: 'digest',
  },
  digestEnabled: true,
  digestTime: '09:00',
  quietHours: {
    enabled: false,
    startHour: 22,
    endHour: 6,
    timezone: 'UTC',
  },
};

expectTranslations(['Save notification preferences', 'Loading preferences…', 'Email']);

describe.each(TEST_LOCALES)('NotificationPreferencesPanel (%s)', (locale) => {
  beforeEach(async () => {
    await setLocale(locale);
    global.fetch = vi
      .fn()
      .mockResolvedValueOnce({ json: () => Promise.resolve(prefs) })
      .mockResolvedValueOnce({ json: () => Promise.resolve({}) }) as unknown as typeof fetch;
  });

  it('loads preferences and saves updates', async () => {
    renderWithI18n(<NotificationPreferencesPanel />);

    expect(
      await screen.findByText(i18n.t('Save notification preferences', { lng: locale }))
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: i18n.t('Email', { lng: locale }) })[0]);
    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Save notification preferences', { lng: locale }) })
    );

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/notification-preferences',
        expect.objectContaining({ method: 'PUT' })
      );
    });
  });
});
