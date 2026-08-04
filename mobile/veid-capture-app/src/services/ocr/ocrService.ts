import type { ParsedOcrResult } from "../../core/ocr/types";
import { parseOcrFields } from "../../core/ocr/fieldParser";
import { VerificationTerminalError } from "../../core/verificationError";

const sampleText = "ID CARD\nNAME JOHN DOE\nDOB 01/02/1990\nDOC 9A1B2C3D\nEXPIRY 01/02/2030";

export async function extractOcr(imageUri: string): Promise<ParsedOcrResult> {
  if (imageUri.startsWith("mock://") && process.env.NODE_ENV === "test") {
    return parseOcrFields(sampleText);
  }

  try {
    const module = await import("@react-native-ml-kit/text-recognition");
    const recognition = await module.default.recognize(imageUri);
    const text = recognition?.text ?? "";
    if (!text.trim()) {
      throw new VerificationTerminalError("ocr_empty_result", "OCR returned no readable text.");
    }
    return parseOcrFields(text);
  } catch (error) {
    if (error instanceof VerificationTerminalError) throw error;
    throw new VerificationTerminalError("ocr_module_unavailable", "OCR capability is unavailable; capture cannot proceed.", { cause: error });
  }
}
