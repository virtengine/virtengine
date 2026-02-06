import type { ParsedOcrResult } from "../../core/ocr/types";
import { parseOcrFields } from "../../core/ocr/fieldParser";

const sampleText = "ID CARD\nNAME JOHN DOE\nDOB 01/02/1990\nDOC 9A1B2C3D\nEXPIRY 01/02/2030";

export async function extractOcr(imageUri: string): Promise<ParsedOcrResult> {
  if (imageUri.startsWith("mock://")) {
    return parseOcrFields(sampleText);
  }

  try {
    const module = await import("@react-native-ml-kit/text-recognition");
    const recognition = await module.default.recognize(imageUri);
    const text = recognition?.text ?? "";
    return parseOcrFields(text);
  } catch (error) {
    return parseOcrFields("");
  }
}
