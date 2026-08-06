import { describe, expect, it } from "vitest";
import { hasCalibratedFieldConfidence } from "../core/ocr/confidence";

describe("OCR confidence safety", () => {
  it("rejects parser-only OCR fields without calibrated confidence", () => {
    expect(hasCalibratedFieldConfidence({ rawText: "NAME TEST", fields: [{ key: "name", value: "TEST", confidence: 0 }] })).toBe(false);
  });
  it("accepts only finite positive calibrated confidence", () => {
    expect(hasCalibratedFieldConfidence({ rawText: "NAME TEST", fields: [{ key: "name", value: "TEST", confidence: 0.91 }] })).toBe(true);
  });
});
