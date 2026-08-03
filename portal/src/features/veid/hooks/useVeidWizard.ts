/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * useVeidWizard Hook
 * State machine for the VEID verification wizard flow.
 */

'use client';

import { useState, useCallback, useEffect, useMemo, useRef } from 'react';
import type { DocumentType } from '@/lib/capture-adapter';
import type { CaptureResult, SelfieResult } from '@/lib/capture-adapter';
import type {
  VeidWizardState,
  WizardStep,
  WizardCaptureData,
  WizardError,
  WizardNavigation,
  WizardStepMeta,
} from '../types';
import { WIZARD_STEPS, MAX_RETRY_COUNT } from '../constants';

const initialCaptureData: WizardCaptureData = {
  documentType: null,
  documentFront: null,
  documentBack: null,
  selfie: null,
  livenessCompleted: false,
};

const initialState: VeidWizardState = {
  currentStep: 'welcome',
  status: 'idle',
  captureData: initialCaptureData,
  error: null,
  retryCount: 0,
  startedAt: null,
};

/** Valid step transitions */
const STEP_ORDER: WizardStep[] = [
  'welcome',
  'document-select',
  'document-front',
  'document-back',
  'selfie',
  'liveness',
  'review',
  'submitting',
  'complete',
];

function getStepIndex(step: WizardStep): number {
  return STEP_ORDER.indexOf(step);
}

export interface UseVeidWizardReturn {
  state: VeidWizardState;
  navigation: WizardNavigation;
  currentStepMeta: WizardStepMeta | undefined;
  progressPercent: number;
  /** Set selected document type */
  selectDocumentType: (type: DocumentType) => void;
  /** Handle document front capture */
  setDocumentFront: (result: CaptureResult) => void;
  /** Handle document back capture */
  setDocumentBack: (result: CaptureResult) => void;
  /** Handle selfie capture */
  setSelfie: (result: SelfieResult) => void;
  /** Mark liveness as completed */
  completeLiveness: () => void;
  /** Submit verification and return whether it reached committed chain state. */
  submit: () => Promise<boolean>;
  /** Set an error */
  setError: (error: WizardError) => void;
  /** Retry from error state */
  retry: () => void;
}

export interface VeidSubmissionInput {
  /** Canonical Evidence Envelope bytes. The wizard treats their content as opaque. */
  evidenceEnvelope: Uint8Array | null;
  payloadDigest: string;
  envelopeDigest: string;
  idempotencyKey: string;
}

export interface VeidUploadRequest {
  evidenceEnvelope: Uint8Array;
  payloadDigest: string;
  envelopeDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}

export interface VeidUploadReceipt {
  receiptId: string;
  payloadDigest: string;
  envelopeDigest: string;
  idempotencyKey: string;
}

export interface VeidTransactionRequest {
  receipt: VeidUploadReceipt;
  payloadDigest: string;
  envelopeDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}

export interface VeidTransactionResult {
  committed: true;
  txHash: string;
  code: 0;
  blockHeight: number;
  receiptId: string;
  payloadDigest: string;
  envelopeDigest: string;
  idempotencyKey: string;
}

export interface VeidSubmissionTransport {
  uploadEvidence(request: VeidUploadRequest): Promise<unknown>;
  authenticateUploadReceipt(
    receipt: VeidUploadReceipt,
    request: VeidUploadRequest
  ): Promise<boolean>;
  submitVerification(request: VeidTransactionRequest): Promise<unknown>;
}

export interface UseVeidWizardOptions {
  submission?: VeidSubmissionInput;
  transport?: VeidSubmissionTransport;
}

function bytesEqual(left: Uint8Array, right: Uint8Array): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}

function submissionMatches(
  submission: VeidSubmissionInput | undefined,
  envelope: Uint8Array,
  payloadDigest: string,
  envelopeDigest: string,
  idempotencyKey: string
): boolean {
  return Boolean(
    submission?.evidenceEnvelope &&
    submission.payloadDigest.trim() === payloadDigest &&
    submission.envelopeDigest.trim() === envelopeDigest &&
    submission.idempotencyKey.trim() === idempotencyKey &&
    bytesEqual(submission.evidenceEnvelope, envelope)
  );
}

function isBoundReceipt(value: unknown, request: VeidUploadRequest): value is VeidUploadReceipt {
  if (typeof value !== 'object' || value === null) return false;
  const receipt = value as Partial<VeidUploadReceipt>;
  return (
    typeof receipt.receiptId === 'string' &&
    receipt.receiptId.trim().length > 0 &&
    receipt.payloadDigest === request.payloadDigest &&
    receipt.envelopeDigest === request.envelopeDigest &&
    receipt.idempotencyKey === request.idempotencyKey
  );
}

