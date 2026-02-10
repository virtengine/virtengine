'use client';

import { useEffect, useMemo, useState } from 'react';
import { Button } from '@/components/ui/Button';
import type {
  NotificationChannel,
  NotificationPreferences,
  NotificationType,
} from '@/types/notifications';
import { useTranslation } from 'react-i18next';

const CHANNELS: NotificationChannel[] = ['push', 'email', 'in_app'];

export function NotificationPreferencesPanel() {
  const { t } = useTranslation();
  const [prefs, setPrefs] = useState<NotificationPreferences | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetch('/api/notification-preferences')
      .then((res) => res.json())
      .then((data: NotificationPreferences) => setPrefs(data))
      .catch(() => {});
  }, []);

  const toggleChannel = (type: NotificationType, channel: NotificationChannel) => {
    if (!prefs) return;
    const current = prefs.channels[type] ?? [];
    const next = current.includes(channel)
      ? current.filter((item) => item !== channel)
      : [...current, channel];
    setPrefs({ ...prefs, channels: { ...prefs.channels, [type]: next } });
  };

  const toggleFrequency = (type: NotificationType) => {
    if (!prefs) return;
    const next = prefs.frequencies[type] === 'digest' ? 'immediate' : ('digest' as const);
    setPrefs({ ...prefs, frequencies: { ...prefs.frequencies, [type]: next } });
  };

  const handleSave = async () => {
    if (!prefs) return;
    setSaving(true);
    await fetch('/api/notification-preferences', {
      method: 'PUT',
      body: JSON.stringify(prefs),
      headers: { 'Content-Type': 'application/json' },
    });
    setSaving(false);
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

  if (!prefs) {
    return <p className="text-sm text-muted-foreground">{t('Loading preferences…')}</p>;
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-border bg-muted/20 p-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="text-sm font-semibold">{t('Quiet hours')}</p>
            <p className="text-xs text-muted-foreground">{quietHoursLabel}</p>
          </div>
          <button
            type="button"
            onClick={() =>
              setPrefs({
                ...prefs,
                quietHours: { ...prefs.quietHours, enabled: !prefs.quietHours.enabled },
              })
            }
            className="rounded-lg border border-border px-3 py-1.5 text-xs"
            role="switch"
            aria-checked={prefs.quietHours.enabled}
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
              onChange={(event) =>
                setPrefs({
                  ...prefs,
                  quietHours: { ...prefs.quietHours, startHour: Number(event.target.value) },
                })
              }
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
              onChange={(event) =>
                setPrefs({
                  ...prefs,
                  quietHours: { ...prefs.quietHours, endHour: Number(event.target.value) },
                })
              }
              className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
            />
          </label>
          <label className="text-xs text-muted-foreground">
            {t('Timezone')}
            <input
              type="text"
              value={prefs.quietHours.timezone}
              onChange={(event) =>
                setPrefs({
                  ...prefs,
                  quietHours: { ...prefs.quietHours, timezone: event.target.value },
                })
              }
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
            onClick={() => setPrefs({ ...prefs, digestEnabled: !prefs.digestEnabled })}
            className="rounded-lg border border-border px-3 py-1.5 text-xs"
            role="switch"
            aria-checked={prefs.digestEnabled}
          >
            {prefs.digestEnabled ? t('Enabled') : t('Disabled')}
          </button>
        </div>
        <label className="mt-3 block text-xs text-muted-foreground">
          {t('Digest time (UTC)')}
          <input
            type="time"
            value={prefs.digestTime}
            onChange={(event) => setPrefs({ ...prefs, digestTime: event.target.value })}
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

      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={saving}>
          {saving ? t('Saving…') : t('Save notification preferences')}
        </Button>
      </div>
    </div>
  );
}
