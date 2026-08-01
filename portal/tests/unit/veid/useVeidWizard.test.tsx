import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  useVeidWizard,
  type UseVeidWizardOptions,
  type VeidSubmissionTransport,
  type VeidUploadRequest,
} from '@/features/veid/hooks/useVeidWizard';

const submission = {
  evidenceEnvelope: new Uint8Array([1, 2, 3]),
  payloadDigest: 'payload-digest',
  envelopeDigest: 'caller-supplied-envelope-digest',
  idempotencyKey: 'idempotency-key',
};

function createTransport(
  overrides: Partial<VeidSubmissionTransport> = {}
): VeidSubmissionTransport {
  return {
    uploadEvidence: vi.fn(async (request: VeidUploadRequest) => {
      return {
        receiptId: 'receipt-1',
        payloadDigest: request.payloadDigest,
        envelopeDigest: request.envelopeDigest,
        idempotencyKey: request.idempotencyKey,
      };
    }),
    authenticateUploadReceipt: vi.fn(async () => true),
    submitVerification: vi.fn(async (request) => ({
      committed: true,
      txHash: 'tx-hash',
      code: 0,
      blockHeight: 42,
      receiptId: request.receipt.receiptId,
      payloadDigest: request.payloadDigest,
      envelopeDigest: request.envelopeDigest,
      idempotencyKey: request.idempotencyKey,
    })),
    ...overrides,
  };
}

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

  it('fails closed when canonical envelope content is unavailable', async () => {
    const { result } = renderHook(() => useVeidWizard());

    let succeeded = true;
    await act(async () => {
      succeeded = await result.current.submit();
    });

    expect(succeeded).toBe(false);
    expect(result.current.state.status).toBe('error');
    expect(result.current.state.error?.code).toBe('evidence_envelope_unavailable');
  });

  it('fails closed when submission transport is unavailable', async () => {
    const { result } = renderHook(() => useVeidWizard({ submission }));

    await act(async () => {
      expect(await result.current.submit()).toBe(false);
    });

    expect(result.current.state.error?.code).toBe('transport_unavailable');
    expect(result.current.state.currentStep).not.toBe('complete');
  });

  it('completes only with an authenticated receipt and bound committed transaction', async () => {
    const transport = createTransport();
    const { result } = renderHook(() => useVeidWizard({ submission, transport }));

    await act(async () => {
      expect(await result.current.submit()).toBe(true);
    });

    expect(transport.authenticateUploadReceipt).toHaveBeenCalledOnce();
    expect(transport.submitVerification).toHaveBeenCalledOnce();
    expect(transport.uploadEvidence).toHaveBeenCalledWith(
      expect.objectContaining({ envelopeDigest: submission.envelopeDigest })
    );
    expect(transport.submitVerification).toHaveBeenCalledWith(
      expect.objectContaining({ envelopeDigest: submission.envelopeDigest })
    );
    expect(result.current.state.currentStep).toBe('complete');
    expect(result.current.state.status).toBe('complete');
  });

  it.each(['payloadDigest', 'envelopeDigest', 'idempotencyKey'] as const)(
    'fails closed when %s is blank',
    async (binding) => {
      const transport = createTransport();
      const { result } = renderHook(() =>
        useVeidWizard({
          submission: { ...submission, [binding]: '   ' },
          transport,
        })
      );

      await act(async () => {
        expect(await result.current.submit()).toBe(false);
      });

      expect(result.current.state.error?.code).toBe('submission_binding_invalid');
      expect(transport.uploadEvidence).not.toHaveBeenCalled();
    }
  );

  it.each([
    ['stale receipt', { receiptId: 'receipt-1', payloadDigest: 'stale' }],
    ['wrong envelope digest', { receiptId: 'receipt-1', envelopeDigest: 'wrong' }],
    ['wrong idempotency key', { receiptId: 'receipt-1', idempotencyKey: 'wrong' }],
  ])('rejects a %s', async (_label, receiptOverride) => {
    const transport = createTransport({
      uploadEvidence: vi.fn(async (request) =>
        Object.assign(
          {
            receiptId: 'receipt-1',
            payloadDigest: request.payloadDigest,
            envelopeDigest: request.envelopeDigest,
            idempotencyKey: request.idempotencyKey,
          },
          receiptOverride
        )
      ),
    });
    const { result } = renderHook(() => useVeidWizard({ submission, transport }));

    await act(async () => {
      expect(await result.current.submit()).toBe(false);
    });

    expect(result.current.state.error?.code).toBe('upload_receipt_invalid');
    expect(transport.submitVerification).not.toHaveBeenCalled();
  });

  it('rejects an unauthenticated receipt', async () => {
    const transport = createTransport({
      authenticateUploadReceipt: vi.fn(async () => false),
    });
    const { result } = renderHook(() => useVeidWizard({ submission, transport }));

    await act(async () => {
      expect(await result.current.submit()).toBe(false);
    });

    expect(result.current.state.error?.code).toBe('upload_receipt_unauthenticated');
    expect(transport.submitVerification).not.toHaveBeenCalled();
  });

  it.each([
    ['uncommitted', { committed: false }],
    ['nonzero code', { code: 9 }],
    ['missing transaction hash', { txHash: '' }],
    ['nonpositive block height', { blockHeight: 0 }],
    ['wrong receipt binding', { receiptId: 'other-receipt' }],
    ['wrong digest binding', { payloadDigest: 'other-payload' }],
    ['wrong envelope binding', { envelopeDigest: 'other-envelope' }],
  ])('rejects a %s transaction result', async (_label, transactionOverride) => {
    const baseline = createTransport();
    const transport = createTransport({
      submitVerification: vi.fn(async (request) => ({
        committed: true,
        txHash: 'tx-hash',
        code: 0,
        blockHeight: 42,
        receiptId: request.receipt.receiptId,
        payloadDigest: request.payloadDigest,
        envelopeDigest: request.envelopeDigest,
        idempotencyKey: request.idempotencyKey,
        ...transactionOverride,
      })),
      authenticateUploadReceipt: baseline.authenticateUploadReceipt,
    });
    const { result } = renderHook(() => useVeidWizard({ submission, transport }));

    await act(async () => {
      expect(await result.current.submit()).toBe(false);
    });

    expect(result.current.state.error?.code).toBe('transaction_not_committed');
    expect(result.current.state.currentStep).not.toBe('complete');
  });

  it('rejects a payload changed during upload', async () => {
    let finishUpload!: (value: unknown) => void;
    const pendingUpload = new Promise<unknown>((resolve) => {
      finishUpload = resolve;
    });
    const transport = createTransport({
      uploadEvidence: vi.fn(() => pendingUpload),
    });
    const initialOptions: UseVeidWizardOptions = { submission, transport };
    const { result, rerender } = renderHook(
      ({ options }: { options: UseVeidWizardOptions }) => useVeidWizard(options),
      { initialProps: { options: initialOptions } }
    );

    let submitPromise!: Promise<boolean>;
    act(() => {
      submitPromise = result.current.submit();
    });
    rerender({
      options: {
        transport,
        submission: { ...submission, payloadDigest: 'changed-payload' },
      },
    });
    await act(async () => {
      finishUpload({
        receiptId: 'receipt-1',
        payloadDigest: submission.payloadDigest,
        envelopeDigest: 'ignored-after-payload-change',
        idempotencyKey: submission.idempotencyKey,
      });
      expect(await submitPromise).toBe(false);
    });

    expect(result.current.state.error?.code).toBe('submission_payload_changed');
    expect(transport.submitVerification).not.toHaveBeenCalled();
  });

  it('rejects interrupted submission', async () => {
    const transport = createTransport({
      uploadEvidence: vi.fn(async () => {
        throw new DOMException('interrupted', 'AbortError');
      }),
    });
    const { result } = renderHook(() => useVeidWizard({ submission, transport }));

    await act(async () => {
      expect(await result.current.submit()).toBe(false);
    });

    expect(result.current.state.error?.code).toBe('submission_interrupted');
  });

  it('does not allow preview navigation to submission success', () => {
    const { result } = renderHook(() => useVeidWizard());

    act(() => result.current.navigation.goToStep('complete'));

    expect(result.current.state.currentStep).toBe('welcome');
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
