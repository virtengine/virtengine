import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';

// Mock the portal-adapter module
const mockRefresh = vi.fn().mockResolvedValue(undefined);
const mockState = {
  status: 'unknown' as string,
  score: null,
  completedScopes: [],
  isLoading: false,
  error: null,
};

vi.mock('@/lib/portal-adapter', () => ({
  useIdentity: () => ({
    state: mockState,
    actions: {
      refresh: mockRefresh,
      checkRequirements: vi.fn(),
    },
  }),
}));

import { useVeidPolling } from '@/features/veid/hooks/useVeidPolling';

describe('useVeidPolling', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    mockRefresh.mockClear();
    mockState.status = 'unknown';
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('does not poll when status is unknown', () => {
    mockState.status = 'unknown';
    const { result } = renderHook(() => useVeidPolling());

    expect(result.current.isPolling).toBe(false);
    expect(result.current.pollCount).toBe(0);
  });

  it('does not poll when status is verified', () => {
    mockState.status = 'verified';
    const { result } = renderHook(() => useVeidPolling());

    expect(result.current.isPolling).toBe(false);
  });

  it('polls when status is pending', () => {
    mockState.status = 'pending';
    const { result } = renderHook(() => useVeidPolling());

    expect(result.current.isPolling).toBe(true);
  });

  it('polls when status is processing', () => {
    mockState.status = 'processing';
    const { result } = renderHook(() => useVeidPolling());

    expect(result.current.isPolling).toBe(true);
  });

  it('calls refresh on interval when polling', async () => {
    mockState.status = 'pending';
    renderHook(() => useVeidPolling({ intervalMs: 5000 }));

    expect(mockRefresh).not.toHaveBeenCalled();

    await act(async () => {
      vi.advanceTimersByTime(5000);
    });

    expect(mockRefresh).toHaveBeenCalledTimes(1);

    await act(async () => {
      vi.advanceTimersByTime(5000);
    });

    expect(mockRefresh).toHaveBeenCalledTimes(2);
  });

  it('does not poll when enabled is false', () => {
    mockState.status = 'pending';
    const { result } = renderHook(() => useVeidPolling({ enabled: false }));

    expect(result.current.isPolling).toBe(false);
  });

  it('pollNow triggers immediate refresh', async () => {
    mockState.status = 'unknown';
    const { result } = renderHook(() => useVeidPolling());

    await act(async () => {
      await result.current.pollNow();
    });

    expect(mockRefresh).toHaveBeenCalledTimes(1);
    expect(result.current.pollCount).toBe(1);
    expect(result.current.lastPollAt).not.toBeNull();
  });

  it('stop stops polling', async () => {
    mockState.status = 'pending';
    const { result } = renderHook(() => useVeidPolling({ intervalMs: 5000 }));

    expect(result.current.isPolling).toBe(true);

    act(() => {
      result.current.stop();
    });

    await act(async () => {
      vi.advanceTimersByTime(10000);
    });

    expect(mockRefresh).not.toHaveBeenCalled();
  });

  it('handles refresh errors gracefully', async () => {
    mockState.status = 'unknown';
    mockRefresh.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useVeidPolling());

    await act(async () => {
      await result.current.pollNow();
    });

    expect(result.current.error).toBe('Network error');
  });

  it('clears error on successful poll after error', async () => {
    mockState.status = 'unknown';
    mockRefresh.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useVeidPolling());

    await act(async () => {
      await result.current.pollNow();
    });

    expect(result.current.error).toBe('Network error');

    await act(async () => {
      await result.current.pollNow();
    });

    expect(result.current.error).toBeNull();
    // pollCount is 1: error poll doesn't increment, only successful poll does
    expect(result.current.pollCount).toBe(1);
  });
});
