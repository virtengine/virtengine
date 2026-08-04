import { describe, expect, it, vi } from 'vitest';
import { createDeletionPostHandler } from './route';
import { createRequestsPostHandler } from '../requests/route';
import {
  DeletionWorkflowError,
  listDeletionRequests,
  submitDeletionRequest,
  type DeletionWorkflowAdapter,
} from './workflow';

const subject = 'virtengine1subject';
const record = {
  id: 'deletion-authoritative-1',
  dataSubject: subject,
  requestedAt: '2026-08-04T00:00:00.000Z',
  status: 'pending',
};

const request = () =>
  new Request('http://localhost/api/consent/deletion', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ dataSubject: subject }),
  });

const adapter = (overrides: Partial<DeletionWorkflowAdapter> = {}): DeletionWorkflowAdapter => ({
  submitDeletion: vi.fn((submission) =>
    Promise.resolve({ ...record, idempotencyKey: submission.idempotencyKey })
  ),
  listDeletions: vi.fn().mockResolvedValue([record]),
  ...overrides,
});

describe('consent deletion workflow', () => {
  it('returns unavailable without a durable workflow adapter', async () => {
    const response = await createDeletionPostHandler()(request());
    expect(response.status).toBe(503);
    await expect(response.json()).resolves.toEqual({ error: 'feature_unavailable' });
  });

  it('accepts only an exact authoritative server record', async () => {
    const submitDeletion = vi.fn((submission) =>
      Promise.resolve({ ...record, idempotencyKey: submission.idempotencyKey })
    );
    const workflow = adapter({ submitDeletion });
    const response = await createDeletionPostHandler(workflow)(request());
    expect(response.status).toBe(202);
    await expect(response.json()).resolves.toEqual(record);
    expect(submitDeletion).toHaveBeenCalledWith(
      expect.objectContaining({ dataSubject: subject, idempotencyKey: expect.any(String) })
    );
  });

  it('rejects mismatched or malformed workflow evidence', async () => {
    await expect(
      submitDeletionRequest(
        adapter({ submitDeletion: vi.fn().mockResolvedValue({ ...record, dataSubject: 'other' }) }),
        subject
      )
    ).rejects.toBeInstanceOf(DeletionWorkflowError);
  });

  it('blocks duplicate submissions for the same subject', async () => {
    let resolveSubmission!: (value: unknown) => void;
    let idempotencyKey = '';
    const submitDeletion = vi.fn(
      (submission) =>
        new Promise((resolve) => {
          idempotencyKey = submission.idempotencyKey;
          resolveSubmission = resolve;
        })
    );
    const workflow = adapter({ submitDeletion });
    const pending = submitDeletionRequest(workflow, subject);

    await expect(submitDeletionRequest(workflow, subject)).rejects.toMatchObject({
      code: 'submission_in_progress',
    });
    resolveSubmission({ ...record, idempotencyKey });
    await expect(pending).resolves.toEqual(record);
    expect(submitDeletion).toHaveBeenCalledTimes(1);
  });

  it('materializes an immutable authoritative deletion list', async () => {
    const source = { ...record };
    const records = await listDeletionRequests(
      adapter({ listDeletions: vi.fn().mockResolvedValue([source]) }),
      subject
    );
    source.id = 'mutated';
    expect(records[0].id).toBe(record.id);
    expect(Object.isFrozen(records)).toBe(true);
    expect(Object.isFrozen(records[0])).toBe(true);
  });

  it('accepts strict terminal history records', async () => {
    const completed = {
      ...record,
      status: 'complete',
      completedAt: '2026-08-04T01:00:00.000Z',
    };
    await expect(
      listDeletionRequests(
        adapter({ listDeletions: vi.fn().mockResolvedValue([completed]) }),
        subject
      )
    ).resolves.toEqual([completed]);
  });

  it('rejects contradictory status-specific fields', async () => {
    await expect(
      listDeletionRequests(
        adapter({
          listDeletions: vi
            .fn()
            .mockResolvedValue([{ ...record, status: 'blocked', blockers: [] }]),
        }),
        subject
      )
    ).rejects.toMatchObject({ code: 'invalid_receipt' });
    await expect(
      listDeletionRequests(
        adapter({
          listDeletions: vi.fn().mockResolvedValue([{ ...record, error: 'contradictory' }]),
        }),
        subject
      )
    ).rejects.toMatchObject({ code: 'invalid_receipt' });
  });

  it('maps malformed JSON and malformed adapter evidence separately', async () => {
    const malformedRequest = new Request('http://localhost/api/consent/deletion', {
      method: 'POST',
      body: '{',
    });
    expect((await createDeletionPostHandler(adapter())(malformedRequest)).status).toBe(400);
    const wrongShape = new Request('http://localhost/api/consent/deletion', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dataSubject: 123 }),
    });
    expect((await createDeletionPostHandler(adapter())(wrongShape)).status).toBe(400);
    expect(
      (
        await createDeletionPostHandler(
          adapter({ submitDeletion: vi.fn().mockResolvedValue({ id: 'bad' }) })
        )(request())
      ).status
    ).toBe(502);
  });

  it('fails request history closed without a durable adapter', async () => {
    const response = await createRequestsPostHandler()(request());
    expect(response.status).toBe(503);
    await expect(response.json()).resolves.toEqual({ error: 'feature_unavailable' });
  });
});
