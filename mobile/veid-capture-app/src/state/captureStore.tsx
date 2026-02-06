import React, { createContext, useContext, useReducer } from "react";
import type { CaptureSession, DocumentCapture, LivenessResult, OcrResult, SelfieCapture } from "../core/captureModels";
import { initializeCaptureSession } from "../core/captureSession";

export type CaptureStep =
  | "consent"
  | "document_front"
  | "document_back"
  | "selfie"
  | "liveness"
  | "review"
  | "upload"
  | "complete";

interface CaptureState {
  currentStep: CaptureStep;
  consentAccepted: boolean;
  session: CaptureSession;
}

type CaptureAction =
  | { type: "accept_consent" }
  | { type: "set_document"; payload: DocumentCapture }
  | { type: "set_selfie"; payload: SelfieCapture }
  | { type: "set_liveness"; payload: LivenessResult }
  | { type: "set_ocr"; payload: OcrResult }
  | { type: "next" }
  | { type: "prev" };

const steps: CaptureStep[] = [
  "consent",
  "document_front",
  "document_back",
  "selfie",
  "liveness",
  "review",
  "upload",
  "complete"
];

const CaptureContext = createContext<
  { state: CaptureState; dispatch: React.Dispatch<CaptureAction> } | undefined
>(undefined);

function reducer(state: CaptureState, action: CaptureAction): CaptureState {
  switch (action.type) {
    case "accept_consent":
      return {
        ...state,
        consentAccepted: true,
        currentStep: "document_front"
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
    case "set_ocr":
      return {
        ...state,
        session: { ...state.session, ocr: action.payload }
      };
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

const initialState: CaptureState = {
  currentStep: "consent",
  consentAccepted: false,
  session: initializeCaptureSession("id_card")
};

export function CaptureProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState);
  return <CaptureContext.Provider value={{ state, dispatch }}>{children}</CaptureContext.Provider>;
}

export function useCaptureStore() {
  const context = useContext(CaptureContext);
  if (!context) {
    throw new Error("CaptureStore not initialized");
  }
  return context;
}
