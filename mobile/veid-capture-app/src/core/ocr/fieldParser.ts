import type { ParsedField, ParsedOcrResult } from "./types";

const datePattern = /\b(\d{2}[\/\-.]\d{2}[\/\-.]\d{4})\b/;
const documentIdPattern = /\b([A-Z0-9]{6,12})\b/;

function normalizeText(text: string): string {
  return text.replace(/\s+/g, " ").trim();
}

function findFirstMatch(text: string, pattern: RegExp): string | undefined {
  const match = text.match(pattern);
  return match?.[1];
}

export function parseOcrFields(rawText: string): ParsedOcrResult {
  const normalized = normalizeText(rawText.toUpperCase());
  const fields: ParsedField[] = [];

  const nameMatch = normalized.match(/NAME[:\s]+([A-Z\s]+)/);
  if (nameMatch?.[1]) {
    fields.push({ key: "name", value: nameMatch[1].trim(), confidence: 0.72 });
  }

  const surnameMatch = normalized.match(/SURNAME[:\s]+([A-Z\s]+)/);
  if (surnameMatch?.[1]) {
    fields.push({ key: "surname", value: surnameMatch[1].trim(), confidence: 0.72 });
  }

  const dob = findFirstMatch(normalized, datePattern);
  if (dob) {
    fields.push({ key: "date_of_birth", value: dob, confidence: 0.64 });
  }

  const expiryMatch = normalized.match(/EXPIRY[:\s]+(\d{2}[\/\-.]\d{2}[\/\-.]\d{4})/);
  if (expiryMatch?.[1]) {
    fields.push({ key: "expiry", value: expiryMatch[1], confidence: 0.64 });
  }

  const docId = findFirstMatch(normalized, documentIdPattern);
  if (docId) {
    fields.push({ key: "document_number", value: docId, confidence: 0.58 });
  }

  return {
    rawText,
    fields
  };
}
