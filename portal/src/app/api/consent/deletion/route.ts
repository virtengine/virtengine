import { NextResponse } from 'next/server';
import {
  DeletionWorkflowError,
  getDeletionWorkflowAdapter,
  submitDeletionRequest,
  type DeletionWorkflowAdapter,
} from './workflow';

const statusFor = (error: DeletionWorkflowError) => {
  if (error.code === 'feature_unavailable' || error.code === 'workflow_failed') return 503;
  if (error.code === 'submission_in_progress') return 409;
  if (error.code === 'invalid_receipt') return 502;
  return 400;
};

export const createDeletionPostHandler = (adapter?: DeletionWorkflowAdapter) =>
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
      const record = await submitDeletionRequest(adapter, body.dataSubject ?? '');
      return NextResponse.json(record, { status: 202 });
    } catch (error) {
      const workflowError =
        error instanceof DeletionWorkflowError
          ? error
          : new DeletionWorkflowError('workflow_failed');
      return NextResponse.json({ error: workflowError.code }, { status: statusFor(workflowError) });
    }
  };

export const POST = (request: Request) =>
  createDeletionPostHandler(getDeletionWorkflowAdapter())(request);
