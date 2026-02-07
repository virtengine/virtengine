import { NextResponse } from 'next/server';
import { getPreferences, updatePreferences } from '@/lib/notifications/store';

export function GET() {
  return NextResponse.json({ preferences: getPreferences() });
}

export async function POST(request: Request) {
  const body = await request.json().catch(() => ({}));
  const preferences = updatePreferences(body?.preferences ?? {});
  return NextResponse.json({ preferences });
}
