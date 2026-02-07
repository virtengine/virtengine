import { NextResponse } from 'next/server';
import { markNotificationsRead } from '@/lib/notifications/store';

export async function POST(request: Request) {
  const body = await request.json().catch(() => ({}));
  const ids = Array.isArray(body?.ids) ? body.ids : [];
  markNotificationsRead(ids);
  return NextResponse.json({ ok: true });
}
