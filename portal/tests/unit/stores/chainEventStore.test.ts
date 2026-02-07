/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { describe, it, expect, beforeEach } from 'vitest';
import {
  useChainEventStore,
  selectIsConnected,
  selectRecentEvents,
  selectEventsByType,
} from '@/stores/chainEventStore';
import type { ChainEvent } from '@/types/chain-events';

function makeEvent(overrides: Partial<ChainEvent> = {}): ChainEvent {
  return {
    id: `evt-${Math.random().toString(36).slice(2)}`,
    type: 'order.created',
    blockHeight: 100,
    timestamp: new Date(),
    txHash: 'hash123',
    attributes: {},
    ...overrides,
  };
}

describe('chainEventStore', () => {
  beforeEach(() => {
    // Reset store to initial state.
    useChainEventStore.setState({
      connectionStatus: 'disconnected',
      events: [],
      isPolling: false,
      error: null,
    });
  });

  it('addEvent prepends events newest-first', () => {
    const store = useChainEventStore.getState();
    const evt1 = makeEvent({ id: 'evt-1', blockHeight: 10 });
    const evt2 = makeEvent({ id: 'evt-2', blockHeight: 20 });

    store.addEvent(evt1);
    store.addEvent(evt2);

    const { events } = useChainEventStore.getState();
    expect(events.length).toBe(2);
    expect(events[0]!.id).toBe('evt-2');
    expect(events[1]!.id).toBe('evt-1');
  });

  it('caps events at 100', () => {
    const store = useChainEventStore.getState();
    for (let i = 0; i < 110; i++) {
      store.addEvent(makeEvent({ id: `evt-${i}` }));
    }
    expect(useChainEventStore.getState().events.length).toBe(100);
  });

  it('clearEvents empties the event list', () => {
    const store = useChainEventStore.getState();
    store.addEvent(makeEvent());
    store.clearEvents();
    expect(useChainEventStore.getState().events.length).toBe(0);
  });

  it('enablePolling / disablePolling toggle isPolling', () => {
    const store = useChainEventStore.getState();
    expect(store.isPolling).toBe(false);
    store.enablePolling();
    expect(useChainEventStore.getState().isPolling).toBe(true);
    useChainEventStore.getState().disablePolling();
    expect(useChainEventStore.getState().isPolling).toBe(false);
  });

  it('clearError resets error to null', () => {
    useChainEventStore.setState({ error: 'test error' });
    useChainEventStore.getState().clearError();
    expect(useChainEventStore.getState().error).toBeNull();
  });

  // Selectors
  it('selectIsConnected returns true only when connected', () => {
    useChainEventStore.setState({ connectionStatus: 'disconnected' });
    expect(selectIsConnected(useChainEventStore.getState())).toBe(false);

    useChainEventStore.setState({ connectionStatus: 'connected' });
    expect(selectIsConnected(useChainEventStore.getState())).toBe(true);
  });

  it('selectRecentEvents returns limited events', () => {
    const store = useChainEventStore.getState();
    for (let i = 0; i < 10; i++) {
      store.addEvent(makeEvent({ id: `evt-${i}` }));
    }
    const selector = selectRecentEvents(3);
    const result = selector(useChainEventStore.getState());
    expect(result.length).toBe(3);
  });

  it('selectEventsByType filters correctly', () => {
    const store = useChainEventStore.getState();
    store.addEvent(makeEvent({ id: 'a', type: 'order.created' }));
    store.addEvent(makeEvent({ id: 'b', type: 'bid.created' }));
    store.addEvent(makeEvent({ id: 'c', type: 'order.created' }));

    const selector = selectEventsByType('order.created');
    const result = selector(useChainEventStore.getState());
    expect(result.length).toBe(2);
    expect(result.every((e) => e.type === 'order.created')).toBe(true);
  });
});
