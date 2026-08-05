import type { ParsedOcrResult } from "./types";

// Parser-derived text alone is useful for display, but cannot be verification evidence.
export function hasCalibratedFieldConfidence(result: ParsedOcrResult): boolean {
  return result.fields.length > 0 && result.fields.every((field) => Number.isFinite(field.confidence) && field.confidence > 0);
}
