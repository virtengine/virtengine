'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Alert, AlertDescription } from '@/components/ui/Alert';
import type {
  NotificationChannel,
  NotificationPreferences,
  NotificationType,
} from '@/types/notifications';
import { useTranslation } from 'react-i18next';

const CHANNELS: NotificationChannel[] = ['push', 'email', 'in_app'];
const TYPES: NotificationType[] = [
  'veid_status',
  'order_update',
  'escrow_deposit',
  'security_alert',
  'provider_alert',
];
const SETTINGS_KEYS = ['channels', 'digestEnabled', 'digestTime', 'frequencies', 'quietHours'];
const PREFERENCE_KEYS = ['userAddress', ...SETTINGS_KEYS];

const exactKeys = (value: Record<string, unknown>, keys: readonly string[]) => {
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return actual.length === expected.length && actual.every((key, index) => key === expected[index]);
};

const snapshotArray = (value: unknown[]) => {
  const length = value.length;
  return Array.from({ length }, (_, index) => value[index]);
};

const materializePreferences = (value: unknown): NotificationPreferences => {
  if (!value || typeof value !== 'object' || Array.isArray(value))
    throw new Error('invalid_response');
  const source = value as Record<string, unknown>;
  if (!exactKeys(source, PREFERENCE_KEYS)) throw new Error('invalid_response');
  const snapshot = Object.fromEntries(PREFERENCE_KEYS.map((key) => [key, source[key]]));
  const { userAddress, channels, frequencies, digestEnabled, digestTime, quietHours } = snapshot;
  if (
    typeof userAddress !== 'string' ||
    userAddress.length === 0 ||
    userAddress.length > 256 ||
    userAddress.trim() !== userAddress ||
    !channels ||
    typeof channels !== 'object' ||
    Array.isArray(channels) ||
    !frequencies ||
    typeof frequencies !== 'object' ||
    Array.isArray(frequencies) ||
    typeof digestEnabled !== 'boolean' ||
    typeof digestTime !== 'string' ||
    !/^([01]\d|2[0-3]):[0-5]\d$/.test(digestTime) ||
    !quietHours ||
    typeof quietHours !== 'object' ||
    Array.isArray(quietHours)
  ) {
    throw new Error('invalid_response');
  }

  const channelSource = channels as Record<string, unknown>;
  const frequencySource = frequencies as Record<string, unknown>;
  const quietHoursSource = quietHours as Record<string, unknown>;
  if (
    !exactKeys(channelSource, TYPES) ||
    !exactKeys(frequencySource, TYPES) ||
    !exactKeys(quietHoursSource, ['enabled', 'endHour', 'startHour', 'timezone'])
  ) {
    throw new Error('invalid_response');
  }
  const channelSnapshot = Object.fromEntries(TYPES.map((type) => [type, channelSource[type]]));
  const frequencySnapshot = Object.fromEntries(TYPES.map((type) => [type, frequencySource[type]]));
  const materializedChannels = {} as Record<NotificationType, NotificationChannel[]>;
  const materializedFrequencies = {} as Record<NotificationType, 'immediate' | 'digest'>;
  for (const type of TYPES) {
    const channelValue = channelSnapshot[type];
    const frequency = frequencySnapshot[type];
    if (!Array.isArray(channelValue) || !['immediate', 'digest'].includes(frequency as string)) {
      throw new Error('invalid_response');
    }
    const channelArray = snapshotArray(channelValue);
    if (
      channelArray.some((channel) => !CHANNELS.includes(channel as NotificationChannel)) ||
      new Set(channelArray).size !== channelArray.length
    ) {
      throw new Error('invalid_response');
    }
    materializedChannels[type] = channelArray as NotificationChannel[];
    materializedFrequencies[type] = frequency as 'immediate' | 'digest';
  }

  const quietHoursSnapshot = Object.fromEntries(
    ['enabled', 'endHour', 'startHour', 'timezone'].map((key) => [key, quietHoursSource[key]])
  );
  const { enabled, endHour, startHour, timezone } = quietHoursSnapshot;
  if (
    typeof enabled !== 'boolean' ||
    !Number.isInteger(startHour) ||
    Number(startHour) < 0 ||
    Number(startHour) > 23 ||
    !Number.isInteger(endHour) ||
    Number(endHour) < 0 ||
    Number(endHour) > 23 ||
    typeof timezone !== 'string' ||
    timezone.length === 0 ||
    timezone.length > 64 ||
    timezone.trim() !== timezone
  ) {
    throw new Error('invalid_response');
  }

  return {
    userAddress,
    channels: materializedChannels,
    frequencies: materializedFrequencies,
    digestEnabled,
    digestTime,
    quietHours: { enabled, endHour: Number(endHour), startHour: Number(startHour), timezone },
  };
};

