/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */
/**
 * Capture Adapter
 * Re-exports from @virtengine/capture for Next.js integration.
 */

import * as capture from '@virtengine/capture';

export const DocumentCapture = (capture as any).DocumentCapture;
export const CaptureGuidance = (capture as any).CaptureGuidance;
export const SelfieCapture = (capture as any).SelfieCapture;

export type DocumentType = any;
export type DocumentSide = any;
export type CaptureResult = any;
export type CaptureError = any;
export type GuidanceState = any;
export type ClientKeyProvider = any;
export type UserKeyProvider = any;
export type SelfieCaptureMode = any;
export type SelfieResult = any;
