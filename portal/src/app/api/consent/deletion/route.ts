import { NextResponse } from 'next/server';
import { requestDeletion } from '../data';

export async function POST(req: Request) {
  const body = (await req.json()) as { dataSubject: string };
  const record = requestDeletion(body.dataSubject);
  return NextResponse.json(record);
}
