import { getNotificationPreferenceWorkflow } from './workflow';
import { createNotificationPreferenceHandlers } from './handlers';

export const GET = (request: Request) => {
  const workflow = getNotificationPreferenceWorkflow();
  return createNotificationPreferenceHandlers(workflow.resolver, workflow.adapter).GET(request);
};

export const PUT = (request: Request) => {
  const workflow = getNotificationPreferenceWorkflow();
  return createNotificationPreferenceHandlers(workflow.resolver, workflow.adapter).PUT(request);
};
