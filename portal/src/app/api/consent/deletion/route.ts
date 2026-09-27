import { getDeletionWorkflowAdapter } from './workflow';
import { createDeletionPostHandler } from './handlers';

export const POST = (request: Request) =>
  createDeletionPostHandler(getDeletionWorkflowAdapter())(request);
