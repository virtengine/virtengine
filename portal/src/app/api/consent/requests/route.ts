import { NextResponse } from 'next/server';
import { listRequests } from '../data';
import {
  DeletionWorkflowError,
  getDeletionWorkflowAdapter,
  listDeletionRequests,
  type DeletionWorkflowAdapter,
} from '../deletion/workflow';

export const createRequestsPostHandler = (adapter?: DeletionWorkflowAdapter) =>
  async function POST(req: Request) {
    try {
      let body: { dataSubject?: string };
      try {
        body = (await req.json()) as { dataSubject?: string };
      } catch {
        return NextResponse.json({ error: 'invalid_subject' }, { status: 400 });
      }
      if (!body || typeof body.dataSubject !== 'string') {
        return NextResponse.json({ error: 'invalid_subject' }, { status: 400 });
      }
      const dataSubject = body.dataSubject ?? '';
      const local = listRequests(dataSubject);
      const deletions = await listDeletionRequests(adapter, dataSubject);
      return NextResponse.json({ exports: local.exports, deletions });
    } catch (error) {
      const code = error instanceof DeletionWorkflowError ? error.code : 'workflow_failed';
      const status = code === 'invalid_subject' ? 400 : code === 'invalid_receipt' ? 502 : 503;
      return NextResponse.json({ error: code }, { status });
    }
  };

export const POST = (request: Request) =>
  createRequestsPostHandler(getDeletionWorkflowAdapter())(request);
