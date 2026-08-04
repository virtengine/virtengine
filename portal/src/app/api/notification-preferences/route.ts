import { NextResponse } from 'next/server';
import {
  getNotificationPreferenceWorkflow,
  loadNotificationPreferences,
  NotificationPreferenceError,
  saveNotificationPreferences,
  type NotificationPreferencePersistenceAdapter,
  type NotificationPreferenceSessionResolver,
} from './workflow';

const statusFor = (error: NotificationPreferenceError) => {
  if (error.code === 'unauthenticated') return 401;
  if (error.code === 'invalid_request') return 400;
  if (error.code === 'invalid_receipt') return 502;
  return 503;
};

const failure = (error: unknown) => {
  const workflowError =
    error instanceof NotificationPreferenceError
      ? error
      : new NotificationPreferenceError('persistence_failed');
  return NextResponse.json({ error: workflowError.code }, { status: statusFor(workflowError) });
};

export const createNotificationPreferenceHandlers = (
  resolver?: NotificationPreferenceSessionResolver,
  adapter?: NotificationPreferencePersistenceAdapter
) => ({
  GET: async (request: Request) => {
    try {
      return NextResponse.json(await loadNotificationPreferences(request, resolver, adapter));
    } catch (error) {
      return failure(error);
    }
  },
  PUT: async (request: Request) => {
    let body: unknown;
    try {
      body = await request.json();
    } catch {
      return failure(new NotificationPreferenceError('invalid_request'));
    }
    try {
      return NextResponse.json(await saveNotificationPreferences(request, body, resolver, adapter));
    } catch (error) {
      return failure(error);
    }
  },
});

export const GET = (request: Request) => {
  const workflow = getNotificationPreferenceWorkflow();
  return createNotificationPreferenceHandlers(workflow.resolver, workflow.adapter).GET(request);
};

export const PUT = (request: Request) => {
  const workflow = getNotificationPreferenceWorkflow();
  return createNotificationPreferenceHandlers(workflow.resolver, workflow.adapter).PUT(request);
};
