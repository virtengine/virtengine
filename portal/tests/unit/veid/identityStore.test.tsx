import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useIdentityStore } from '@/stores/identityStore';

describe('identityStore (VEID extensions)', () => {
  beforeEach(() => {
    // Reset store between tests
    const { result } = renderHook(() => useIdentityStore());
    act(() => {
      result.current.reset();
    });
  });

  it('has correct initial state', () => {
    const { result } = renderHook(() => useIdentityStore());

    expect(result.current.veidScore).toBe(0);
    expect(result.current.isVerified).toBe(false);
    expect(result.current.wizardStep).toBeNull();
    expect(result.current.wizardStatus).toBe('idle');
    expect(result.current.hasCompletedOnboarding).toBe(false);
    expect(result.current.lastSubmissionAt).toBeNull();
  });

  it('setVeidScore updates score and verified status', () => {
    const { result } = renderHook(() => useIdentityStore());

    act(() => {
      result.current.setVeidScore(85);
    });

    expect(result.current.veidScore).toBe(85);
    expect(result.current.isVerified).toBe(true);
  });

  it('setVeidScore below threshold keeps unverified', () => {
    const { result } = renderHook(() => useIdentityStore());

    act(() => {
      result.current.setVeidScore(50);
    });

    expect(result.current.veidScore).toBe(50);
    expect(result.current.isVerified).toBe(false);
  });

  it('setWizardStep updates wizard step', () => {
    const { result } = renderHook(() => useIdentityStore());

    act(() => {
      result.current.setWizardStep('selfie');
    });

    expect(result.current.wizardStep).toBe('selfie');
  });

  it('setWizardStatus updates wizard status', () => {
    const { result } = renderHook(() => useIdentityStore());

    act(() => {
      result.current.setWizardStatus('in-progress');
    });

    expect(result.current.wizardStatus).toBe('in-progress');
  });

  it('completeOnboarding sets completion state', () => {
    const { result } = renderHook(() => useIdentityStore());

    act(() => {
      result.current.setWizardStep('review');
      result.current.completeOnboarding();
    });

    expect(result.current.hasCompletedOnboarding).toBe(true);
    expect(result.current.wizardStep).toBeNull();
    expect(result.current.wizardStatus).toBe('complete');
  });

  it('recordSubmission sets timestamp', () => {
    const { result } = renderHook(() => useIdentityStore());
    const before = Date.now();

    act(() => {
      result.current.recordSubmission();
    });

    const after = Date.now();
    expect(result.current.lastSubmissionAt).toBeGreaterThanOrEqual(before);
    expect(result.current.lastSubmissionAt).toBeLessThanOrEqual(after);
  });

  it('reset returns to initial state', () => {
    const { result } = renderHook(() => useIdentityStore());

    act(() => {
      result.current.setVeidScore(95);
      result.current.setWizardStep('complete');
      result.current.completeOnboarding();
    });

    act(() => {
      result.current.reset();
    });

    expect(result.current.veidScore).toBe(0);
    expect(result.current.isVerified).toBe(false);
    expect(result.current.wizardStep).toBeNull();
    expect(result.current.hasCompletedOnboarding).toBe(false);
  });
});
