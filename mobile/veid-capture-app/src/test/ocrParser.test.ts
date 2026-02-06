import { describe, expect, it } from "vitest";
import { parseOcrFields } from "../core/ocr/fieldParser";

const sample = "ID CARD\nNAME JOHN DOE\nSURNAME DOE\nDOB 01/02/1990\nDOC 9A1B2C3D\nEXPIRY 01/02/2030";

describe("parseOcrFields", () => {
  it("extracts key fields", () => {
    const result = parseOcrFields(sample);
    const keys = result.fields.map((field) => field.key);
    expect(keys).toContain("name");
    expect(keys).toContain("surname");
    expect(keys).toContain("date_of_birth");
    expect(keys).toContain("document_number");
  });
});
