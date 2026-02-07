/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * useVeidPolling Hook
 * Polls verification status from the chain while processing is active.
 */

'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { useIdentity } from '@/lib/portal-adapter';
import { VERIFICATION_POLL_INTERVAL_MS } from '../constants';

export interface UseVeidPollingOptions {
  /** Enable/disable polling */
  enabled?: boolean;
  /** Custom polling interval in ms */
  intervalMs?: number;
}

export interface UseVeidPollingReturn {
  /** Whether polling is currently active */
  isPolling: boolean;
  /** Number of poll cycles completed */
  pollCount: number;
  /** Last poll timestamp */
  lastPollAt: number | null;
  /** Any polling error */
  error: string | null;
  /** Manually trigger a poll */
  pollNow: () => Promise<void>;
  /** Stop polling */
  stop: () => void;
}

export function useVeidPolling(options: UseVeidPollingOptions = {}): UseVeidPollingReturn {
  const { enabled = true, intervalMs = VERIFICATION_POLL_INTERVAL_MS } = options;
  const { state, actions } = useIdentity();

  const [pollCount, setPollCount] = useState(0);
  const [lastPollAt, setLastPollAt] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const stoppedRef = useRef(false);

  const shouldPoll = enabled && (state.status === 'pending' || state.status === 'processing');

  const pollNow = useCallback(async () => {
    try {
      setError(null);
      await actions.refresh();
      setPollCount((prev) => prev + 1);
      setLastPollAt(Date.now());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to poll verification status');
    }
  }, [actions]);

  const stop = useCallback(() => {
    stoppedRef.current = true;
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (!shouldPoll || stoppedRef.current) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    intervalRef.current = setInterval(() => void pollNow(), intervalMs);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [shouldPoll, intervalMs, pollNow]);

  // Reset stopped flag when enabled changes
  useEffect(() => {
    if (enabled) {
      stoppedRef.current = false;
    }
  }, [enabled]);

  return {
    isPolling: shouldPoll && !stoppedRef.current,
    pollCount,
    lastPollAt,
    error,
    pollNow,
    stop,
  };
}
