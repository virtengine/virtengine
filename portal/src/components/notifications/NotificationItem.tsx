'use client';

import type { Notification } from '@/types/notifications';
import { cn, formatRelativeTime } from '@/lib/utils';

interface NotificationItemProps {
  notification: Notification;
  onRead: () => void;
}

export function NotificationItem({ notification, onRead }: NotificationItemProps) {
  const unread = !notification.readAt;

  return (
    <button
      type="button"
      onClick={onRead}
      className={cn(
        'w-full rounded-lg border border-border px-3 py-3 text-left transition hover:bg-accent',
        unread && 'border-primary/40 bg-primary/5'
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <p
            className={cn(
              'text-sm font-medium',
              unread ? 'text-foreground' : 'text-muted-foreground'
            )}
          >
            {notification.title}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">{notification.body}</p>
        </div>
        {unread && <span className="mt-1 h-2 w-2 rounded-full bg-primary" />}
      </div>
      <p className="mt-2 text-xs text-muted-foreground">
        {formatRelativeTime(notification.createdAt)}
      </p>
    </button>
  );
}
