'use client';

import { formatRelativeTime } from '@/lib/utils';
import type { Notification } from '@/types/notifications';
import { cn } from '@/lib/utils';

const accentByType: Record<Notification['type'], string> = {
  veid_status: 'bg-indigo-500/10 text-indigo-700',
  order_update: 'bg-sky-500/10 text-sky-700',
  escrow_deposit: 'bg-emerald-500/10 text-emerald-700',
  security_alert: 'bg-rose-500/10 text-rose-700',
  provider_alert: 'bg-amber-500/10 text-amber-700',
};

export function NotificationItem({
  notification,
  onRead,
}: {
  notification: Notification;
  onRead: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onRead}
      className={cn(
        'flex w-full items-start gap-3 border-b border-border px-4 py-3 text-left transition-colors hover:bg-accent',
        notification.readAt ? 'opacity-70' : 'bg-background'
      )}
    >
      <span
        className={cn(
          'mt-1 h-2 w-2 flex-shrink-0 rounded-full',
          notification.readAt ? 'bg-muted-foreground/40' : 'bg-primary'
        )}
        aria-hidden="true"
      />
      <div className="flex-1 space-y-1">
        <div className="flex items-center justify-between gap-2">
          <span className="text-sm font-semibold">{notification.title}</span>
          <span className="text-xs text-muted-foreground">
            {formatRelativeTime(notification.createdAt)}
          </span>
        </div>
        <p className="text-sm text-muted-foreground">{notification.body}</p>
        <span
          className={cn(
            'inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium',
            accentByType[notification.type]
          )}
        >
          {notification.type.replace('_', ' ')}
        </span>
      </div>
    </button>
  );
}
