import crypto from 'crypto';
import { describe, expect, it, vi } from 'vitest';
import { createNotificationPreferenceHandlers } from './route';
import type {
  NotificationPreferencePersistenceAdapter,
  NotificationPreferenceSaveRequest,
  NotificationPreferenceSessionResolver,
  NotificationPreferenceSettings,
} from './workflow';

const accountAddress = 'virtengine1session';
const otherAddress = 'virtengine1other';
const settings: NotificationPreferenceSettings = {
  channels: {
    veid_status: ['email', 'push'],
    order_update: ['push', 'in_app'],
    escrow_deposit: ['email'],
    security_alert: ['email', 'push', 'in_app'],
    provider_alert: ['in_app'],
  },
  frequencies: {
    veid_status: 'immediate',
    order_update: 'immediate',
    escrow_deposit: 'digest',
    security_alert: 'immediate',
    provider_alert: 'digest',
  },
  digestEnabled: true,
  digestTime: '09:30',
  quietHours: {
    enabled: true,
    startHour: 22,
    endHour: 6,
    timezone: 'UTC',
  },
};
const preferences = { userAddress: accountAddress, ...settings };

const resolver: NotificationPreferenceSessionResolver = vi
  .fn()
  .mockResolvedValue({ accountAddress });

const request = (method: 'GET' | 'PUT', body?: string) =>
  new Request('http://localhost/api/notification-preferences', {
    method,
    headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
    body,
  });

const jsonRequest = (body: unknown) => request('PUT', JSON.stringify(body));

const adapterWith = (
  getMock = vi.fn<(address: string) => Promise<unknown>>().mockResolvedValue(preferences),
  saveMock = vi
    .fn<(submission: NotificationPreferenceSaveRequest) => Promise<unknown>>()
    .mockRejectedValue(new Error('save not configured'))
): NotificationPreferencePersistenceAdapter => ({
  get: (address) => getMock(address),
  save: (submission) => saveMock(submission),
});

const digestFor = (value: NotificationPreferenceSettings) =>
  crypto
    .createHash('sha256')
    .update(JSON.stringify({ accountAddress, preferences: value }))
    .digest('hex');

const committedReceipt = (submission: NotificationPreferenceSaveRequest) => ({
  status: 'committed',
  operationId: 'notification-preferences-1',
  requestDigest: submission.requestDigest,
  idempotencyKey: submission.idempotencyKey,
  preferences: { userAddress: submission.accountAddress, ...submission.preferences },
});

