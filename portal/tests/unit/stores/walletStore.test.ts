import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWalletStore } from '@/stores/walletStore';

// Mock Keplr wallet in window
beforeEach(() => {
  // @ts-expect-error - mocking window.keplr
  window.keplr = {
    enable: vi.fn(),
    getKey: vi.fn().mockResolvedValue({ bech32Address: 'virtengine1testaddress' }),
  };
});

describe('walletStore', () => {
  beforeEach(() => {
    // Reset the store before each test
    useWalletStore.setState({
      isConnected: false,
      isConnecting: false,
      address: null,
      walletType: null,
      balance: null,
      error: null,
    });
  });

  it('should have initial state', () => {
    const { result } = renderHook(() => useWalletStore());
    
    expect(result.current.isConnected).toBe(false);
    expect(result.current.isConnecting).toBe(false);
    expect(result.current.address).toBeNull();
    expect(result.current.walletType).toBeNull();
  });

  it('should connect wallet', async () => {
    const { result } = renderHook(() => useWalletStore());
    
    await act(async () => {
      await result.current.connect('keplr');
    });
    
    expect(result.current.isConnected).toBe(true);
    expect(result.current.walletType).toBe('keplr');
    expect(result.current.address).toBeTruthy();
  });

  it('should disconnect wallet', async () => {
    const { result } = renderHook(() => useWalletStore());
    
    // First connect
    await act(async () => {
      await result.current.connect('keplr');
    });
    
    expect(result.current.isConnected).toBe(true);
    
    // Then disconnect
    act(() => {
      result.current.disconnect();
    });
    
    expect(result.current.isConnected).toBe(false);
    expect(result.current.address).toBeNull();
    expect(result.current.walletType).toBeNull();
  });

  it('should set error', () => {
    const { result } = renderHook(() => useWalletStore());
    
    act(() => {
      result.current.setError('Test error');
    });
    
    expect(result.current.error).toBe('Test error');
  });
});
