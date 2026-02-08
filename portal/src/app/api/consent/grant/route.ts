import { NextResponse } from 'next/server';
import { grantConsent } from '../data';
import type { ConsentPurpose } from '@/types/consent';

export async function POST(req: Request) {
  const body = (await req.json()) as {
    dataSubject: string;
    scopeId: string;
    purpose: ConsentPurpose;
    consentText: string;
    signature: string;
  };

  const record = grantConsent(body);
  return NextResponse.json(record);
}
