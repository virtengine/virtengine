/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { formatRelativeTime } from '@/lib/utils';
import type { DashboardNotification } from '@/types/customer';
import { NOTIFICATION_SEVERITY_VARIANT } from '@/types/customer';
import { useTranslation } from 'react-i18next';

interface NotificationsFeedProps {
  notifications: DashboardNotification[];
  onMarkRead: (id: string) => void;
  onDismiss: (id: string) => void;
}

function NotificationRow({
  notification,
  onMarkRead,
  onDismiss,
}: {
  notification: DashboardNotification;
  onMarkRead: (id: string) => void;
  onDismiss: (id: string) => void;
}) {
  const { t } = useTranslation();
  return (
    <div
      className={`flex items-start gap-3 rounded-md px-3 py-2 ${
        !notification.read ? 'bg-muted/50' : ''
      }`}
    >
      <div className="flex-1 space-y-0.5">
        <div className="flex items-center gap-2">
          <Badge variant={NOTIFICATION_SEVERITY_VARIANT[notification.severity]} size="sm">
            {t(notification.severity)}
          </Badge>
          <span className="text-sm font-medium">{notification.title}</span>
        </div>
        <p className="text-xs text-muted-foreground">{notification.message}</p>
        <p className="text-xs text-muted-foreground">
          {formatRelativeTime(notification.createdAt)}
        </p>
      </div>
      <div className="flex shrink-0 gap-1">
        {!notification.read && (
          <Button variant="ghost" size="icon-sm" onClick={() => onMarkRead(notification.id)}>
            <span className="sr-only">{t('Mark read')}</span>
            <span aria-hidden="true" className="text-xs">
              ✓
            </span>
          </Button>
        )}
        <Button variant="ghost" size="icon-sm" onClick={() => onDismiss(notification.id)}>
          <span className="sr-only">{t('Dismiss')}</span>
          <span aria-hidden="true" className="text-xs">
            ✕
          </span>
        </Button>
      </div>
    </div>
  );
}

export function NotificationsFeed({
  notifications,
  onMarkRead,
  onDismiss,
}: NotificationsFeedProps) {
  const { t } = useTranslation();
  if (notifications.length === 0) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base">{t('Notifications')}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-center text-sm text-muted-foreground">{t('No notifications.')}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{t('Notifications')}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-1 p-2">
        {notifications.map((n) => (
          <NotificationRow
            key={n.id}
            notification={n}
            onMarkRead={onMarkRead}
            onDismiss={onDismiss}
          />
        ))}
      </CardContent>
    </Card>
  );
}
