/**
 * Validation Utilities
 * VE-700: Input validation helpers
 */

/**
 * Validate blockchain address format
 */
export function validateAddress(address: string, prefix: string = 've'): boolean {
  if (!address) return false;
  if (!address.startsWith(prefix)) return false;
  if (address.length < prefix.length + 38) return false; // Approximate bech32 length
  if (address.length > prefix.length + 60) return false;

  // Check for valid bech32 characters
  const validChars = /^[a-z0-9]+$/;
  const dataPart = address.slice(prefix.length + 1); // Skip prefix and separator
  return validChars.test(dataPart);
}

/**
 * Validate mnemonic phrase
 */
export function validateMnemonic(mnemonic: string): boolean {
  if (!mnemonic) return false;

  const words = mnemonic.trim().split(/\s+/);
  const validLengths = [12, 15, 18, 21, 24];

  if (!validLengths.includes(words.length)) {
    return false;
  }

  // Basic check: all words should be lowercase alphabetic
  const validWord = /^[a-z]+$/;
  return words.every(word => validWord.test(word) && word.length >= 2);
}

/**
 * Validate identity score range
 */
export function isValidScore(score: number): boolean {
  return typeof score === 'number' && score >= 0 && score <= 100;
}

/**
 * Validate email format
 */
export function validateEmail(email: string): boolean {
  if (!email) return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Validate URL format
 */
export function validateUrl(url: string): boolean {
  if (!url) return false;
  try {
    const parsed = new URL(url);
    return ['http:', 'https:'].includes(parsed.protocol);
  } catch {
    return false;
  }
}

/**
 * Validate domain name
 */
export function validateDomain(domain: string): boolean {
  if (!domain) return false;
  // Basic domain validation
  const domainRegex = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$/i;
  return domainRegex.test(domain);
}

/**
 * Validate hex string
 */
export function isHex(value: string): boolean {
  if (!value) return false;
  const hex = value.startsWith('0x') ? value.slice(2) : value;
  return /^[0-9a-fA-F]+$/.test(hex);
}

/**
 * Validate token amount
 */
export function validateTokenAmount(amount: string): boolean {
  if (!amount) return false;
  const num = parseFloat(amount);
  return !isNaN(num) && num >= 0 && isFinite(num);
}

/**
 * Validate positive integer
 */
export function isPositiveInteger(value: number): boolean {
  return Number.isInteger(value) && value > 0;
}

/**
 * Validate duration in seconds
 */
export function validateDuration(seconds: number, min?: number, max?: number): boolean {
  if (!isPositiveInteger(seconds)) return false;
  if (min !== undefined && seconds < min) return false;
  if (max !== undefined && seconds > max) return false;
  return true;
}

/**
 * Sanitize string for display (prevent XSS)
 */
export function sanitizeString(str: string): string {
  if (!str) return '';
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

/**
 * Validate OTP code format
 */
export function validateOTPCode(code: string): boolean {
  if (!code) return false;
  // Standard TOTP is 6 digits
  return /^\d{6}$/.test(code);
}

/**
 * Validate private key format (hex, 32 bytes)
 */
export function validatePrivateKey(key: string): boolean {
  if (!key) return false;
  const hex = key.startsWith('0x') ? key.slice(2) : key;
  return /^[0-9a-fA-F]{64}$/.test(hex);
}
