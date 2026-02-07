export interface OcrBlock {
  text: string;
  confidence: number;
}

export interface RawOcrResult {
  blocks: OcrBlock[];
  text: string;
}

export interface ParsedField {
  key: string;
  value: string;
  confidence: number;
}

export interface ParsedOcrResult {
  rawText: string;
  fields: ParsedField[];
}
