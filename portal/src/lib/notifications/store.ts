import type {
  Notification,
  NotificationPreferences,
  NotificationType,
  NotificationChannel,
} from '@/types/notifications';

interface NotificationStore {
  notifications: Notification[];
  preferences: NotificationPreferences;
}

const defaultPreferences: NotificationPreferences = {
  channels: {
    veid_status: ['in_app', 'email'],
    order_update: ['in_app', 'push', 'email'],
    escrow_deposit: ['in_app', 'email'],
    security_alert: ['in_app', 'push', 'email'],
    provider_alert: ['in_app', 'push'],
  },
  quietHours: {
    enabled: true,
    startHour: 22,
    endHour: 6,
    timezone: 'UTC',
  },
  digestEnabled: false,
  digestTime: '09:00',
};

const seedNotifications: Notification[] = [
  {
    id: 'notif-001',
    type: 'security_alert',
    title: 'New login detected',
    body: 'A new login from Chicago, IL was detected. Review if this was you.',
    createdAt: new Date(Date.now() - 1000 * 60 * 12).toISOString(),
    readAt: null,
    data: { action_url: '/account/settings/security' },
  },
  {
    id: 'notif-002',
    type: 'order_update',
    title: 'Order VE-3821 confirmed',
    body: 'Your compute order is now active and billing has started.',
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 3).toISOString(),
    readAt: null,
    data: { order_id: 'VE-3821' },
  },
  {
    id: 'notif-003',
    type: 'veid_status',
    title: 'VEID verification approved',
    body: 'Your VEID status changed to approved.',
    createdAt: new Date(Date.now() - 1000 * 60 * 60 * 22).toISOString(),
    readAt: new Date(Date.now() - 1000 * 60 * 60 * 20).toISOString(),
  },
];

function ensureStore(): NotificationStore {
  const globalAny = globalThis as typeof globalThis & { __veNotifications?: NotificationStore };
  if (!globalAny.__veNotifications) {
    globalAny.__veNotifications = {
      notifications: seedNotifications,
      preferences: defaultPreferences,
    };
  }
  return globalAny.__veNotifications;
}

export function listNotifications(): { notifications: Notification[]; unreadCount: number } {
  const store = ensureStore();
  const unreadCount = store.notifications.filter((n) => !n.readAt).length;
  return { notifications: store.notifications, unreadCount };
}

export function markNotificationsRead(ids: string[]): void {
  const store = ensureStore();
  store.notifications = store.notifications.map((notification) =>
    ids.includes(notification.id)
      ? { ...notification, readAt: new Date().toISOString() }
      : notification
  );
}

export function getPreferences(): NotificationPreferences {
  const store = ensureStore();
  return store.preferences;
}

export function updatePreferences(
  update: Partial<NotificationPreferences>
): NotificationPreferences {
  const store = ensureStore();
  store.preferences = {
    ...store.preferences,
    ...update,
    channels: update.channels ?? store.preferences.channels,
    quietHours: update.quietHours ?? store.preferences.quietHours,
  };
  return store.preferences;
}

export function setChannelPreference(
  notificationType: NotificationType,
  channel: NotificationChannel,
  enabled: boolean
): NotificationPreferences {
  const store = ensureStore();
  const current = new Set(store.preferences.channels[notificationType]);
  if (enabled) {
    current.add(channel);
  } else {
    current.delete(channel);
  }
  store.preferences.channels = {
    ...store.preferences.channels,
    [notificationType]: Array.from(current),
  };
  return store.preferences;
}
