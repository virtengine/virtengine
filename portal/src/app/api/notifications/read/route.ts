import { NextResponse } from 'next/server';
import { notificationsData } from '../data';

export async function POST(req: Request) {
  const body = (await req.json()) as { ids?: string[] };
  const ids = new Set(body.ids ?? []);
  const now = new Date().toISOString();

  for (const notif of notificationsData) {
    if (ids.has(notif.id)) {
      notif.readAt = now;
    }
  }

  return NextResponse.json({ ok: true });
}
