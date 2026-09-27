import { getDeletionWorkflowAdapter } from '../deletion/workflow';
import { createRequestsPostHandler } from './handlers';

export const POST = (request: Request) =>
  createRequestsPostHandler(getDeletionWorkflowAdapter())(request);
