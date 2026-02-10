import { NextResponse } from 'next/server';
import { getPreferences, setPreferences } from '../notifications/data';
import type { NotificationPreferences } from '@/types/notifications';

export function GET() {
  return NextResponse.json(getPreferences());
}

export async function PUT(req: Request) {
  const body = (await req.json()) as NotificationPreferences;
  const merged = {
    ...getPreferences(),
    ...body,
  };
  setPreferences(merged);
  return NextResponse.json(merged);
}
