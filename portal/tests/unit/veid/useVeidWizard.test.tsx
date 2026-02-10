import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useVeidWizard } from '@/features/veid/hooks/useVeidWizard';

describe('useVeidWizard', () => {
  it('initializes with welcome step and idle status', () => {
    const { result } = renderHook(() => useVeidWizard());

    expect(result.current.state.currentStep).toBe('welcome');
    expect(result.current.state.status).toBe('idle');
    expect(result.current.state.error).toBeNull();
    expect(result.current.state.retryCount).toBe(0);
  });

  it('calculates progress percent correctly', () => {
    const { result } = renderHook(() => useVeidWizard());

    // welcome is step 0 of 9, so (0+1)/9 * 100 = ~11%
    expect(result.current.progressPercent).toBe(11);
  });

  it('navigates forward from welcome', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.navigation.goForward();
    });

    expect(result.current.state.currentStep).toBe('document-select');
    expect(result.current.state.status).toBe('in-progress');
  });

  it('navigates back from document-select to welcome', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.navigation.goForward(); // -> document-select
    });
    act(() => {
      result.current.navigation.goBack(); // -> welcome
    });

    expect(result.current.state.currentStep).toBe('welcome');
  });

  it('cannot go back from welcome', () => {
    const { result } = renderHook(() => useVeidWizard());

    expect(result.current.navigation.canGoBack).toBe(false);

    act(() => {
      result.current.navigation.goBack();
    });

    expect(result.current.state.currentStep).toBe('welcome');
  });

  it('selectDocumentType advances to document-front', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.selectDocumentType('passport');
    });

    expect(result.current.state.currentStep).toBe('document-front');
    expect(result.current.state.captureData.documentType).toBe('passport');
    expect(result.current.state.status).toBe('in-progress');
    expect(result.current.state.startedAt).toBeTruthy();
  });

  it('setDocumentFront advances to document-back', () => {
    const { result } = renderHook(() => useVeidWizard());

    const mockResult = {
      blob: new Blob(['test'], { type: 'image/jpeg' }),
      metadata: {} as any,
      quality: {} as any,
      signatures: {} as any,
    };

    act(() => {
      result.current.selectDocumentType('id_card');
    });
    act(() => {
      result.current.setDocumentFront(mockResult as any);
    });

    expect(result.current.state.currentStep).toBe('document-back');
    expect(result.current.state.captureData.documentFront).toBeTruthy();
  });

  it('setDocumentBack advances to selfie', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.selectDocumentType('id_card');
    });
    act(() => {
      result.current.setDocumentFront({} as any);
    });
    act(() => {
      result.current.setDocumentBack({} as any);
    });

    expect(result.current.state.currentStep).toBe('selfie');
  });

  it('setSelfie advances to liveness', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.selectDocumentType('id_card');
    });
    act(() => {
      result.current.setDocumentFront({} as any);
    });
    act(() => {
      result.current.setDocumentBack({} as any);
    });
    act(() => {
      result.current.setSelfie({} as any);
    });

    expect(result.current.state.currentStep).toBe('liveness');
  });

  it('completeLiveness advances to review', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.selectDocumentType('id_card');
    });
    act(() => {
      result.current.setDocumentFront({} as any);
    });
    act(() => {
      result.current.setDocumentBack({} as any);
    });
    act(() => {
      result.current.setSelfie({} as any);
    });
    act(() => {
      result.current.completeLiveness();
    });

    expect(result.current.state.currentStep).toBe('review');
    expect(result.current.state.captureData.livenessCompleted).toBe(true);
  });

  it('submit transitions to complete', async () => {
    const { result } = renderHook(() => useVeidWizard());

    await act(async () => {
      await result.current.submit();
    });

    expect(result.current.state.currentStep).toBe('complete');
    expect(result.current.state.status).toBe('complete');
  });

  it('setError transitions to error state', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.setError({
        step: 'document-front',
        code: 'capture_error',
        message: 'Camera not available',
        retryable: true,
      });
    });

    expect(result.current.state.status).toBe('error');
    expect(result.current.state.error?.code).toBe('capture_error');
  });

  it('retry clears error and increments retryCount', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.navigation.goForward(); // -> document-select
    });
    act(() => {
      result.current.setError({
        step: 'document-select',
        code: 'capture_error',
        message: 'Camera not available',
        retryable: true,
      });
    });
    act(() => {
      result.current.retry();
    });

    expect(result.current.state.error).toBeNull();
    expect(result.current.state.status).toBe('in-progress');
    expect(result.current.state.retryCount).toBe(1);
  });

  it('retry is capped at MAX_RETRY_COUNT', () => {
    const { result } = renderHook(() => useVeidWizard());

    // Exhaust retries
    for (let i = 0; i < 4; i++) {
      act(() => {
        result.current.setError({
          step: 'selfie',
          code: 'capture_error',
          message: 'Failed',
          retryable: true,
        });
      });
      act(() => {
        result.current.retry();
      });
    }

    // After 3 retries, retry should not work anymore
    expect(result.current.state.retryCount).toBe(3);
  });

  it('reset returns to initial state', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.selectDocumentType('passport');
    });
    act(() => {
      result.current.navigation.reset();
    });

    expect(result.current.state.currentStep).toBe('welcome');
    expect(result.current.state.status).toBe('idle');
    expect(result.current.state.captureData.documentType).toBeNull();
  });

  it('goToStep navigates directly to a step', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => {
      result.current.navigation.goToStep('review');
    });

    expect(result.current.state.currentStep).toBe('review');
    expect(result.current.state.status).toBe('in-progress');
  });
});
