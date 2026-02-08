import { NextResponse } from 'next/server';
import { listRequests } from '../data';

export async function POST(req: Request) {
  const body = (await req.json()) as { dataSubject: string };
  return NextResponse.json(listRequests(body.dataSubject));
}
