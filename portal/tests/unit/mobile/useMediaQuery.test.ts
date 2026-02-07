import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  useMediaQuery,
  useIsMobile,
  useIsTablet,
  useIsDesktop,
  useIsTouchDevice,
} from '@/hooks/useMediaQuery';

// Helper to configure matchMedia mock
function mockMatchMedia(matches: boolean) {
  const listeners: Array<(e: { matches: boolean }) => void> = [];
  const mql = {
    matches,
    media: '',
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn((_, cb: (e: { matches: boolean }) => void) => listeners.push(cb)),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  };

  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockReturnValue(mql),
  });

  return { mql, fire: (m: boolean) => listeners.forEach((cb) => cb({ matches: m })) };
}

describe('useMediaQuery', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('returns false when query does not match', () => {
    mockMatchMedia(false);
    const { result } = renderHook(() => useMediaQuery('(max-width: 767px)'));
    expect(result.current).toBe(false);
  });

  it('returns true when query matches', () => {
    mockMatchMedia(true);
    const { result } = renderHook(() => useMediaQuery('(max-width: 767px)'));
    expect(result.current).toBe(true);
  });

  it('updates when match state changes', () => {
    const { mql } = mockMatchMedia(false);
    const { result } = renderHook(() => useMediaQuery('(max-width: 767px)'));

    expect(result.current).toBe(false);

    // Simulate viewport resize
    act(() => {
      mql.matches = true;
      const handler = mql.addEventListener.mock.calls[0]?.[1] as
        | ((e: { matches: boolean }) => void)
        | undefined;
      handler?.({ matches: true });
    });

    expect(result.current).toBe(true);
  });
});

describe('useIsMobile', () => {
  it('returns true for mobile viewports', () => {
    mockMatchMedia(true);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(true);
  });

  it('returns false for desktop viewports', () => {
    mockMatchMedia(false);
    const { result } = renderHook(() => useIsMobile());
    expect(result.current).toBe(false);
  });
});

describe('useIsTablet', () => {
  it('returns the matchMedia result', () => {
    mockMatchMedia(true);
    const { result } = renderHook(() => useIsTablet());
    expect(result.current).toBe(true);
  });
});

describe('useIsDesktop', () => {
  it('returns the matchMedia result', () => {
    mockMatchMedia(true);
    const { result } = renderHook(() => useIsDesktop());
    expect(result.current).toBe(true);
  });
});

describe('useIsTouchDevice', () => {
  it('returns the matchMedia result', () => {
    mockMatchMedia(true);
    const { result } = renderHook(() => useIsTouchDevice());
    expect(result.current).toBe(true);
  });
});
