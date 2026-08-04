import crypto from 'crypto';
import type {
  NotificationChannel,
  NotificationPreferences,
  NotificationType,
} from '@/types/notifications';

export type NotificationPreferenceSettings = Omit<NotificationPreferences, 'userAddress'>;

export interface NotificationPreferenceSession {
  readonly accountAddress: string;
}

export type NotificationPreferenceSessionResolver = (
  request: Request
) => Promise<NotificationPreferenceSession | null>;

export interface NotificationPreferenceSaveRequest {
  readonly accountAddress: string;
  readonly preferences: Readonly<NotificationPreferenceSettings>;
  readonly requestDigest: string;
  readonly idempotencyKey: string;
}

export interface NotificationPreferenceSaveReceipt {
  readonly status: 'committed';
  readonly operationId: string;
  readonly requestDigest: string;
  readonly idempotencyKey: string;
  readonly preferences: NotificationPreferences;
}

export interface NotificationPreferencePersistenceAdapter {
  get(accountAddress: string): Promise<unknown>;
  save(request: NotificationPreferenceSaveRequest): Promise<unknown>;
}

export class NotificationPreferenceError extends Error {
  constructor(
    public readonly code:
      | 'unauthenticated'
      | 'feature_unavailable'
      | 'invalid_request'
      | 'invalid_receipt'
      | 'persistence_failed'
  ) {
    super(code);
    this.name = 'NotificationPreferenceError';
  }
}

const TYPES: NotificationType[] = [
  'veid_status',
  'order_update',
  'escrow_deposit',
  'security_alert',
  'provider_alert',
];
const CHANNELS: NotificationChannel[] = ['push', 'email', 'in_app'];
const SETTINGS_KEYS = ['channels', 'digestEnabled', 'digestTime', 'frequencies', 'quietHours'];
const PREFERENCE_KEYS = ['userAddress', ...SETTINGS_KEYS];
const RECEIPT_KEYS = ['idempotencyKey', 'operationId', 'preferences', 'requestDigest', 'status'];

const canonicalText = (value: unknown, maxLength = 256): value is string =>
  typeof value === 'string' &&
  value.length > 0 &&
  value.length <= maxLength &&
  value.trim() === value;

const exactKeys = (value: Record<string, unknown>, keys: readonly string[]) => {
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return actual.length === expected.length && actual.every((key, index) => key === expected[index]);
};

const snapshotFields = (value: Record<string, unknown>, keys: readonly string[]) =>
  Object.fromEntries(keys.map((key) => [key, value[key]])) as Record<string, unknown>;

const snapshotArray = (value: unknown[]) => {
  const length = value.length;
  return Array.from({ length }, (_, index) => value[index]);
};

const materializeSettings = (value: unknown): Readonly<NotificationPreferenceSettings> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new NotificationPreferenceError('invalid_request');
  }
  const source = value as Record<string, unknown>;
  if (!exactKeys(source, SETTINGS_KEYS)) throw new NotificationPreferenceError('invalid_request');

  const settingsSnapshot = snapshotFields(source, SETTINGS_KEYS);
  const channelsValue = settingsSnapshot.channels;
  const frequenciesValue = settingsSnapshot.frequencies;
  const quietHoursValue = settingsSnapshot.quietHours;
  const digestEnabled = settingsSnapshot.digestEnabled;
  const digestTime = settingsSnapshot.digestTime;
  if (
    !channelsValue ||
    typeof channelsValue !== 'object' ||
    Array.isArray(channelsValue) ||
    !frequenciesValue ||
    typeof frequenciesValue !== 'object' ||
    Array.isArray(frequenciesValue) ||
    !quietHoursValue ||
    typeof quietHoursValue !== 'object' ||
    Array.isArray(quietHoursValue) ||
    typeof digestEnabled !== 'boolean' ||
    typeof digestTime !== 'string' ||
    !/^([01]\d|2[0-3]):[0-5]\d$/.test(digestTime)
  ) {
    throw new NotificationPreferenceError('invalid_request');
  }

  const channelsSource = channelsValue as Record<string, unknown>;
  const frequenciesSource = frequenciesValue as Record<string, unknown>;
  const quietHoursSource = quietHoursValue as Record<string, unknown>;
  if (
    !exactKeys(channelsSource, TYPES) ||
    !exactKeys(frequenciesSource, TYPES) ||
    !exactKeys(quietHoursSource, ['enabled', 'endHour', 'startHour', 'timezone'])
  ) {
    throw new NotificationPreferenceError('invalid_request');
  }

  const channelsSnapshot = snapshotFields(channelsSource, TYPES);
  const frequenciesSnapshot = snapshotFields(frequenciesSource, TYPES);
  const channels = {} as Record<NotificationType, NotificationChannel[]>;
  const frequencies = {} as Record<NotificationType, 'immediate' | 'digest'>;
  for (const type of TYPES) {
    const channelValue = channelsSnapshot[type];
    const frequency = frequenciesSnapshot[type];
    if (!Array.isArray(channelValue) || !['immediate', 'digest'].includes(frequency as string)) {
      throw new NotificationPreferenceError('invalid_request');
    }
    const channelSnapshot = snapshotArray(channelValue);
    if (
      channelSnapshot.some((channel) => !CHANNELS.includes(channel as NotificationChannel)) ||
      new Set(channelSnapshot).size !== channelSnapshot.length
    ) {
      throw new NotificationPreferenceError('invalid_request');
    }
    channels[type] = Object.freeze(channelSnapshot) as NotificationChannel[];
    frequencies[type] = frequency as 'immediate' | 'digest';
  }

  const quietHoursSnapshot = snapshotFields(quietHoursSource, [
    'enabled',
    'endHour',
    'startHour',
    'timezone',
  ]);
  const enabled = quietHoursSnapshot.enabled;
  const startHour = quietHoursSnapshot.startHour;
  const endHour = quietHoursSnapshot.endHour;
  const timezone = quietHoursSnapshot.timezone;
  if (
    typeof enabled !== 'boolean' ||
    !Number.isInteger(startHour) ||
    Number(startHour) < 0 ||
    Number(startHour) > 23 ||
    !Number.isInteger(endHour) ||
    Number(endHour) < 0 ||
    Number(endHour) > 23 ||
    !canonicalText(timezone, 64)
  ) {
    throw new NotificationPreferenceError('invalid_request');
  }

  return Object.freeze({
    channels: Object.freeze(channels),
    frequencies: Object.freeze(frequencies),
    digestEnabled,
    digestTime,
    quietHours: Object.freeze({
      enabled,
      startHour: Number(startHour),
      endHour: Number(endHour),
      timezone: String(timezone),
    }),
  });
};

