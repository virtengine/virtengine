'use client';

import { useCallback, useEffect, useState } from 'react';
import type { Notification, NotificationPreferences } from '@/types/notifications';

interface NotificationsState {
  notifications: Notification[];
  unreadCount: number;
  loading: boolean;
  preferences: NotificationPreferences | null;
}

export function useNotifications() {
  const [state, setState] = useState<NotificationsState>({
    notifications: [],
    unreadCount: 0,
    loading: true,
    preferences: null,
  });

  const fetchNotifications = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true }));
    const res = await fetch('/api/notifications');
    const data = await res.json();
    setState((prev) => ({
      ...prev,
      notifications: data.notifications ?? [],
      unreadCount: data.unreadCount ?? 0,
      loading: false,
    }));
  }, []);

  const fetchPreferences = useCallback(async () => {
    const res = await fetch('/api/notifications/preferences');
    const data = await res.json();
    setState((prev) => ({ ...prev, preferences: data.preferences }));
  }, []);

  const markAsRead = useCallback(async (ids: string[]) => {
    if (ids.length === 0) return;
    await fetch('/api/notifications/read', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids }),
    });
    setState((prev) => ({
      ...prev,
      notifications: prev.notifications.map((notification) =>
        ids.includes(notification.id)
          ? { ...notification, readAt: new Date().toISOString() }
          : notification
      ),
      unreadCount: Math.max(0, prev.unreadCount - ids.length),
    }));
  }, []);

  const updatePreferences = useCallback(async (preferences: NotificationPreferences) => {
    const res = await fetch('/api/notifications/preferences', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ preferences }),
    });
    const data = await res.json();
    setState((prev) => ({ ...prev, preferences: data.preferences }));
  }, []);

  useEffect(() => {
    void fetchNotifications();
    void fetchPreferences();
  }, [fetchNotifications, fetchPreferences]);

  return {
    ...state,
    refresh: fetchNotifications,
    markAsRead,
    markAllAsRead: () =>
      void markAsRead(state.notifications.filter((n) => !n.readAt).map((n) => n.id)),
    updatePreferences,
  };
}
