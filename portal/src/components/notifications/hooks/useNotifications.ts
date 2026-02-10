'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { Notification } from '@/types/notifications';
import { env } from '@/config/env';

interface NotificationsResponse {
  notifications: Notification[];
  unreadCount: number;
}

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const refresh = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await fetch('/api/notifications');
      const data = (await res.json()) as NotificationsResponse;
      setNotifications(data.notifications);
      setUnreadCount(data.unreadCount);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load notifications');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const markAsRead = useCallback(async (ids: string[]) => {
    if (ids.length === 0) return;
    await fetch('/api/notifications/read', {
      method: 'POST',
      body: JSON.stringify({ ids }),
      headers: { 'Content-Type': 'application/json' },
    });
    setNotifications((prev) =>
      prev.map((notif) =>
        ids.includes(notif.id) ? { ...notif, readAt: new Date().toISOString() } : notif
      )
    );
    setUnreadCount((prev) => Math.max(0, prev - ids.length));
  }, []);

  const markAllAsRead = useCallback(() => {
    const ids = notifications.filter((notif) => !notif.readAt).map((notif) => notif.id);
    void markAsRead(ids);
  }, [markAsRead, notifications]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!env.notificationsWsUrl || wsRef.current) return;

    const ws = new WebSocket(env.notificationsWsUrl);
    wsRef.current = ws;

    ws.addEventListener('message', (event) => {
      try {
        const payload = JSON.parse(event.data as string) as Notification;
        setNotifications((prev) => [payload, ...prev]);
        setUnreadCount((prev) => prev + 1);
      } catch {
        // Ignore malformed payloads.
      }
    });

    ws.addEventListener('close', () => {
      wsRef.current = null;
    });

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, []);

  const value = useMemo(
    () => ({
      notifications,
      unreadCount,
      isLoading,
      error,
      refresh,
      markAsRead,
      markAllAsRead,
    }),
    [notifications, unreadCount, isLoading, error, refresh, markAsRead, markAllAsRead]
  );

  return value;
}