export function NotificationPreferencesPanel() {
  const { t } = useTranslation();
  const [prefs, setPrefs] = useState<NotificationPreferences | null>(null);
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const mountedRef = useRef(false);
  const generationRef = useRef(0);
  const saveInFlightRef = useRef(false);
  const saveControllerRef = useRef<AbortController | null>(null);
  const translationRef = useRef(t);
  translationRef.current = t;

  useEffect(() => {
    mountedRef.current = true;
    const controller = new AbortController();
    const load = async () => {
      try {
        const response = await fetch('/api/notification-preferences', {
          signal: controller.signal,
        });
        if (!response.ok) throw new Error('load_failed');
        const data = materializePreferences(await response.json());
        if (mountedRef.current && !controller.signal.aborted) setPrefs(data);
      } catch {
        if (mountedRef.current && !controller.signal.aborted) {
          setError(translationRef.current('Notification preferences are unavailable.'));
        }
      } finally {
        if (mountedRef.current && !controller.signal.aborted) setLoading(false);
      }
    };
    void load();
    return () => {
      mountedRef.current = false;
      controller.abort();
      saveControllerRef.current?.abort();
    };
  }, []);

  const editPreferences = (
    update: (current: NotificationPreferences) => NotificationPreferences
  ) => {
    if (!prefs || saveInFlightRef.current) return;
    generationRef.current += 1;
    setError(null);
    setSaved(false);
    setPrefs(update(prefs));
  };

  const toggleChannel = (type: NotificationType, channel: NotificationChannel) => {
    editPreferences((currentPreferences) => {
      const current = currentPreferences.channels[type] ?? [];
      const next = current.includes(channel)
        ? current.filter((item) => item !== channel)
        : [...current, channel];
      return {
        ...currentPreferences,
        channels: { ...currentPreferences.channels, [type]: next },
      };
    });
  };

  const toggleFrequency = (type: NotificationType) => {
    editPreferences((currentPreferences) => {
      const next = currentPreferences.frequencies[type] === 'digest' ? 'immediate' : 'digest';
      return {
        ...currentPreferences,
        frequencies: { ...currentPreferences.frequencies, [type]: next },
      };
    });
  };

  const handleSave = async () => {
    if (!prefs || saveInFlightRef.current) return;
    let submittedPreferences: NotificationPreferences;
    try {
      submittedPreferences = materializePreferences(prefs);
    } catch {
      setSaved(false);
      setError(t('Notification preferences were not saved.'));
      return;
    }
    saveInFlightRef.current = true;
    const controller = new AbortController();
    saveControllerRef.current = controller;
    const submittedGeneration = generationRef.current;
    setSaving(true);
    setError(null);
    setSaved(false);
    const settings = Object.fromEntries(
      SETTINGS_KEYS.map((key) => [key, submittedPreferences[key as keyof NotificationPreferences]])
    );
    try {
      const response = await fetch('/api/notification-preferences', {
        method: 'PUT',
        body: JSON.stringify(settings),
        headers: { 'Content-Type': 'application/json' },
        signal: controller.signal,
      });
      if (!response.ok) throw new Error('save_failed');
      const savedPreferences = materializePreferences(await response.json());
      if (JSON.stringify(savedPreferences) !== JSON.stringify(submittedPreferences)) {
        throw new Error('save_failed');
      }
      if (
        !mountedRef.current ||
        controller.signal.aborted ||
        generationRef.current !== submittedGeneration
      ) {
        return;
      }
      setPrefs(savedPreferences);
      setSaved(true);
    } catch {
      if (mountedRef.current && !controller.signal.aborted) {
        setError(translationRef.current('Notification preferences were not saved.'));
      }
    } finally {
      if (saveControllerRef.current === controller) {
        saveControllerRef.current = null;
        saveInFlightRef.current = false;
        if (mountedRef.current) setSaving(false);
      }
    }
  };

  const categoryLabels = useMemo<Record<NotificationType, string>>(
    () => ({
      veid_status: t('VEID verification'),
      order_update: t('Order updates'),
      escrow_deposit: t('Escrow activity'),
      security_alert: t('Security alerts'),
      provider_alert: t('Provider availability'),
    }),
    [t]
  );

  const channelLabels = useMemo<Record<NotificationChannel, string>>(
    () => ({
      push: t('Push'),
      email: t('Email'),
      in_app: t('In-app'),
    }),
    [t]
  );

  const quietHoursLabel = useMemo(() => {
    if (!prefs?.quietHours?.enabled) return t('Quiet hours off');
    return t('Quiet hours {{start}}:00–{{end}}:00 {{timezone}}', {
      start: prefs.quietHours.startHour,
      end: prefs.quietHours.endHour,
      timezone: prefs.quietHours.timezone,
    });
  }, [prefs, t]);

  if (loading) {
    return (
      <p role="status" aria-busy="true" className="text-sm text-muted-foreground">
        {t('Loading preferences…')}
      </p>
    );
  }

  if (!prefs) {
    return (
      <Alert variant="warning">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <p role="status" aria-busy={saving} className="sr-only">
        {saving ? t('Saving…') : ''}
      </p>
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {saved && (
        <Alert variant="success">
          <AlertDescription>{t('Notification preferences saved.')}</AlertDescription>
        </Alert>
      )}
      <fieldset disabled={saving} className="space-y-6">
        <div className="rounded-lg border border-border bg-muted/20 p-4">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <p className="text-sm font-semibold">{t('Quiet hours')}</p>
              <p className="text-xs text-muted-foreground">{quietHoursLabel}</p>
            </div>
            <button
              type="button"
              onClick={() =>
                editPreferences((current) => ({
                  ...current,
                  quietHours: { ...current.quietHours, enabled: !current.quietHours.enabled },
                }))
              }
              className="rounded-lg border border-border px-3 py-1.5 text-xs"
              role="switch"
              aria-checked={prefs.quietHours.enabled}
              aria-label={t('Quiet hours')}
            >
              {prefs.quietHours.enabled ? t('Enabled') : t('Disabled')}
            </button>
          </div>
          <div className="mt-4 grid gap-3 sm:grid-cols-3">
            <label className="text-xs text-muted-foreground">
              {t('Start hour')}
              <input
                type="number"
                min={0}
                max={23}
                value={prefs.quietHours.startHour}
                onChange={(event) => {
                  const startHour = Number(event.target.value);
                  editPreferences((current) => ({
                    ...current,
                    quietHours: { ...current.quietHours, startHour },
                  }));
                }}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </label>
            <label className="text-xs text-muted-foreground">
              {t('End hour')}
              <input
                type="number"
                min={0}
                max={23}
                value={prefs.quietHours.endHour}
                onChange={(event) => {
                  const endHour = Number(event.target.value);
                  editPreferences((current) => ({
                    ...current,
                    quietHours: { ...current.quietHours, endHour },
                  }));
                }}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </label>
            <label className="text-xs text-muted-foreground">
              {t('Timezone')}
              <input
                type="text"
                value={prefs.quietHours.timezone}
                onChange={(event) => {
                  const timezone = event.target.value;
                  editPreferences((current) => ({
                    ...current,
                    quietHours: { ...current.quietHours, timezone },
                  }));
                }}
                className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
              />
            </label>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-muted/20 p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-semibold">{t('Weekly digest')}</p>
              <p className="text-xs text-muted-foreground">
                {t('Bundle non-critical updates into a digest.')}
              </p>
            </div>
            <button
              type="button"
              onClick={() =>
                editPreferences((current) => ({
                  ...current,
                  digestEnabled: !current.digestEnabled,
                }))
              }
              className="rounded-lg border border-border px-3 py-1.5 text-xs"
              role="switch"
              aria-checked={prefs.digestEnabled}
              aria-label={t('Weekly digest')}
            >
              {prefs.digestEnabled ? t('Enabled') : t('Disabled')}
            </button>
          </div>
          <label className="mt-3 block text-xs text-muted-foreground">
            {t('Digest time (UTC)')}
            <input
              type="time"
              value={prefs.digestTime}
              onChange={(event) => {
                const digestTime = event.target.value;
                editPreferences((current) => ({ ...current, digestTime }));
              }}
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
            />
          </label>
        </div>

        <div className="space-y-4">
          {Object.entries(categoryLabels).map(([type, label]) => {
            const category = type as NotificationType;
            const channels = prefs.channels[category] ?? [];
            return (
              <div key={type} className="rounded-lg border border-border bg-muted/20 p-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="text-sm font-semibold">{label}</p>
                    <p className="text-xs text-muted-foreground">{category.replace('_', ' ')}</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => toggleFrequency(category)}
                    className="rounded-lg border border-border px-3 py-1.5 text-xs"
                    aria-label={t('{{category}} frequency', { category: label })}
                    aria-pressed={prefs.frequencies[category] === 'digest'}
                  >
                    {prefs.frequencies[category] === 'digest' ? t('Digest') : t('Immediate')}
                  </button>
                </div>
                <div className="mt-4 flex flex-wrap gap-3">
                  {CHANNELS.map((channel) => (
                    <button
                      key={channel}
                      type="button"
                      onClick={() => toggleChannel(category, channel)}
                      aria-pressed={channels.includes(channel)}
                      aria-label={t('{{channel}}, {{category}}', {
                        channel: channelLabels[channel],
                        category: label,
                      })}
                      className={`rounded-full px-3 py-1 text-xs ${
                        channels.includes(channel)
                          ? 'bg-primary text-primary-foreground'
                          : 'border border-border text-muted-foreground'
                      }`}
                    >
                      {channelLabels[channel]}
                    </button>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </fieldset>

      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={saving}>
          {saving ? t('Saving…') : t('Save notification preferences')}
        </Button>
      </div>
    </div>
  );
}