describe('notification preference route', () => {
  it.each(['GET', 'PUT'] as const)(
    'returns 401 for %s without a session resolver',
    async (method) => {
      const handlers = createNotificationPreferenceHandlers(undefined, adapterWith());
      const response = await handlers[method](
        method === 'GET' ? request('GET') : jsonRequest(settings)
      );

      expect(response.status).toBe(401);
      await expect(response.json()).resolves.toEqual({ error: 'unauthenticated' });
    }
  );

  it.each(['GET', 'PUT'] as const)(
    'returns 503 for authenticated %s without an adapter',
    async (method) => {
      const handlers = createNotificationPreferenceHandlers(resolver);
      const response = await handlers[method](
        method === 'GET' ? request('GET') : jsonRequest(settings)
      );

      expect(response.status).toBe(503);
      await expect(response.json()).resolves.toEqual({ error: 'feature_unavailable' });
    }
  );

  it.each([
    ['malformed JSON', '{'],
    ['non-object settings', JSON.stringify(null)],
    ['invalid settings', JSON.stringify({ ...settings, digestTime: '25:00' })],
    ['a body-selected user address', JSON.stringify({ ...settings, userAddress: otherAddress })],
  ])('returns 400 for %s without calling the adapter', async (_name, body) => {
    const saveMock = vi.fn<(submission: NotificationPreferenceSaveRequest) => Promise<unknown>>();
    const workflowAdapter = adapterWith(undefined, saveMock);
    const response = await createNotificationPreferenceHandlers(resolver, workflowAdapter).PUT(
      request('PUT', body)
    );

    expect(response.status).toBe(400);
    await expect(response.json()).resolves.toEqual({ error: 'invalid_request' });
    expect(saveMock).not.toHaveBeenCalled();
  });

  it('loads preferences for exactly the authenticated account', async () => {
    const getMock = vi.fn<(address: string) => Promise<unknown>>().mockResolvedValue(preferences);
    const response = await createNotificationPreferenceHandlers(resolver, adapterWith(getMock)).GET(
      request('GET')
    );

    expect(getMock).toHaveBeenCalledOnce();
    expect(getMock).toHaveBeenCalledWith(accountAddress);
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toEqual(preferences);
  });

  it('saves immutable settings bound only to the authenticated account and accepts exact evidence', async () => {
    const expectedDigest = digestFor(settings);
    const saveMock = vi.fn((submission: NotificationPreferenceSaveRequest) => {
      expect(Object.keys(submission).sort()).toEqual([
        'accountAddress',
        'idempotencyKey',
        'preferences',
        'requestDigest',
      ]);
      expect(submission.accountAddress).toBe(accountAddress);
      expect(submission.preferences).toEqual(settings);
      expect(Object.isFrozen(submission.preferences)).toBe(true);
      expect(Object.isFrozen(submission.preferences.channels)).toBe(true);
      expect(Object.isFrozen(submission.preferences.channels.veid_status)).toBe(true);
      expect(Object.isFrozen(submission.preferences.frequencies)).toBe(true);
      expect(Object.isFrozen(submission.preferences.quietHours)).toBe(true);
      expect(Object.isFrozen(submission)).toBe(true);
      expect(submission.requestDigest).toBe(expectedDigest);
      expect(submission.idempotencyKey).toBe(expectedDigest);
      return Promise.resolve(committedReceipt(submission));
    });

    const response = await createNotificationPreferenceHandlers(
      resolver,
      adapterWith(undefined, saveMock)
    ).PUT(jsonRequest(settings));

    expect(saveMock).toHaveBeenCalledOnce();
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toEqual(preferences);
  });

  it('snapshots getter-backed request and receipt fields exactly once', async () => {
    const reads = new Map<string, number>();
    const getterObject = (source: Record<string, unknown>, prefix: string) =>
      Object.fromEntries(
        Object.entries(source).map(([key, value]) => [
          key,
          {
            enumerable: true,
            get: () => {
              const name = `${prefix}.${key}`;
              reads.set(name, (reads.get(name) ?? 0) + 1);
              return value;
            },
          },
        ])
      );
    const getterArray = (source: readonly unknown[], prefix: string) =>
      new Proxy([...source], {
        get: (target, property, receiver) => {
          if (property === 'length' || (typeof property === 'string' && /^\d+$/.test(property))) {
            const name = `${prefix}.${String(property)}`;
            reads.set(name, (reads.get(name) ?? 0) + 1);
          }
          return Reflect.get(target, property, receiver);
        },
      });
    const requestChannels = Object.create(
      Object.prototype,
      getterObject(
        Object.fromEntries(
          Object.entries(settings.channels).map(([type, channels]) => [
            type,
            getterArray(channels, `request.channels.${type}`),
          ])
        ),
        'request.channels'
      )
    );
    const requestFrequencies = Object.create(
      Object.prototype,
      getterObject(settings.frequencies, 'request.frequencies')
    );
    const requestQuietHours = Object.create(
      Object.prototype,
      getterObject({ ...settings.quietHours }, 'request.quietHours')
    );
    const requestSettings = Object.create(
      Object.prototype,
      getterObject(
        {
          ...settings,
          channels: requestChannels,
          frequencies: requestFrequencies,
          quietHours: requestQuietHours,
        },
        'request'
      )
    );
    const saveMock = vi.fn((submission: NotificationPreferenceSaveRequest) => {
      const receiptPreferences = Object.create(
        Object.prototype,
        getterObject(
          { userAddress: submission.accountAddress, ...submission.preferences },
          'receipt.preferences'
        )
      );
      return Promise.resolve(
        Object.create(
          Object.prototype,
          getterObject(
            { ...committedReceipt(submission), preferences: receiptPreferences },
            'receipt'
          )
        )
      );
    });

    const response = await createNotificationPreferenceHandlers(
      resolver,
      adapterWith(undefined, saveMock)
    ).PUT({ json: () => Promise.resolve(requestSettings) } as Request);

    expect(response.status).toBe(200);
    expect([...reads.values()]).toEqual(expect.arrayContaining([1]));
    expect([...reads.entries()].filter(([, count]) => count !== 1)).toEqual([]);
  });

  it.each([
    ['loaded preferences', 'GET' as const],
    ['save receipt', 'PUT' as const],
    ['saved preferences', 'PUT' as const],
  ])('rejects additional keys in %s', async (location, method) => {
    const getMock = vi.fn().mockResolvedValue({ ...preferences, additional: true });
    const saveMock = vi.fn((submission: NotificationPreferenceSaveRequest) => {
      const receipt = committedReceipt(submission);
      if (location === 'save receipt') return Promise.resolve({ ...receipt, additional: true });
      return Promise.resolve({
        ...receipt,
        preferences: { ...receipt.preferences, additional: true },
      });
    });
    const handlers = createNotificationPreferenceHandlers(resolver, adapterWith(getMock, saveMock));
    const response = await handlers[method](
      method === 'GET' ? request('GET') : jsonRequest(settings)
    );

    expect(response.status).toBe(502);
    await expect(response.json()).resolves.toEqual({ error: 'invalid_receipt' });
  });

  it.each([
    [
      'a mismatched subject',
      (submission: NotificationPreferenceSaveRequest) => ({
        ...committedReceipt(submission),
        preferences: { userAddress: otherAddress, ...submission.preferences },
      }),
    ],
    [
      'mismatched preferences',
      (submission: NotificationPreferenceSaveRequest) => ({
        ...committedReceipt(submission),
        preferences: {
          userAddress: accountAddress,
          ...submission.preferences,
          digestEnabled: !submission.preferences.digestEnabled,
        },
      }),
    ],
    [
      'a mismatched digest',
      (submission: NotificationPreferenceSaveRequest) => ({
        ...committedReceipt(submission),
        requestDigest: '0'.repeat(64),
      }),
    ],
    ['a malformed receipt', () => ({ status: 'committed' })],
  ])('rejects %s with a non-2xx response', async (_name, receiptFor) => {
    const saveMock = vi.fn((submission: NotificationPreferenceSaveRequest) =>
      Promise.resolve(receiptFor(submission))
    );
    const response = await createNotificationPreferenceHandlers(
      resolver,
      adapterWith(undefined, saveMock)
    ).PUT(jsonRequest(settings));

    expect(response.status).toBe(502);
    await expect(response.json()).resolves.toEqual({ error: 'invalid_receipt' });
  });

  it('fails closed when the adapter rejects the save', async () => {
    const saveMock = vi
      .fn<(submission: NotificationPreferenceSaveRequest) => Promise<unknown>>()
      .mockRejectedValue(new Error('persistence unavailable'));
    const response = await createNotificationPreferenceHandlers(
      resolver,
      adapterWith(undefined, saveMock)
    ).PUT(jsonRequest(settings));

    expect(response.status).toBe(503);
    await expect(response.json()).resolves.toEqual({ error: 'persistence_failed' });
  });
});