function isCommittedTransaction(
  value: unknown,
  request: VeidTransactionRequest
): value is VeidTransactionResult {
  if (typeof value !== 'object' || value === null) return false;
  const result = value as Partial<VeidTransactionResult>;
  return (
    result.committed === true &&
    typeof result.txHash === 'string' &&
    result.txHash.trim().length > 0 &&
    result.code === 0 &&
    typeof result.blockHeight === 'number' &&
    Number.isInteger(result.blockHeight) &&
    result.blockHeight > 0 &&
    result.receiptId === request.receipt.receiptId &&
    result.payloadDigest === request.payloadDigest &&
    result.envelopeDigest === request.envelopeDigest &&
    result.idempotencyKey === request.idempotencyKey
  );
}

function submissionError(code: string, message: string, retryable: boolean): WizardError {
  return { step: 'submitting', code, message, retryable };
}

export function useVeidWizard(options: UseVeidWizardOptions = {}): UseVeidWizardReturn {
  const [state, setState] = useState<VeidWizardState>(initialState);
  const optionsRef = useRef(options);
  const submissionIdRef = useRef(0);
  const activeSubmissionRef = useRef<{ id: number; controller: AbortController } | null>(null);
  optionsRef.current = options;

  useEffect(
    () => () => {
      activeSubmissionRef.current?.controller.abort();
      activeSubmissionRef.current = null;
    },
    []
  );

  const goToStep = useCallback((step: WizardStep) => {
    if (activeSubmissionRef.current || step === 'submitting' || step === 'complete') return;
    setState((prev) => ({
      ...prev,
      currentStep: step,
      status: 'in-progress',
      error: null,
    }));
  }, []);

  const goBack = useCallback(() => {
    if (activeSubmissionRef.current) return;
    setState((prev) => {
      const currentIndex = getStepIndex(prev.currentStep);
      if (currentIndex <= 0) return prev;
      const prevStep = STEP_ORDER[currentIndex - 1];
      if (!prevStep) return prev;
      return { ...prev, currentStep: prevStep, error: null };
    });
  }, []);

  const goForward = useCallback(() => {
    if (activeSubmissionRef.current) return;
    setState((prev) => {
      const currentIndex = getStepIndex(prev.currentStep);
      if (currentIndex >= STEP_ORDER.length - 1) return prev;
      const nextStep = STEP_ORDER[currentIndex + 1];
      if (!nextStep || nextStep === 'submitting' || nextStep === 'complete') return prev;
      return {
        ...prev,
        currentStep: nextStep,
        status: 'in-progress',
        error: null,
      };
    });
  }, []);

  const reset = useCallback(() => {
    activeSubmissionRef.current?.controller.abort();
    activeSubmissionRef.current = null;
    setState(initialState);
  }, []);

  const selectDocumentType = useCallback((type: DocumentType) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, documentType: type },
      currentStep: 'document-front',
      status: 'in-progress',
      startedAt: prev.startedAt ?? Date.now(),
    }));
  }, []);

  const setDocumentFront = useCallback((result: CaptureResult) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, documentFront: result },
      currentStep: 'document-back',
    }));
  }, []);

  const setDocumentBack = useCallback((result: CaptureResult) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, documentBack: result },
      currentStep: 'selfie',
    }));
  }, []);

  const setSelfie = useCallback((result: SelfieResult) => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, selfie: result },
      currentStep: 'liveness',
    }));
  }, []);

  const completeLiveness = useCallback(() => {
    setState((prev) => ({
      ...prev,
      captureData: { ...prev.captureData, livenessCompleted: true },
      currentStep: 'review',
    }));
  }, []);

  const setError = useCallback((error: WizardError) => {
    setState((prev) => ({
      ...prev,
      error,
      status: 'error',
    }));
  }, []);

  const retry = useCallback(() => {
    setState((prev) => {
      if (prev.retryCount >= MAX_RETRY_COUNT) return prev;
      const retryStep =
        prev.error?.step === 'submitting' ? 'review' : (prev.error?.step ?? prev.currentStep);
      return {
        ...prev,
        currentStep: retryStep,
        status: 'in-progress',
        error: null,
        retryCount: prev.retryCount + 1,
      };
    });
  }, []);

  const submit = useCallback(async (): Promise<boolean> => {
    const submission = optionsRef.current.submission;
    const transport = optionsRef.current.transport;

    if (!submission?.evidenceEnvelope?.length) {
      setState((prev) => ({
        ...prev,
        error: submissionError(
          'evidence_envelope_unavailable',
          'Canonical Evidence Envelope content is unavailable.',
          false
        ),
        status: 'error',
      }));
      return false;
    }
    if (!transport) {
      setState((prev) => ({
        ...prev,
        error: submissionError(
          'transport_unavailable',
          'Verification submission transport is unavailable.',
          true
        ),
        status: 'error',
      }));
      return false;
    }

    activeSubmissionRef.current?.controller.abort();
    const activeSubmission = {
      id: (submissionIdRef.current += 1),
      controller: new AbortController(),
    };
    activeSubmissionRef.current = activeSubmission;
    const envelopeSnapshot = new Uint8Array(submission.evidenceEnvelope);
    const payloadDigest = submission.payloadDigest.trim();
    const envelopeDigest = submission.envelopeDigest.trim();
    const idempotencyKey = submission.idempotencyKey.trim();

    setState((prev) => ({
      ...prev,
      currentStep: 'submitting',
      status: 'submitting',
      error: null,
    }));

    try {
      if (!payloadDigest || !envelopeDigest || !idempotencyKey) {
        throw submissionError(
          'submission_binding_invalid',
          'Payload digest, envelope digest, and idempotency key are required.',
          false
        );
      }
      const uploadRequest: VeidUploadRequest = {
        evidenceEnvelope: envelopeSnapshot,
        payloadDigest,
        envelopeDigest,
        idempotencyKey,
        signal: activeSubmission.controller.signal,
      };
      const receiptValue = await transport.uploadEvidence(uploadRequest);
      if (
        activeSubmission.controller.signal.aborted ||
        !submissionMatches(
          optionsRef.current.submission,
          envelopeSnapshot,
          payloadDigest,
          envelopeDigest,
          idempotencyKey
        )
      ) {
        throw submissionError(
          'submission_payload_changed',
          'Verification payload changed while submission was in progress.',
          true
        );
      }
      if (!isBoundReceipt(receiptValue, uploadRequest)) {
        throw submissionError(
          'upload_receipt_invalid',
          'Upload receipt does not match the submitted evidence.',
          true
        );
      }
      if (!(await transport.authenticateUploadReceipt(receiptValue, uploadRequest))) {
        throw submissionError(
          'upload_receipt_unauthenticated',
          'Upload receipt authentication failed.',
          true
        );
      }
      if (
        activeSubmission.controller.signal.aborted ||
        !submissionMatches(
          optionsRef.current.submission,
          envelopeSnapshot,
          payloadDigest,
          envelopeDigest,
          idempotencyKey
        )
      ) {
        throw submissionError(
          'submission_payload_changed',
          'Verification payload changed while submission was in progress.',
          true
        );
      }

      const transactionRequest: VeidTransactionRequest = {
        receipt: receiptValue,
        payloadDigest,
        envelopeDigest,
        idempotencyKey,
        signal: activeSubmission.controller.signal,
      };
      const transaction = await transport.submitVerification(transactionRequest);
      if (activeSubmission.controller.signal.aborted) {
        throw submissionError(
          'submission_interrupted',
          'Verification submission was interrupted.',
          true
        );
      }
      if (
        !submissionMatches(
          optionsRef.current.submission,
          envelopeSnapshot,
          payloadDigest,
          envelopeDigest,
          idempotencyKey
        )
      ) {
        throw submissionError(
          'submission_payload_changed',
          'Verification payload changed while submission was in progress.',
          true
        );
      }
      if (!isCommittedTransaction(transaction, transactionRequest)) {
        throw submissionError(
          'transaction_not_committed',
          'Verification transaction was not committed with matching evidence.',
          true
        );
      }
      if (activeSubmissionRef.current?.id !== activeSubmission.id) return false;

      setState((prev) => ({
        ...prev,
        currentStep: 'complete',
        status: 'complete',
      }));
      activeSubmissionRef.current = null;
      return true;
    } catch (error) {
      if (activeSubmissionRef.current?.id !== activeSubmission.id) return false;
      const interrupted =
        activeSubmission.controller.signal.aborted ||
        (typeof error === 'object' &&
          error !== null &&
          'name' in error &&
          error.name === 'AbortError');
      const wizardError =
        typeof error === 'object' &&
        error !== null &&
        'step' in error &&
        'code' in error &&
        'message' in error &&
        'retryable' in error
          ? (error as WizardError)
          : interrupted
            ? submissionError(
                'submission_interrupted',
                'Verification submission was interrupted.',
                true
              )
            : submissionError(
                'submission_failed',
                'Failed to submit verification to chain. Please try again.',
                true
              );
      setState((prev) => ({
        ...prev,
        currentStep: 'review',
        error: wizardError,
        status: 'error',
      }));
      activeSubmissionRef.current = null;
      return false;
    }
  }, []);

  const navigation: WizardNavigation = useMemo(
    () => ({
      canGoBack:
        getStepIndex(state.currentStep) > 0 &&
        state.currentStep !== 'submitting' &&
        state.currentStep !== 'complete',
      canGoForward:
        getStepIndex(state.currentStep) < STEP_ORDER.length - 1 &&
        state.currentStep !== 'submitting',
      goBack,
      goForward,
      goToStep,
      reset,
    }),
    [state.currentStep, goBack, goForward, goToStep, reset]
  );

  const currentStepMeta = WIZARD_STEPS.find((s) => s.key === state.currentStep);

  const progressPercent = useMemo(() => {
    const idx = getStepIndex(state.currentStep);
    return Math.round(((idx + 1) / STEP_ORDER.length) * 100);
  }, [state.currentStep]);

  return {
    state,
    navigation,
    currentStepMeta,
    progressPercent,
    selectDocumentType,
    setDocumentFront,
    setDocumentBack,
    setSelfie,
    completeLiveness,
    submit,
    setError,
    retry,
  };
}
