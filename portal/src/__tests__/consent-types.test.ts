import { describe, expect, it } from 'vitest';
import {
  normalizeConsentEvent,
  projectConsentEventToChain,
  type ConsentEventBase,
} from '@/types/consent';

const event: ConsentEventBase = {
  id: 'event-1',
  consentId: 'consent-1',
  dataSubject: 'virtengine1test',
  scopeId: 'veid.biometric',
  purpose: 'biometric_processing',
  eventType: 'granted',
  occurredAt: '2026-08-02T00:00:00.000Z',
};

describe('consent event provenance', () => {
  it('projects chain evidence only when committed and bound to the event and consent', () => {
    expect(
      projectConsentEventToChain(event, {
        blockHeight: 42,
        txHash: 'ABC123',
        chainId: 'virtengine-1',
        code: 0,
        eventId: event.id,
        consentId: event.consentId,
      })
    ).toMatchObject({ source: 'chain', chain: { blockHeight: 42, txHash: 'ABC123' } });
  });

  it.each([
    ['nonpositive height', { blockHeight: 0 }],
    ['fractional height', { blockHeight: 1.5 }],
    ['empty transaction hash', { txHash: '' }],
    ['empty chain ID', { chainId: '' }],
    ['failed transaction', { code: 7 }],
    ['wrong event', { eventId: 'event-other' }],
    ['wrong consent', { consentId: 'consent-other' }],
  ])('rejects %s', (_name, override) => {
    expect(
      projectConsentEventToChain(event, {
        blockHeight: 42,
        txHash: 'ABC123',
        chainId: 'virtengine-1',
        code: 0,
        eventId: event.id,
        consentId: event.consentId,
        ...override,
      })
    ).toBeNull();
  });

  it('downgrades a legacy bare block height to local provenance and strips it', () => {
    expect(normalizeConsentEvent({ ...event, blockHeight: 999 })).toEqual({
      ...event,
      source: 'local',
    });
  });
});