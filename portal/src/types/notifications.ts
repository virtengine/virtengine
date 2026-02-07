export type NotificationType =
  | 'veid_status'
  | 'order_update'
  | 'escrow_deposit'
  | 'security_alert'
  | 'provider_alert';

export type NotificationChannel = 'push' | 'email' | 'in_app';

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  body: string;
  createdAt: string;
  readAt: string | null;
  data?: Record<string, string>;
}

export interface QuietHours {
  enabled: boolean;
  startHour: number;
  endHour: number;
  timezone: string;
}

export interface NotificationPreferences {
  channels: Record<NotificationType, NotificationChannel[]>;
  quietHours: QuietHours;
  digestEnabled: boolean;
  digestTime: string;
}
