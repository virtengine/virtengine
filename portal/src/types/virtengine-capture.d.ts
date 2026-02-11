/* eslint-disable @typescript-eslint/no-explicit-any */
declare module '@virtengine/capture' {
  export const DocumentCapture: any;
  export const CaptureGuidance: any;
  export const SelfieCapture: any;
  export const submitCaptureScope: any;
  export const createUploadNonce: any;

  export type DocumentType = any;
  export type DocumentSide = any;
  export type CaptureResult = any;
  export type CaptureError = any;
  export type GuidanceState = any;
  export type ClientKeyProvider = any;
  export type UserKeyProvider = any;
  export type SelfieCaptureMode = any;
  export type SelfieResult = any;
  export type SubmissionRequest = any;
  export type SubmissionResult = any;
  export type SubmissionUpdate = any;
  export type SubmissionStatus = any;
  export type TxBroadcaster = any;
  export type UploadScopeMessage = any;
  export type ScopeTypeInput = any;
}

declare module '@virtengine/capture/*' {
  const captureModule: Record<string, unknown>;
  export default captureModule;
}

export {};
