import type { Notification, NotificationPreferences } from '@/types/notifications';

const now = new Date();

export const notificationsData: Notification[] = [
  {
    id: 'notif-001',
    userAddress: 'virtengine1demo',
    type: 'order_update',
    title: 'Order deployed',
    body: 'Order ve-order-1002 is now running on CloudCore.',
    data: { orderId: 've-order-1002' },
    createdAt: new Date(now.getTime() - 5 * 60 * 1000).toISOString(),
    readAt: null,
    channels: ['push', 'in_app'],
  },
  {
    id: 'notif-002',
    userAddress: 'virtengine1demo',
    type: 'security_alert',
    title: 'New login detected',
    body: 'A new login from Chicago, IL was detected.',
    data: { ip: '203.0.113.18' },
    createdAt: new Date(now.getTime() - 60 * 60 * 1000).toISOString(),
    readAt: null,
    channels: ['email', 'push', 'in_app'],
  },
  {
    id: 'notif-003',
    userAddress: 'virtengine1demo',
    type: 'escrow_deposit',
    title: 'Escrow funded',
    body: 'Your escrow balance was updated with 2,500 USDC.',
    data: { amount: '2500', currency: 'USDC' },
    createdAt: new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString(),
    readAt: new Date(now.getTime() - 90 * 60 * 1000).toISOString(),
    channels: ['email', 'in_app'],
  },
];

let preferencesData: NotificationPreferences = {
  userAddress: 'virtengine1demo',
  channels: {
    veid_status: ['email', 'push', 'in_app'],
    order_update: ['push', 'in_app'],
    escrow_deposit: ['email', 'in_app'],
    security_alert: ['email', 'push', 'in_app'],
    provider_alert: ['push', 'in_app'],
  },
  frequencies: {
    veid_status: 'immediate',
    order_update: 'immediate',
    escrow_deposit: 'immediate',
    security_alert: 'immediate',
    provider_alert: 'digest',
  },
  digestEnabled: true,
  digestTime: '09:00',
  quietHours: {
    enabled: true,
    startHour: 22,
    endHour: 6,
    timezone: 'UTC',
  },
};

export function getPreferences(): NotificationPreferences {
  return preferencesData;
}

export function setPreferences(next: NotificationPreferences) {
  preferencesData = next;
}
