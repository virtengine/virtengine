import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useWalletConnect } from '@/hooks/useWalletConnect';
import { renderHook, act } from '@testing-library/react';

// Mock the config module
vi.mock('@/config', () => ({
  WALLET_CONNECT_PROJECT_ID: '',
}));

describe('useWalletConnect', () => {
  it('has idle initial state', () => {
    const { result } = renderHook(() => useWalletConnect());

    expect(result.current.status).toBe('idle');
    expect(result.current.uri).toBeNull();
    expect(result.current.session).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('reports unsupported when project ID is empty', () => {
    const { result } = renderHook(() => useWalletConnect());
    expect(result.current.isSupported).toBe(false);
  });

  it('sets error when trying to connect without project ID', async () => {
    const { result } = renderHook(() => useWalletConnect());

    await act(async () => {
      await result.current.connect();
    });

    expect(result.current.status).toBe('error');
    expect(result.current.error).toBeTruthy();
  });

  it('resets state on disconnect', async () => {
    const { result } = renderHook(() => useWalletConnect());

    await act(async () => {
      await result.current.connect();
    });

    expect(result.current.status).toBe('error');

    await act(async () => {
      await result.current.disconnect();
    });

    expect(result.current.status).toBe('idle');
    expect(result.current.error).toBeNull();
    expect(result.current.uri).toBeNull();
    expect(result.current.session).toBeNull();
  });
});
