/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { useRef, useCallback, type RefObject } from 'react';

export type SwipeDirection = 'left' | 'right' | 'up' | 'down';

interface SwipeHandlers {
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  onSwipeUp?: () => void;
  onSwipeDown?: () => void;
}

interface SwipeOptions {
  /** Minimum distance in px to trigger a swipe (default: 50) */
  threshold?: number;
  /** Maximum time in ms for the swipe gesture (default: 300) */
  timeout?: number;
}

interface TouchState {
  startX: number;
  startY: number;
  startTime: number;
}

/**
 * Hook providing touch event handlers for swipe gesture detection.
 * Attach the returned handlers to any element.
 */
export function useSwipeGesture(
  handlers: SwipeHandlers,
  options: SwipeOptions = {}
): {
  ref: RefObject<HTMLElement | null>;
  onTouchStart: (e: React.TouchEvent) => void;
  onTouchEnd: (e: React.TouchEvent) => void;
} {
  const { threshold = 50, timeout = 300 } = options;
  const touchState = useRef<TouchState | null>(null);
  const ref = useRef<HTMLElement | null>(null);

  const onTouchStart = useCallback((e: React.TouchEvent) => {
    const touch = e.touches[0];
    if (!touch) return;
    touchState.current = {
      startX: touch.clientX,
      startY: touch.clientY,
      startTime: Date.now(),
    };
  }, []);

  const onTouchEnd = useCallback(
    (e: React.TouchEvent) => {
      if (!touchState.current) return;

      const touch = e.changedTouches[0];
      if (!touch) return;

      const { startX, startY, startTime } = touchState.current;
      const deltaX = touch.clientX - startX;
      const deltaY = touch.clientY - startY;
      const elapsed = Date.now() - startTime;

      touchState.current = null;

      if (elapsed > timeout) return;

      const absX = Math.abs(deltaX);
      const absY = Math.abs(deltaY);

      if (absX < threshold && absY < threshold) return;

      if (absX > absY) {
        if (deltaX > 0) {
          handlers.onSwipeRight?.();
        } else {
          handlers.onSwipeLeft?.();
        }
      } else {
        if (deltaY > 0) {
          handlers.onSwipeDown?.();
        } else {
          handlers.onSwipeUp?.();
        }
      }
    },
    [handlers, threshold, timeout]
  );

  return { ref, onTouchStart, onTouchEnd };
}
