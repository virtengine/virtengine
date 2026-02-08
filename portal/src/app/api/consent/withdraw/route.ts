import { NextResponse } from 'next/server';
import { withdrawConsent } from '../data';

export async function POST(req: Request) {
  const body = (await req.json()) as { dataSubject: string; consentId: string };
  const record = withdrawConsent(body);
  return NextResponse.json(record);
}
