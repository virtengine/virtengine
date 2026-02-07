import { NextResponse } from 'next/server';
import { notificationsData } from './data';

export function GET() {
  const unreadCount = notificationsData.filter((notif) => !notif.readAt).length;
  return NextResponse.json({ notifications: notificationsData, unreadCount });
}
