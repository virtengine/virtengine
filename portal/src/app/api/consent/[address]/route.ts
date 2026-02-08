import { NextResponse } from 'next/server';
import { getConsentSettings } from '../data';

export function GET(_req: Request, { params }: { params: { address: string } }) {
  return NextResponse.json(getConsentSettings(params.address));
}
