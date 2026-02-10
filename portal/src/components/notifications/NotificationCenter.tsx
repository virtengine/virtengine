'use client';

import { Bell } from 'lucide-react';
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from '@/components/ui/Dropdown';
import { Button } from '@/components/ui/Button';
import { NotificationItem } from './NotificationItem';
import { useNotifications } from './hooks/useNotifications';
import { useTranslation } from 'react-i18next';

export function NotificationCenter() {
  const { t } = useTranslation();
  const { notifications, unreadCount, isLoading, markAsRead, markAllAsRead } = useNotifications();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="relative rounded-full border border-border bg-background p-2 text-muted-foreground hover:bg-accent hover:text-foreground"
          aria-label={t('Open notifications')}
        >
          <Bell className="h-4 w-4" aria-hidden="true" />
          {unreadCount > 0 && (
            <span className="absolute -right-1 -top-1 flex h-5 min-w-[20px] items-center justify-center rounded-full bg-destructive px-1 text-[11px] font-semibold text-destructive-foreground">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent sideOffset={8} className="w-96 p-0">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <div>
            <p className="text-sm font-semibold">{t('Notifications')}</p>
            <p className="text-xs text-muted-foreground">
              {t('{{count}} unread · {{total}} total', {
                count: unreadCount,
                total: notifications.length,
              })}
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={markAllAsRead} disabled={unreadCount === 0}>
            {t('Mark all read')}
          </Button>
        </div>
        <div className="max-h-[420px] overflow-y-auto">
          {isLoading ? (
            <p className="px-4 py-6 text-center text-sm text-muted-foreground">
              {t('Loading notifications…')}
            </p>
          ) : notifications.length === 0 ? (
            <p className="px-4 py-6 text-center text-sm text-muted-foreground">
              {t('No notifications yet.')}
            </p>
          ) : (
            notifications.map((notif) => (
              <NotificationItem
                key={notif.id}
                notification={notif}
                onRead={() => markAsRead([notif.id])}
              />
            ))
          )}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