const materializePreferences = (
  value: unknown,
  accountAddress: string
): Readonly<NotificationPreferences> => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  const source = value as Record<string, unknown>;
  if (!exactKeys(source, PREFERENCE_KEYS)) {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  const preferenceSnapshot = snapshotFields(source, PREFERENCE_KEYS);
  const receivedAddress = preferenceSnapshot.userAddress;
  if (receivedAddress !== accountAddress) throw new NotificationPreferenceError('invalid_receipt');
  let settings: Readonly<NotificationPreferenceSettings>;
  try {
    settings = materializeSettings(snapshotFields(preferenceSnapshot, SETTINGS_KEYS));
  } catch {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  return Object.freeze({ userAddress: accountAddress, ...settings });
};

const normalizeSubject = (session: NotificationPreferenceSession | null): string => {
  const accountAddress = session?.accountAddress;
  if (!canonicalText(accountAddress)) throw new NotificationPreferenceError('unauthenticated');
  return accountAddress;
};

const digestSettings = (accountAddress: string, settings: NotificationPreferenceSettings) =>
  crypto
    .createHash('sha256')
    .update(JSON.stringify({ accountAddress, preferences: settings }))
    .digest('hex');

let configuredResolver: NotificationPreferenceSessionResolver | undefined;
let configuredAdapter: NotificationPreferencePersistenceAdapter | undefined;

export const configureNotificationPreferenceWorkflow = (
  resolver: NotificationPreferenceSessionResolver | undefined,
  adapter: NotificationPreferencePersistenceAdapter | undefined
) => {
  configuredResolver = resolver;
  configuredAdapter = adapter;
};

export const getNotificationPreferenceWorkflow = () => ({
  resolver: configuredResolver,
  adapter: configuredAdapter,
});

export const loadNotificationPreferences = async (
  request: Request,
  resolver: NotificationPreferenceSessionResolver | undefined,
  adapter: NotificationPreferencePersistenceAdapter | undefined
): Promise<Readonly<NotificationPreferences>> => {
  if (!resolver) throw new NotificationPreferenceError('unauthenticated');
  const accountAddress = normalizeSubject(await resolver(request));
  if (!adapter) throw new NotificationPreferenceError('feature_unavailable');
  try {
    return materializePreferences(await adapter.get(accountAddress), accountAddress);
  } catch (error) {
    if (error instanceof NotificationPreferenceError) throw error;
    throw new NotificationPreferenceError('persistence_failed');
  }
};

export const saveNotificationPreferences = async (
  request: Request,
  value: unknown,
  resolver: NotificationPreferenceSessionResolver | undefined,
  adapter: NotificationPreferencePersistenceAdapter | undefined
): Promise<Readonly<NotificationPreferences>> => {
  if (!resolver) throw new NotificationPreferenceError('unauthenticated');
  const accountAddress = normalizeSubject(await resolver(request));
  if (!adapter) throw new NotificationPreferenceError('feature_unavailable');
  const preferences = materializeSettings(value);
  const requestDigest = digestSettings(accountAddress, preferences);
  let receiptValue: unknown;
  try {
    receiptValue = await adapter.save(
      Object.freeze({
        accountAddress,
        preferences,
        requestDigest,
        idempotencyKey: requestDigest,
      })
    );
  } catch {
    throw new NotificationPreferenceError('persistence_failed');
  }
  if (!receiptValue || typeof receiptValue !== 'object' || Array.isArray(receiptValue)) {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  const receipt = receiptValue as Record<string, unknown>;
  if (!exactKeys(receipt, RECEIPT_KEYS)) {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  const receiptSnapshot = snapshotFields(receipt, RECEIPT_KEYS);
  const status = receiptSnapshot.status;
  const operationId = receiptSnapshot.operationId;
  const receiptDigest = receiptSnapshot.requestDigest;
  const idempotencyKey = receiptSnapshot.idempotencyKey;
  const savedPreferences = receiptSnapshot.preferences;
  if (
    status !== 'committed' ||
    !canonicalText(operationId) ||
    receiptDigest !== requestDigest ||
    idempotencyKey !== requestDigest
  ) {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  const materialized = materializePreferences(savedPreferences, accountAddress);
  if (
    JSON.stringify(materialized) !== JSON.stringify({ userAddress: accountAddress, ...preferences })
  ) {
    throw new NotificationPreferenceError('invalid_receipt');
  }
  return materialized;
};
