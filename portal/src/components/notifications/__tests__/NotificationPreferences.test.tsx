import { describe, it, expect, beforeEach, vi } from 'vitest';
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
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

const jsonResponse = (body: unknown, ok = true) =>
  ({
    ok,
    json: () => Promise.resolve(body),
  }) as Response;

const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((promiseResolve, promiseReject) => {
    resolve = promiseResolve;
    reject = promiseReject;
  });
  return { promise, resolve, reject };
};

const channelName = (channel: string, category: string, locale = 'en') =>
  i18n.t('{{channel}}, {{category}}', {
    channel: i18n.t(channel, { lng: locale }),
    category: i18n.t(category, { lng: locale }),
    lng: locale,
  });

expectTranslations(['Save notification preferences', 'Loading preferences…', 'Email']);

describe.each(TEST_LOCALES)('NotificationPreferencesPanel (%s)', (locale) => {
  beforeEach(async () => {
    await setLocale(locale);
  });

  it('loads preferences and shows saved only after the updated preferences are returned', async () => {
    const saveResponse = deferred<Response>();
    const updatedPreferences = {
      ...prefs,
      channels: { ...prefs.channels, veid_status: ['in_app', 'email'] },
    };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockReturnValueOnce(saveResponse.promise);
    global.fetch = fetchMock as typeof fetch;

    renderWithI18n(<NotificationPreferencesPanel />);

    const saveLabel = i18n.t('Save notification preferences', { lng: locale });
    const savedLabel = i18n.t('Notification preferences saved.', { lng: locale });
    const saveButton = await screen.findByRole('button', { name: saveLabel });

    fireEvent.click(
      screen.getByRole('button', {
        name: channelName('Email', 'VEID verification', locale),
      })
    );
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/notification-preferences',
        expect.objectContaining({ method: 'PUT' })
      );
    });

    const putOptions = fetchMock.mock.calls[1]?.[1] as RequestInit;
    const putBody = JSON.parse(putOptions.body as string) as Record<string, unknown>;
    expect(putBody).not.toHaveProperty('userAddress');
    expect(putBody).toMatchObject({
      channels: { veid_status: ['in_app', 'email'] },
    });
    expect(screen.queryByText(savedLabel)).not.toBeInTheDocument();
    expect(saveButton).toBeDisabled();

    await act(async () => {
      saveResponse.resolve(jsonResponse(updatedPreferences));
      await saveResponse.promise;
    });

    expect(await screen.findByText(savedLabel)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: saveLabel })).toBeEnabled();
  });

  it('gives every channel a category-specific accessible name', async () => {
    global.fetch = vi.fn().mockResolvedValueOnce(jsonResponse(prefs)) as typeof fetch;

    renderWithI18n(<NotificationPreferencesPanel />);
    await screen.findByRole('button', {
      name: i18n.t('Save notification preferences', { lng: locale }),
    });

    for (const category of [
      'VEID verification',
      'Order updates',
      'Escrow activity',
      'Security alerts',
      'Provider availability',
    ]) {
      for (const channel of ['Push', 'Email', 'In-app']) {
        expect(
          screen.getByRole('button', { name: channelName(channel, category, locale) })
        ).toBeInTheDocument();
      }
    }
    expect(
      screen.queryByRole('button', { name: i18n.t('Email', { lng: locale }) })
    ).not.toBeInTheDocument();
  });
});

