/**
 * Security Utilities
 * VE-700: Client-side input sanitization helpers
 */

import { sanitizeString } from './validation';

const CONTROL_CHARS = /[\u0000-\u001f\u007f]/g;
const FORBIDDEN_KEYS = new Set(['__proto__', 'prototype', 'constructor']);

export interface SanitizeOptions {
  maxLength?: number;
}

export interface SanitizeObjectOptions {
  maxDepth?: number;
  maxKeyLength?: number;
  maxStringLength?: number;
  escapeHtmlStrings?: boolean;
}

/**
 * Sanitize plain text input for safe display/storage.
 */
export function sanitizePlainText(value: string, options: SanitizeOptions = {}): string {
  const maxLength = options.maxLength ?? 2000;
  const normalized = (value ?? '')
    .toString()
    .normalize('NFKC')
    .replace(CONTROL_CHARS, '')
    .trim()
    .slice(0, maxLength);

  return sanitizeString(normalized);
}

/**
 * Sanitize numeric-only input (e.g., OTP codes).
 */
export function sanitizeDigits(value: string, maxLength?: number): string {
  const digits = (value ?? '').replace(/\D/g, '');
  return typeof maxLength === 'number' ? digits.slice(0, maxLength) : digits;
}

/**
 * Parse JSON safely and sanitize keys/values to prevent prototype pollution.
 */
export function sanitizeJsonInput(
  value: string,
  options: SanitizeObjectOptions = {}
): Record<string, unknown> | null {
  if (!value || !value.trim()) return null;

  const parsed = JSON.parse(value);
  const sanitized = sanitizeObject(parsed, options);
  return (sanitized && typeof sanitized === 'object') ? sanitized as Record<string, unknown> : null;
}

/**
 * Deep-sanitize objects/arrays to remove dangerous keys and control chars.
 */
export function sanitizeObject(
  value: unknown,
  options: SanitizeObjectOptions = {},
  depth: number = 0
): unknown {
  const maxDepth = options.maxDepth ?? 6;
  const maxKeyLength = options.maxKeyLength ?? 64;
  const maxStringLength = options.maxStringLength ?? 2000;
  const escapeHtml = options.escapeHtmlStrings ?? false;

  if (depth > maxDepth) return null;

  if (Array.isArray(value)) {
    return value
      .map(item => sanitizeObject(item, options, depth + 1))
      .filter(item => item !== undefined);
  }

  if (value && typeof value === 'object') {
    const output: Record<string, unknown> = {};
    for (const [rawKey, rawValue] of Object.entries(value)) {
      if (FORBIDDEN_KEYS.has(rawKey)) continue;

      const safeKey = rawKey
        .toString()
        .normalize('NFKC')
        .replace(CONTROL_CHARS, '')
        .replace(/[^\w\-.:]/g, '')
        .slice(0, maxKeyLength);

      if (!safeKey) continue;

      output[safeKey] = sanitizeObject(rawValue, options, depth + 1);
    }
    return output;
  }

  if (typeof value === 'string') {
    const normalized = value
      .normalize('NFKC')
      .replace(CONTROL_CHARS, '')
      .slice(0, maxStringLength);
    return escapeHtml ? sanitizeString(normalized) : normalized;
  }

  if (typeof value === 'number' || typeof value === 'boolean' || value === null) {
    return value;
  }

  return undefined;
}
