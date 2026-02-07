'use client';

import type {
  NotificationChannel,
  NotificationPreferences,
  NotificationType,
} from '@/types/notifications';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { useNotifications } from '@/components/notifications/hooks/useNotifications';

const categories: { type: NotificationType; label: string; description: string }[] = [
  {
    type: 'veid_status',
    label: 'VEID updates',
    description: 'Verification approvals, rejections, and document requests.',
  },
  {
    type: 'order_update',
    label: 'Order lifecycle',
    description: 'Order confirmations, provisioning, and completion updates.',
  },
  {
    type: 'escrow_deposit',
    label: 'Escrow activity',
    description: 'Deposits, releases, and settlement notices.',
  },
  {
    type: 'security_alert',
    label: 'Security alerts',
    description: 'Logins, MFA changes, and suspicious activity.',
  },
  {
    type: 'provider_alert',
    label: 'Provider availability',
    description: 'Capacity and availability alerts from providers you follow.',
  },
];

const channelLabels: Record<NotificationChannel, string> = {
  push: 'Push',
  email: 'Email',
  in_app: 'In-app',
};

export function NotificationPreferences() {
  const { preferences, updatePreferences } = useNotifications();

  if (!preferences) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Notification Preferences</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Loading preferences...</p>
        </CardContent>
      </Card>
    );
  }

  const handleChannelToggle = (
    category: NotificationType,
    channel: NotificationChannel,
    enabled: boolean
  ) => {
    const nextChannels = new Set(preferences.channels[category]);
    if (enabled) {
      nextChannels.add(channel);
    } else {
      nextChannels.delete(channel);
    }
    const next: NotificationPreferences = {
      ...preferences,
      channels: {
        ...preferences.channels,
        [category]: Array.from(nextChannels),
      },
    };
    void updatePreferences(next);
  };

  const toggleDigest = (enabled: boolean) => {
    void updatePreferences({ ...preferences, digestEnabled: enabled });
  };

  const updateQuietHours = (
    field: 'enabled' | 'startHour' | 'endHour',
    value: boolean | number
  ) => {
    void updatePreferences({
      ...preferences,
      quietHours: {
        ...preferences.quietHours,
        [field]: value,
      },
    });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Notification Preferences</CardTitle>
        <p className="text-sm text-muted-foreground">
          Choose how you want to be notified for critical events.
        </p>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-4">
          {categories.map((category) => (
            <div key={category.type} className="rounded-lg border border-border bg-muted/40 p-4">
              <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <h4 className="text-sm font-semibold">{category.label}</h4>
                  <p className="text-xs text-muted-foreground">{category.description}</p>
                </div>
                <div className="flex flex-wrap gap-3">
                  {(Object.keys(channelLabels) as NotificationChannel[]).map((channel) => (
                    <label key={channel} className="flex items-center gap-2 text-xs">
                      <input
                        type="checkbox"
                        className="h-4 w-4 rounded border-border"
                        checked={preferences.channels[category.type]?.includes(channel)}
                        onChange={(event) =>
                          handleChannelToggle(category.type, channel, event.target.checked)
                        }
                      />
                      <span>{channelLabels[channel]}</span>
                    </label>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="grid gap-4 rounded-lg border border-border bg-muted/40 p-4 sm:grid-cols-2">
          <div>
            <h4 className="text-sm font-semibold">Quiet Hours</h4>
            <p className="text-xs text-muted-foreground">
              Suppress push/email notifications during your night hours.
            </p>
          </div>
          <div className="space-y-2">
            <label className="flex items-center gap-2 text-xs">
              <input
                type="checkbox"
                className="h-4 w-4 rounded border-border"
                checked={preferences.quietHours.enabled}
                onChange={(event) => updateQuietHours('enabled', event.target.checked)}
              />
              Enable quiet hours
            </label>
            <div className="flex gap-2 text-xs">
              <label className="flex items-center gap-2">
                Start
                <input
                  type="number"
                  min={0}
                  max={23}
                  className="w-16 rounded border border-border bg-background px-2 py-1"
                  value={preferences.quietHours.startHour}
                  onChange={(event) => updateQuietHours('startHour', Number(event.target.value))}
                />
              </label>
              <label className="flex items-center gap-2">
                End
                <input
                  type="number"
                  min={0}
                  max={23}
                  className="w-16 rounded border border-border bg-background px-2 py-1"
                  value={preferences.quietHours.endHour}
                  onChange={(event) => updateQuietHours('endHour', Number(event.target.value))}
                />
              </label>
            </div>
          </div>
        </div>

        <div className="grid gap-4 rounded-lg border border-border bg-muted/40 p-4 sm:grid-cols-2">
          <div>
            <h4 className="text-sm font-semibold">Digest Email</h4>
            <p className="text-xs text-muted-foreground">
              Consolidate non-critical updates into a weekly digest.
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs">
            <input
              type="checkbox"
              className="h-4 w-4 rounded border-border"
              checked={preferences.digestEnabled}
              onChange={(event) => toggleDigest(event.target.checked)}
            />
            Enable digest
            <span className="text-muted-foreground">({preferences.digestTime} UTC)</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
