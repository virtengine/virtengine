import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { SelfieResult } from '@/lib/capture-adapter';
import { createVeidCaptureProviders } from './VeidCaptureProviders';
import type { VeidCaptureProviders } from './VeidCaptureProviders';
import { createTestCaptureProviderInput } from './__tests__/captureProviderTestUtils';
import { VerificationWizard } from './VerificationWizard';

const wizard = vi.hoisted(() => ({
  setSelfie: vi.fn(),
  completeLiveness: vi.fn(),
  setError: vi.fn(),
}));

vi.mock('@/features/veid', () => ({
  WIZARD_STEPS: [{ key: 'selfie', label: 'Selfie', description: '', order: 1 }],
  MAX_RETRY_COUNT: 3,
  useVeidWizard: () => ({
    state: {
      currentStep: 'selfie',
      status: 'in-progress',
      captureData: {
        documentType: 'passport',
        documentFront: null,
        documentBack: null,
        selfie: null,
        livenessCompleted: false,
      },
      error: null,
      retryCount: 0,
    },
    navigation: { goBack: vi.fn(), goForward: vi.fn(), goToStep: vi.fn(), reset: vi.fn() },
    currentStepMeta: { key: 'selfie', label: 'Selfie', description: '', order: 1 },
    progressPercent: 50,
    selectDocumentType: vi.fn(),
    setDocumentFront: vi.fn(),
    setDocumentBack: vi.fn(),
    setSelfie: wizard.setSelfie,
    completeLiveness: wizard.completeLiveness,
    submit: vi.fn(),
    setError: wizard.setError,
    retry: vi.fn(),
  }),
}));

vi.mock('./SelfieCapture', () => ({
  VeidSelfieCapture: ({
    providers,
    onCapture,
  }: {
    providers: VeidCaptureProviders;
    onCapture: (result: SelfieResult) => void;
  }) => (
    <div>
      {providers.status === 'unavailable' && <span>{providers.reason}</span>}
      <button
        disabled={providers.status === 'unavailable'}
        onClick={() =>
          onCapture({
            metadata: { sessionId: 'session-1' },
            livenessCheck: {
              passed: false,
              providerId: 'test-liveness',
              providerVersion: '1.0.0',
              challengeId: 'challenge-1',
              sessionId: 'session-1',
              score: 0.2,
              challengeDurationMs: 10,
              evidenceDigest: 'sha256:failed',
            },
          } as SelfieResult)
        }
      >
        Failed evidence
      </button>
      <button
        disabled={providers.status === 'unavailable'}
        onClick={() =>
          onCapture({
            metadata: { sessionId: 'session-1' },
            livenessCheck: {
              passed: true,
              providerId: 'test-liveness',
              providerVersion: '1.0.0',
              challengeId: 'challenge-1',
              sessionId: 'session-1',
              score: 0.99,
              challengeDurationMs: 10,
              evidenceDigest: 'sha256:valid',
            },
          } as SelfieResult)
        }
      >
        Valid evidence
      </button>
    </div>
  ),
}));

describe('VerificationWizard liveness evidence', () => {
  beforeEach(() => vi.clearAllMocks());

  it('surfaces unavailable providers and disables capture', () => {
    render(<VerificationWizard />);
    expect(screen.getByText(/secure capture providers are unavailable/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Valid evidence' })).toBeDisabled();
  });

  it('blocks failed evidence', () => {
    const providers = createVeidCaptureProviders(createTestCaptureProviderInput(), 'test');
    render(<VerificationWizard providers={providers} />);
    fireEvent.click(screen.getByRole('button', { name: 'Failed evidence' }));
    expect(wizard.setError).toHaveBeenCalledOnce();
    expect(wizard.setSelfie).not.toHaveBeenCalled();
    expect(wizard.completeLiveness).not.toHaveBeenCalled();
  });

  it('advances exactly once for valid bound evidence', () => {
    const providers = createVeidCaptureProviders(createTestCaptureProviderInput(), 'test');
    render(<VerificationWizard providers={providers} />);
    fireEvent.click(screen.getByRole('button', { name: 'Valid evidence' }));
    expect(wizard.setSelfie).toHaveBeenCalledOnce();
    expect(wizard.completeLiveness).toHaveBeenCalledOnce();
    expect(wizard.setError).not.toHaveBeenCalled();
  });
});