describe('NotificationPreferencesPanel failures', () => {
  beforeEach(async () => {
    await setLocale('en');
  });

  it.each(['HTTP failure', 'network failure'])(
    'exits loading and shows unavailable after a load %s',
    async (failure) => {
      const fetchMock = vi.fn();
      if (failure === 'HTTP failure') {
        fetchMock.mockResolvedValueOnce(jsonResponse({ message: 'unavailable' }, false));
      } else {
        fetchMock.mockRejectedValueOnce(new Error('network unavailable'));
      }
      global.fetch = fetchMock as typeof fetch;

      renderWithI18n(<NotificationPreferencesPanel />);

      expect(screen.getByText(i18n.t('Loading preferences…', { lng: 'en' }))).toBeInTheDocument();
      expect(
        await screen.findByText(i18n.t('Notification preferences are unavailable.', { lng: 'en' }))
      ).toBeInTheDocument();
      expect(
        screen.queryByText(i18n.t('Loading preferences…', { lng: 'en' }))
      ).not.toBeInTheDocument();
    }
  );

  it.each(['500 response', 'network rejection'])(
    'shows not saved and recovers the save button after a %s',
    async (failure) => {
      const fetchMock = vi.fn().mockResolvedValueOnce(jsonResponse(prefs));
      if (failure === '500 response') {
        fetchMock.mockResolvedValueOnce(jsonResponse({ message: 'save failed' }, false));
      } else {
        fetchMock.mockRejectedValueOnce(new Error('network unavailable'));
      }
      global.fetch = fetchMock as typeof fetch;

      renderWithI18n(<NotificationPreferencesPanel />);
      const saveLabel = i18n.t('Save notification preferences', { lng: 'en' });
      fireEvent.click(await screen.findByRole('button', { name: saveLabel }));

      expect(
        await screen.findByText(i18n.t('Notification preferences were not saved.', { lng: 'en' }))
      ).toBeInTheDocument();
      expect(
        screen.queryByText(i18n.t('Notification preferences saved.', { lng: 'en' }))
      ).not.toBeInTheDocument();
      expect(screen.getByRole('button', { name: saveLabel })).toBeEnabled();
    }
  );

  it('rejects a successful response for a different user address', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockResolvedValueOnce(jsonResponse({ ...prefs, userAddress: 'virtengine1other' }));
    global.fetch = fetchMock as typeof fetch;

    renderWithI18n(<NotificationPreferencesPanel />);
    const saveLabel = i18n.t('Save notification preferences', { lng: 'en' });
    fireEvent.click(await screen.findByRole('button', { name: saveLabel }));

    expect(
      await screen.findByText(i18n.t('Notification preferences were not saved.', { lng: 'en' }))
    ).toBeInTheDocument();
    expect(
      screen.queryByText(i18n.t('Notification preferences saved.', { lng: 'en' }))
    ).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: saveLabel })).toBeEnabled();
  });

  it.each([
    ['additional data', { ...prefs, additional: true }],
    ['an invalid channel', { ...prefs, channels: { ...prefs.channels, veid_status: ['sms'] } }],
  ])('rejects a load response containing %s', async (_name, responseBody) => {
    global.fetch = vi.fn().mockResolvedValueOnce(jsonResponse(responseBody)) as typeof fetch;

    renderWithI18n(<NotificationPreferencesPanel />);

    expect(
      await screen.findByText(i18n.t('Notification preferences are unavailable.', { lng: 'en' }))
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('button', {
        name: i18n.t('Save notification preferences', { lng: 'en' }),
      })
    ).not.toBeInTheDocument();
  });

  it.each([
    ['additional data', { ...prefs, additional: true }],
    ['different valid settings', { ...prefs, digestEnabled: false }],
  ])('rejects a successful save response containing %s', async (_name, responseBody) => {
    global.fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockResolvedValueOnce(jsonResponse(responseBody)) as typeof fetch;

    renderWithI18n(<NotificationPreferencesPanel />);
    fireEvent.click(
      await screen.findByRole('button', {
        name: i18n.t('Save notification preferences', { lng: 'en' }),
      })
    );

    expect(
      await screen.findByText(i18n.t('Notification preferences were not saved.', { lng: 'en' }))
    ).toBeInTheDocument();
  });

  it('prevents duplicate same-tick saves and disables edits while saving', async () => {
    const saveResponse = deferred<Response>();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockReturnValueOnce(saveResponse.promise);
    global.fetch = fetchMock as typeof fetch;
    renderWithI18n(<NotificationPreferencesPanel />);
    const saveButton = await screen.findByRole('button', {
      name: i18n.t('Save notification preferences', { lng: 'en' }),
    });

    fireEvent.click(saveButton);
    fireEvent.click(saveButton);

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(saveButton).toBeDisabled();
    expect(
      screen.getByRole('switch', { name: i18n.t('Quiet hours', { lng: 'en' }) })
    ).toBeDisabled();
    expect(
      screen.getByRole('spinbutton', { name: i18n.t('Start hour', { lng: 'en' }) })
    ).toBeDisabled();
    expect(
      screen.getByRole('button', {
        name: channelName('Email', 'VEID verification'),
      })
    ).toBeDisabled();

    await act(async () => {
      saveResponse.resolve(jsonResponse(prefs));
      await saveResponse.promise;
    });
  });

  it('finishes a deferred save after a locale change without reloading or retaining the lock', async () => {
    const saveResponse = deferred<Response>();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockReturnValueOnce(saveResponse.promise);
    global.fetch = fetchMock as typeof fetch;
    renderWithI18n(<NotificationPreferencesPanel />);
    fireEvent.click(
      await screen.findByRole('button', {
        name: i18n.t('Save notification preferences', { lng: 'en' }),
      })
    );
    const signal = (fetchMock.mock.calls[1]?.[1] as RequestInit).signal as AbortSignal;

    await act(async () => {
      await setLocale('es');
    });

    expect(signal.aborted).toBe(false);
    expect(fetchMock).toHaveBeenCalledTimes(2);

    await act(async () => {
      saveResponse.resolve(jsonResponse(prefs));
      await saveResponse.promise;
    });

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(
      await screen.findByText(i18n.t('Notification preferences saved.', { lng: 'es' }))
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', {
        name: i18n.t('Save notification preferences', { lng: 'es' }),
      })
    ).toBeEnabled();
  });

  it('shows a deferred save rejection in the current locale after a locale change', async () => {
    const failureKey = 'Notification preferences were not saved.';
    const spanishFailure = 'No se guardaron las preferencias de notificaciones.';
    const previousSpanishFailure = i18n.getResource('es', 'translation', failureKey) ?? failureKey;
    i18n.addResource('es', 'translation', failureKey, spanishFailure);
    const saveResponse = deferred<Response>();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockReturnValueOnce(saveResponse.promise);
    try {
      global.fetch = fetchMock as typeof fetch;
      renderWithI18n(<NotificationPreferencesPanel />);
      fireEvent.click(
        await screen.findByRole('button', {
          name: i18n.t('Save notification preferences', { lng: 'en' }),
        })
      );

      await act(async () => {
        await setLocale('es');
      });

      act(() => {
        saveResponse.reject(new Error('network unavailable'));
      });

      expect(await screen.findByText(spanishFailure)).toBeInTheDocument();
      expect(
        screen.getByRole('button', {
          name: i18n.t('Save notification preferences', { lng: 'es' }),
        })
      ).toBeEnabled();
    } finally {
      i18n.addResource('es', 'translation', failureKey, previousSpanishFailure);
    }
  });

  it.each([
    {
      field: 'timezone',
      label: 'Timezone',
      invalidValue: ' UTC ',
      correctedValue: 'Europe/London',
      savedPreferences: {
        ...prefs,
        quietHours: { ...prefs.quietHours, timezone: 'Europe/London' },
      },
    },
    {
      field: 'digest time',
      label: 'Digest time (UTC)',
      invalidValue: '',
      correctedValue: '10:30',
      savedPreferences: { ...prefs, digestTime: '10:30' },
    },
  ])(
    'rejects invalid editable $field without a PUT and permits a corrected save',
    async ({ label, invalidValue, correctedValue, savedPreferences }) => {
      const fetchMock = vi
        .fn()
        .mockResolvedValueOnce(jsonResponse(prefs))
        .mockResolvedValueOnce(jsonResponse(savedPreferences));
      global.fetch = fetchMock as typeof fetch;
      renderWithI18n(<NotificationPreferencesPanel />);
      const saveLabel = i18n.t('Save notification preferences', { lng: 'en' });
      const input = await screen.findByLabelText(i18n.t(label, { lng: 'en' }));

      fireEvent.change(input, { target: { value: invalidValue } });
      fireEvent.click(screen.getByRole('button', { name: saveLabel }));

      expect(
        await screen.findByText(i18n.t('Notification preferences were not saved.', { lng: 'en' }))
      ).toBeInTheDocument();
      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(screen.getByRole('button', { name: saveLabel })).toBeEnabled();

      fireEvent.change(input, { target: { value: correctedValue } });
      fireEvent.click(screen.getByRole('button', { name: saveLabel }));

      expect(
        await screen.findByText(i18n.t('Notification preferences saved.', { lng: 'en' }))
      ).toBeInTheDocument();
      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(fetchMock).toHaveBeenLastCalledWith(
        '/api/notification-preferences',
        expect.objectContaining({ method: 'PUT' })
      );
      expect(screen.getByRole('button', { name: saveLabel })).toBeEnabled();
    }
  );

  it('aborts an in-flight save on unmount and ignores its late response', async () => {
    const saveResponse = deferred<Response>();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockReturnValueOnce(saveResponse.promise);
    global.fetch = fetchMock as typeof fetch;
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const view = renderWithI18n(<NotificationPreferencesPanel />);
    fireEvent.click(
      await screen.findByRole('button', {
        name: i18n.t('Save notification preferences', { lng: 'en' }),
      })
    );
    const signal = (fetchMock.mock.calls[1]?.[1] as RequestInit).signal as AbortSignal;

    view.unmount();
    expect(signal.aborted).toBe(true);
    await act(async () => {
      saveResponse.resolve(jsonResponse(prefs));
      await saveResponse.promise;
    });
    expect(consoleError).not.toHaveBeenCalled();
    consoleError.mockRestore();
  });

  it('clears the saved confirmation when preferences are edited', async () => {
    global.fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(prefs))
      .mockResolvedValueOnce(jsonResponse(prefs)) as typeof fetch;
    renderWithI18n(<NotificationPreferencesPanel />);
    fireEvent.click(
      await screen.findByRole('button', {
        name: i18n.t('Save notification preferences', { lng: 'en' }),
      })
    );
    const savedLabel = i18n.t('Notification preferences saved.', { lng: 'en' });
    expect(await screen.findByText(savedLabel)).toBeInTheDocument();

    fireEvent.click(
      screen.getByRole('button', {
        name: channelName('Email', 'VEID verification'),
      })
    );

    expect(screen.queryByText(savedLabel)).not.toBeInTheDocument();
  });
});
