import React, { createContext, useContext, useReducer } from "react";
import type {
  BiometricCapture,
  CaptureSession,
  DocumentCapture,
  LivenessResult,
  OcrResult,
  SelfieCapture,
  SocialMediaProfile
} from "../core/captureModels";
import { initializeCaptureSession } from "../core/captureSession";
import { wipeCaptureSession } from "../services/cleanup/sensitiveData";

export type CaptureStep =
  | "consent"
  | "social_media"
  | "document_front"
  | "document_back"
  | "selfie"
  | "liveness"
  | "biometric"
  | "review"
  | "upload"
  | "complete";

export interface CaptureState {
  currentStep: CaptureStep;
  consentAccepted: boolean;
  session: CaptureSession;
  temporaryEmbeddingReferences: string[];
  encryptedPayloadBuffers: Uint8Array[];
  offlineQueuePlaintext: unknown[];
}

export type CaptureAction =
  | { type: "accept_consent" }
  | { type: "set_document"; payload: DocumentCapture }
  | { type: "set_selfie"; payload: SelfieCapture }
  | { type: "set_liveness"; payload: LivenessResult }
  | { type: "set_biometric"; payload: BiometricCapture }
  | { type: "set_ocr"; payload: OcrResult }
  | { type: "add_social"; payload: SocialMediaProfile }
  | { type: "track_temporary_embeddings"; payload: string[] }
  | { type: "track_encrypted_payload_buffer"; payload: Uint8Array }
  | { type: "set_offline_queue_plaintext"; payload: unknown[] }
  | { type: "wipe" }
  | { type: "complete" }
  | { type: "next" }
  | { type: "prev" };

const steps: CaptureStep[] = [
  "consent",
  "social_media",
  "document_front",
  "document_back",
  "selfie",
  "liveness",
  "biometric",
  "review",
  "upload",
  "complete"
];

const CaptureContext = createContext<
  { state: CaptureState; dispatch: React.Dispatch<CaptureAction> } | undefined
>(undefined);

export function captureReducer(state: CaptureState, action: CaptureAction): CaptureState {
  switch (action.type) {
    case "accept_consent":
      return {
        ...state,
        consentAccepted: true,
        currentStep: "social_media"
      };
    case "set_document":
      return {
        ...state,
        session:
          action.payload.side === "front"
            ? { ...state.session, documentFront: action.payload }
            : { ...state.session, documentBack: action.payload }
      };
    case "set_selfie":
      return {
        ...state,
        session: { ...state.session, selfie: action.payload }
      };
    case "set_liveness":
      return {
        ...state,
        session: { ...state.session, liveness: action.payload }
      };
    case "set_biometric":
      return {
        ...state,
        session: { ...state.session, biometric: action.payload }
      };
    case "set_ocr":
      return {
        ...state,
        session: { ...state.session, ocr: action.payload }
      };
    case "add_social": {
      const existing = state.session.socialMedia ?? [];
      const filtered = existing.filter((profile) => profile.provider !== action.payload.provider);
      return {
        ...state,
        session: { ...state.session, socialMedia: [...filtered, action.payload] }
      };
    }
    case "track_temporary_embeddings":
      return { ...state, temporaryEmbeddingReferences: action.payload };
    case "track_encrypted_payload_buffer":
      return { ...state, encryptedPayloadBuffers: [...state.encryptedPayloadBuffers, action.payload] };
    case "set_offline_queue_plaintext":
      return { ...state, offlineQueuePlaintext: action.payload };
    case "wipe":
      wipeCaptureSession(state.session);
      state.temporaryEmbeddingReferences.length = 0;
      for (const buffer of state.encryptedPayloadBuffers) buffer.fill(0);
      state.encryptedPayloadBuffers.length = 0;
      state.offlineQueuePlaintext.length = 0;
      return createInitialCaptureState();
    case "complete":
      return { ...state, currentStep: "complete" };
    case "next": {
      const index = steps.indexOf(state.currentStep);
      const requiresBack = state.session.documentType !== "passport";
      let nextStep = steps[Math.min(index + 1, steps.length - 1)];
      if (state.currentStep === "document_front" && !requiresBack) {
        nextStep = "selfie";
      }
      return { ...state, currentStep: nextStep };
    }
    case "prev": {
      const index = steps.indexOf(state.currentStep);
      const requiresBack = state.session.documentType !== "passport";
      let prevStep = steps[Math.max(index - 1, 0)];
      if (state.currentStep === "selfie" && !requiresBack) {
        prevStep = "document_front";
      }
      return { ...state, currentStep: prevStep };
    }
    default:
      return state;
  }
}

export function createInitialCaptureState(): CaptureState {
  return {
    currentStep: "consent",
    consentAccepted: false,
    session: initializeCaptureSession("id_card"),
    temporaryEmbeddingReferences: [],
    encryptedPayloadBuffers: [],
    offlineQueuePlaintext: []
  };
}

export function CaptureProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(captureReducer, undefined, createInitialCaptureState);
  return <CaptureContext.Provider value={{ state, dispatch }}>{children}</CaptureContext.Provider>;
}

export function useCaptureStore() {
  const context = useContext(CaptureContext);
  if (!context) {
    throw new Error("CaptureStore not initialized");
  }
  return context;
}
