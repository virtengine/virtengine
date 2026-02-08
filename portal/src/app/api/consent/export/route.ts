import { NextResponse } from 'next/server';
import { requestExport } from '../data';

export async function POST(req: Request) {
  const body = (await req.json()) as { dataSubject: string; format?: 'json' | 'csv' };
  const record = requestExport(body.dataSubject, body.format ?? 'json');
  return NextResponse.json(record);
}
