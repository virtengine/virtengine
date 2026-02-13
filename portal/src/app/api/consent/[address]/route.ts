import { NextResponse } from 'next/server';
import { getConsentSettings } from '../data';

export const dynamic = 'force-static';
export const dynamicParams = false;

export function generateStaticParams() {
  return [{ address: 'virtengine1demo' }];
}

export function GET(_req: Request, { params }: { params: { address: string } }) {
  return NextResponse.json(getConsentSettings(params.address));
}
