import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSwipeGesture } from '@/hooks/useSwipeGesture';
import type { TouchEvent as ReactTouchEvent } from 'react';

function createTouchEvent(clientX: number, clientY: number) {
  return {
    touches: [{ clientX, clientY }],
    changedTouches: [{ clientX, clientY }],
  } as unknown as ReactTouchEvent;
}

describe('useSwipeGesture', () => {
  it('detects left swipe', () => {
    const onSwipeLeft = vi.fn();
    const { result } = renderHook(() =>
      useSwipeGesture({ onSwipeLeft }, { threshold: 30, timeout: 500 })
    );

    act(() => {
      result.current.onTouchStart(createTouchEvent(200, 100));
    });

    act(() => {
      result.current.onTouchEnd(createTouchEvent(100, 100));
    });

    expect(onSwipeLeft).toHaveBeenCalledTimes(1);
  });

  it('detects right swipe', () => {
    const onSwipeRight = vi.fn();
    const { result } = renderHook(() =>
      useSwipeGesture({ onSwipeRight }, { threshold: 30, timeout: 500 })
    );

    act(() => {
      result.current.onTouchStart(createTouchEvent(100, 100));
    });

    act(() => {
      result.current.onTouchEnd(createTouchEvent(200, 100));
    });

    expect(onSwipeRight).toHaveBeenCalledTimes(1);
  });

  it('detects up swipe', () => {
    const onSwipeUp = vi.fn();
    const { result } = renderHook(() =>
      useSwipeGesture({ onSwipeUp }, { threshold: 30, timeout: 500 })
    );

    act(() => {
      result.current.onTouchStart(createTouchEvent(100, 200));
    });

    act(() => {
      result.current.onTouchEnd(createTouchEvent(100, 100));
    });

    expect(onSwipeUp).toHaveBeenCalledTimes(1);
  });

  it('detects down swipe', () => {
    const onSwipeDown = vi.fn();
    const { result } = renderHook(() =>
      useSwipeGesture({ onSwipeDown }, { threshold: 30, timeout: 500 })
    );

    act(() => {
      result.current.onTouchStart(createTouchEvent(100, 100));
    });

    act(() => {
      result.current.onTouchEnd(createTouchEvent(100, 200));
    });

    expect(onSwipeDown).toHaveBeenCalledTimes(1);
  });

  it('ignores swipes below threshold', () => {
    const onSwipeLeft = vi.fn();
    const { result } = renderHook(() => useSwipeGesture({ onSwipeLeft }, { threshold: 100 }));

    act(() => {
      result.current.onTouchStart(createTouchEvent(200, 100));
    });

    act(() => {
      result.current.onTouchEnd(createTouchEvent(160, 100));
    });

    expect(onSwipeLeft).not.toHaveBeenCalled();
  });

  it('returns a ref object', () => {
    const { result } = renderHook(() => useSwipeGesture({}));
    expect(result.current.ref).toBeDefined();
    expect(result.current.ref.current).toBeNull();
  });
});
