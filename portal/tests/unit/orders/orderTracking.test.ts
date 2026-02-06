/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  getOrderProgress,
  isOrderActive,
  isOrderTerminal,
  formatDuration,
  estimateTimeRemaining,
  ORDER_STATUS_CONFIG,
  STATUS_TO_TAB,
  ORDER_TAB_FILTERS,
} from '@/features/orders/tracking-types';
import type { OrderStatus } from '@/stores/orderStore';

describe('Order Tracking Types & Utilities', () => {
  // =========================================================================
  // getOrderProgress
  // =========================================================================

  describe('getOrderProgress', () => {
    it('returns 10 for pending', () => {
      expect(getOrderProgress('pending')).toBe(10);
    });

    it('returns 30 for matched', () => {
      expect(getOrderProgress('matched')).toBe(30);
    });

    it('returns 60 for deploying', () => {
      expect(getOrderProgress('deploying')).toBe(60);
    });

    it('returns 100 for running', () => {
      expect(getOrderProgress('running')).toBe(100);
    });

    it('returns 100 for paused', () => {
      expect(getOrderProgress('paused')).toBe(100);
    });

    it('returns 100 for completed', () => {
      expect(getOrderProgress('completed')).toBe(100);
    });

    it('returns 0 for failed', () => {
      expect(getOrderProgress('failed')).toBe(0);
    });
  });

  // =========================================================================
  // isOrderActive
  // =========================================================================

  describe('isOrderActive', () => {
    it('returns true for running', () => {
      expect(isOrderActive('running')).toBe(true);
    });

    it('returns true for deploying', () => {
      expect(isOrderActive('deploying')).toBe(true);
    });

    it('returns true for paused', () => {
      expect(isOrderActive('paused')).toBe(true);
    });

    it('returns false for pending', () => {
      expect(isOrderActive('pending')).toBe(false);
    });

    it('returns false for stopped', () => {
      expect(isOrderActive('stopped')).toBe(false);
    });

    it('returns false for completed', () => {
      expect(isOrderActive('completed')).toBe(false);
    });

    it('returns false for failed', () => {
      expect(isOrderActive('failed')).toBe(false);
    });
  });

  // =========================================================================
  // isOrderTerminal
  // =========================================================================

  describe('isOrderTerminal', () => {
    it('returns true for stopped', () => {
      expect(isOrderTerminal('stopped')).toBe(true);
    });

    it('returns true for completed', () => {
      expect(isOrderTerminal('completed')).toBe(true);
    });

    it('returns true for failed', () => {
      expect(isOrderTerminal('failed')).toBe(true);
    });

    it('returns false for running', () => {
      expect(isOrderTerminal('running')).toBe(false);
    });

    it('returns false for pending', () => {
      expect(isOrderTerminal('pending')).toBe(false);
    });

    it('returns false for deploying', () => {
      expect(isOrderTerminal('deploying')).toBe(false);
    });
  });

  // =========================================================================
  // formatDuration
  // =========================================================================

  describe('formatDuration', () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date('2025-02-06T12:00:00Z'));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('formats days and hours', () => {
      const start = new Date('2025-02-03T08:00:00Z').toISOString();
      const result = formatDuration(start);
      expect(result).toBe('3d 4h');
    });

    it('formats days only when no remaining hours', () => {
      const start = new Date('2025-02-04T12:00:00Z').toISOString();
      const result = formatDuration(start);
      expect(result).toBe('2d');
    });

    it('formats hours only for sub-day durations', () => {
      const start = new Date('2025-02-06T06:00:00Z').toISOString();
      const result = formatDuration(start);
      expect(result).toBe('6h');
    });

    it('formats minutes for very short durations', () => {
      const start = new Date('2025-02-06T11:30:00Z').toISOString();
      const result = formatDuration(start);
      expect(result).toBe('30m');
    });

    it('returns at least 1m for very recent timestamps', () => {
      const start = new Date('2025-02-06T11:59:50Z').toISOString();
      const result = formatDuration(start);
      expect(result).toBe('1m');
    });

    it('uses end date when provided', () => {
      const start = '2025-02-01T00:00:00Z';
      const end = '2025-02-03T12:00:00Z';
      const result = formatDuration(start, end);
      expect(result).toBe('2d 12h');
    });
  });

  // =========================================================================
  // estimateTimeRemaining
  // =========================================================================

  describe('estimateTimeRemaining', () => {
    beforeEach(() => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date('2025-02-06T12:00:00Z'));
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('returns null when no expiry', () => {
      expect(estimateTimeRemaining()).toBeNull();
      expect(estimateTimeRemaining(undefined)).toBeNull();
    });

    it('returns "Expired" for past dates', () => {
      expect(estimateTimeRemaining('2025-02-05T00:00:00Z')).toBe('Expired');
    });

    it('returns formatted time for future dates', () => {
      const result = estimateTimeRemaining('2025-02-08T12:00:00Z');
      expect(result).toBe('2d');
    });

    it('returns hours for same-day expiry', () => {
      const result = estimateTimeRemaining('2025-02-06T18:00:00Z');
      expect(result).toBe('6h');
    });
  });

  // =========================================================================
  // ORDER_STATUS_CONFIG
  // =========================================================================

  describe('ORDER_STATUS_CONFIG', () => {
    const allStatuses: OrderStatus[] = [
      'pending',
      'matched',
      'deploying',
      'running',
      'paused',
      'stopped',
      'completed',
      'failed',
    ];

    it('has config for every order status', () => {
      for (const status of allStatuses) {
        expect(ORDER_STATUS_CONFIG[status]).toBeDefined();
        expect(ORDER_STATUS_CONFIG[status].label).toBeTruthy();
        expect(ORDER_STATUS_CONFIG[status].variant).toBeTruthy();
        expect(ORDER_STATUS_CONFIG[status].icon).toBeTruthy();
      }
    });

    it('has correct variant for running', () => {
      expect(ORDER_STATUS_CONFIG.running.variant).toBe('success');
    });

    it('has correct variant for failed', () => {
      expect(ORDER_STATUS_CONFIG.failed.variant).toBe('destructive');
    });

    it('has correct variant for pending', () => {
      expect(ORDER_STATUS_CONFIG.pending.variant).toBe('warning');
    });
  });

  // =========================================================================
  // STATUS_TO_TAB
  // =========================================================================

  describe('STATUS_TO_TAB', () => {
    it('maps running to active tab', () => {
      expect(STATUS_TO_TAB.running).toBe('active');
    });

    it('maps deploying to active tab', () => {
      expect(STATUS_TO_TAB.deploying).toBe('active');
    });

    it('maps paused to active tab', () => {
      expect(STATUS_TO_TAB.paused).toBe('active');
    });

    it('maps pending to pending tab', () => {
      expect(STATUS_TO_TAB.pending).toBe('pending');
    });

    it('maps matched to pending tab', () => {
      expect(STATUS_TO_TAB.matched).toBe('pending');
    });

    it('maps stopped to completed tab', () => {
      expect(STATUS_TO_TAB.stopped).toBe('completed');
    });

    it('maps completed to completed tab', () => {
      expect(STATUS_TO_TAB.completed).toBe('completed');
    });

    it('maps failed to completed tab', () => {
      expect(STATUS_TO_TAB.failed).toBe('completed');
    });
  });

  // =========================================================================
  // ORDER_TAB_FILTERS
  // =========================================================================

  describe('ORDER_TAB_FILTERS', () => {
    it('has exactly 4 tabs', () => {
      expect(ORDER_TAB_FILTERS).toHaveLength(4);
    });

    it('contains active, pending, completed, all', () => {
      const values = ORDER_TAB_FILTERS.map((t) => t.value);
      expect(values).toContain('active');
      expect(values).toContain('pending');
      expect(values).toContain('completed');
      expect(values).toContain('all');
    });

    it('each tab has label and value', () => {
      for (const tab of ORDER_TAB_FILTERS) {
        expect(tab.label).toBeTruthy();
        expect(tab.value).toBeTruthy();
      }
    });
  });
});
