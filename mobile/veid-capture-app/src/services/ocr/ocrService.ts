import type { ParsedOcrResult } from "../../core/ocr/types";
import { parseOcrFields } from "../../core/ocr/fieldParser";
import { VerificationTerminalError } from "../../core/verificationError";

const sampleText = "ID CARD\nNAME JOHN DOE\nDOB 01/02/1990\nDOC 9A1B2C3D\nEXPIRY 01/02/2030";

export interface OcrRecognizer {
  recognize(imageUri: string): Promise<{ text?: string }>;
}

export interface OcrServiceOptions {
  loadRecognizer?: () => Promise<OcrRecognizer>;
}

async function loadNativeRecognizer(): Promise<OcrRecognizer> {
  try {
    const module = await import("@react-native-ml-kit/text-recognition");
    return module.default;
  } catch (error) {
    throw new VerificationTerminalError("ocr_module_unavailable", "Declared OCR native module failed to load.", { cause: error });
  }
}

export function createOcrService(options: OcrServiceOptions = {}) {
  const loadRecognizer = options.loadRecognizer ?? loadNativeRecognizer;

  return async (imageUri: string): Promise<ParsedOcrResult> => {
  if (imageUri.startsWith("mock://") && process.env.NODE_ENV === "test") {
    return parseOcrFields(sampleText);
  }

  let recognizer: OcrRecognizer;
  try {
    recognizer = await loadRecognizer();
  } catch (error) {
    if (error instanceof VerificationTerminalError) throw error;
    throw new VerificationTerminalError("ocr_module_unavailable", "Declared OCR native module failed to load.", { cause: error });
  }

  try {
    const recognition = await recognizer.recognize(imageUri);
    const text = recognition?.text ?? "";
    if (!text.trim()) {
      throw new VerificationTerminalError("ocr_empty_result", "OCR returned no readable text.");
    }
    return parseOcrFields(text);
  } catch (error) {
    if (error instanceof VerificationTerminalError) throw error;
    throw new VerificationTerminalError("ocr_recognition_failed", "OCR recognition failed; capture cannot proceed.", { cause: error });
  }
  };
}

export const extractOcr = createOcrService();
