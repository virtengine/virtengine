export type NotificationType =
  | 'veid_status'
  | 'order_update'
  | 'escrow_deposit'
  | 'security_alert'
  | 'provider_alert';

export type NotificationChannel = 'push' | 'email' | 'in_app';

export interface Notification {
  id: string;
  userAddress: string;
  type: NotificationType;
  title: string;
  body: string;
  data?: Record<string, string>;
  createdAt: string;
  readAt?: string | null;
  channels: NotificationChannel[];
}

export interface QuietHours {
  enabled: boolean;
  startHour: number;
  endHour: number;
  timezone: string;
}

export interface NotificationPreferences {
  userAddress: string;
  channels: Record<NotificationType, NotificationChannel[]>;
  frequencies: Record<NotificationType, 'immediate' | 'digest'>;
  digestEnabled: boolean;
  digestTime: string;
  quietHours: QuietHours;
}
