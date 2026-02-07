import { NextResponse } from 'next/server';
import { listNotifications } from '@/lib/notifications/store';

export function GET() {
  const { notifications, unreadCount } = listNotifications();
  return NextResponse.json({ notifications, unreadCount });
}
