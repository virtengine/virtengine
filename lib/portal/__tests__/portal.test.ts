/**
 * Portal Library Tests
 * VE-700 - VE-705: Testing portal utilities and hooks
 */
import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import {
  validateAddress,
  validateMnemonic,
  isValidScore,
  validateEmail,
  validateUrl,
  validateDomain,
  isHex,
  validateTokenAmount,
  isPositiveInteger,
  validateOTPCode,
} from '../utils/validation';
import {
  formatScore,
  formatTokenAmount,
  formatDuration,
  formatAddress,
  formatBytes,
  formatPercent,
  formatHash,
} from '../utils/format';
import { authReducer } from '../types/auth';
import { identityReducer, SCOPE_REQUIREMENTS } from '../types/identity';
import { SessionManager, defaultSessionConfig } from '../utils/session';

// Mock crypto.subtle for testing
const mockCrypto = {
  getRandomValues: (arr: Uint8Array) => {
    for (let i = 0; i < arr.length; i++) {
      arr[i] = Math.floor(Math.random() * 256);
    }
    return arr;
  },
  subtle: {
    digest: async (algo: string, data: BufferSource) => {
      return new ArrayBuffer(32);
    },
    importKey: async () => ({ type: 'secret' }),
    sign: async () => new ArrayBuffer(32),
    generateKey: async () => ({
      publicKey: { type: 'public' },
      privateKey: { type: 'private' },
    }),
    exportKey: async () => new ArrayBuffer(65),
    deriveBits: async () => new ArrayBuffer(32),
    deriveKey: async () => ({ type: 'secret' }),
    encrypt: async () => new ArrayBuffer(48),
    decrypt: async () => new ArrayBuffer(16),
  },
};

const originalCrypto = globalThis.crypto;

beforeAll(() => {
  Object.defineProperty(globalThis, 'crypto', {
    value: mockCrypto,
    configurable: true,
  });
});

afterAll(() => {
  if (originalCrypto) {
    Object.defineProperty(globalThis, 'crypto', {
      value: originalCrypto,
      configurable: true,
    });
  } else {
    // @ts-expect-error - cleanup mocked global
    delete globalThis.crypto;
  }
});

describe('Validation Utilities', () => {
  describe('validateAddress', () => {
    it('should validate correct addresses', () => {
      expect(validateAddress('ve1abc123def456xyz789qrstuvwxyz1234567890', 've')).toBe(true);
    });

    it('should reject addresses with wrong prefix', () => {
      expect(validateAddress('cosmos1abc123def456xyz789qrstuvwxyz', 've')).toBe(false);
    });

    it('should reject empty addresses', () => {
      expect(validateAddress('', 've')).toBe(false);
    });

    it('should reject too short addresses', () => {
      expect(validateAddress('ve1abc', 've')).toBe(false);
    });
  });

  describe('validateMnemonic', () => {
    it('should validate 12-word mnemonic', () => {
      const mnemonic = 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about';
      expect(validateMnemonic(mnemonic)).toBe(true);
    });

    it('should validate 24-word mnemonic', () => {
      const mnemonic = 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art';
      expect(validateMnemonic(mnemonic)).toBe(true);
    });

    it('should reject invalid word count', () => {
      const mnemonic = 'abandon abandon abandon abandon abandon';
      expect(validateMnemonic(mnemonic)).toBe(false);
    });

    it('should reject empty mnemonic', () => {
      expect(validateMnemonic('')).toBe(false);
    });
  });

  describe('isValidScore', () => {
    it('should accept valid scores', () => {
      expect(isValidScore(0)).toBe(true);
      expect(isValidScore(50)).toBe(true);
      expect(isValidScore(100)).toBe(true);
    });

    it('should reject scores out of range', () => {
      expect(isValidScore(-1)).toBe(false);
      expect(isValidScore(101)).toBe(false);
    });

    it('should reject non-numbers', () => {
      expect(isValidScore(NaN)).toBe(false);
    });
  });

  describe('validateEmail', () => {
    it('should validate correct emails', () => {
      expect(validateEmail('test@example.com')).toBe(true);
      expect(validateEmail('user.name@domain.org')).toBe(true);
    });

    it('should reject invalid emails', () => {
      expect(validateEmail('not-an-email')).toBe(false);
      expect(validateEmail('@example.com')).toBe(false);
      expect(validateEmail('test@')).toBe(false);
    });
  });

  describe('validateUrl', () => {
    it('should validate correct URLs', () => {
      expect(validateUrl('https://example.com')).toBe(true);
      expect(validateUrl('http://localhost:3000')).toBe(true);
    });

    it('should reject non-http URLs', () => {
      expect(validateUrl('ftp://example.com')).toBe(false);
      expect(validateUrl('file:///path/to/file')).toBe(false);
    });

    it('should reject invalid URLs', () => {
      expect(validateUrl('not-a-url')).toBe(false);
    });
  });

  describe('validateDomain', () => {
    it('should validate correct domains', () => {
      expect(validateDomain('example.com')).toBe(true);
      expect(validateDomain('sub.domain.org')).toBe(true);
    });

    it('should reject invalid domains', () => {
      expect(validateDomain('not a domain')).toBe(false);
      expect(validateDomain('-invalid.com')).toBe(false);
    });
  });

  describe('isHex', () => {
    it('should validate hex strings', () => {
      expect(isHex('0x1234abcd')).toBe(true);
      expect(isHex('deadbeef')).toBe(true);
    });

    it('should reject non-hex strings', () => {
      expect(isHex('0xGHIJ')).toBe(false);
      expect(isHex('')).toBe(false);
    });
  });

  describe('validateTokenAmount', () => {
    it('should validate positive amounts', () => {
      expect(validateTokenAmount('100')).toBe(true);
      expect(validateTokenAmount('0.001')).toBe(true);
    });

    it('should reject invalid amounts', () => {
      expect(validateTokenAmount('')).toBe(false);
      expect(validateTokenAmount('not-a-number')).toBe(false);
      expect(validateTokenAmount('-10')).toBe(false);
    });
  });

  describe('isPositiveInteger', () => {
    it('should validate positive integers', () => {
      expect(isPositiveInteger(1)).toBe(true);
      expect(isPositiveInteger(100)).toBe(true);
    });

    it('should reject non-positive or non-integers', () => {
      expect(isPositiveInteger(0)).toBe(false);
      expect(isPositiveInteger(-1)).toBe(false);
      expect(isPositiveInteger(1.5)).toBe(false);
    });
  });

  describe('validateOTPCode', () => {
    it('should validate 6-digit codes', () => {
      expect(validateOTPCode('123456')).toBe(true);
      expect(validateOTPCode('000000')).toBe(true);
    });

    it('should reject invalid codes', () => {
      expect(validateOTPCode('12345')).toBe(false);
      expect(validateOTPCode('1234567')).toBe(false);
      expect(validateOTPCode('abcdef')).toBe(false);
    });
  });
});

describe('Format Utilities', () => {
  describe('formatScore', () => {
    it('should format scores as integers', () => {
      expect(formatScore(85.7)).toBe('86');
      expect(formatScore(0)).toBe('0');
      expect(formatScore(100)).toBe('100');
    });
  });

  describe('formatTokenAmount', () => {
    it('should format token amounts correctly', () => {
      expect(formatTokenAmount('1000000', 6, 'VE')).toContain('VE');
    });

    it('should handle zero amounts', () => {
      expect(formatTokenAmount('0', 6, 'VE')).toBe('0.00 VE');
    });
  });

  describe('formatDuration', () => {
    it('should format seconds', () => {
      expect(formatDuration(30)).toBe('30 seconds');
      expect(formatDuration(1)).toBe('1 second');
    });

    it('should format minutes', () => {
      expect(formatDuration(60)).toBe('1 minute');
      expect(formatDuration(120)).toBe('2 minutes');
    });

    it('should format hours', () => {
      expect(formatDuration(3600)).toBe('1 hour');
      expect(formatDuration(7200)).toBe('2 hours');
    });

    it('should format days', () => {
      expect(formatDuration(86400)).toBe('1 day');
      expect(formatDuration(172800)).toBe('2 days');
    });
  });

  describe('formatAddress', () => {
    it('should truncate long addresses', () => {
      const addr = 've1abc123def456xyz789qrstuvwxyz1234567890';
      const formatted = formatAddress(addr, 8);
      expect(formatted).toContain('...');
      expect(formatted.length).toBeLessThan(addr.length);
    });

    it('should not truncate short addresses', () => {
      const addr = 've1abc';
      expect(formatAddress(addr, 8)).toBe(addr);
    });
  });

  describe('formatBytes', () => {
    it('should format bytes correctly', () => {
      expect(formatBytes(0)).toBe('0 B');
      expect(formatBytes(1024)).toBe('1 KB');
      expect(formatBytes(1048576)).toBe('1 MB');
      expect(formatBytes(1073741824)).toBe('1 GB');
    });
  });

  describe('formatPercent', () => {
    it('should format percentages', () => {
      expect(formatPercent(50)).toBe('50.0%');
      expect(formatPercent(99.99, 2)).toBe('99.99%');
    });
  });

  describe('formatHash', () => {
    it('should format and truncate hashes', () => {
      const hash = 'abcdef1234567890abcdef1234567890';
      const formatted = formatHash(hash, 8);
      expect(formatted.startsWith('0x')).toBe(true);
      expect(formatted).toContain('...');
    });
  });
});

describe('Auth Types', () => {
  describe('authReducer', () => {
    it('should handle AUTH_START', () => {
      const state = { isLoading: false, error: null };
      const result = authReducer(state as any, { type: 'AUTH_START' });
      expect(result.isLoading).toBe(true);
    });

    it('should handle AUTH_SUCCESS', () => {
      const state = { isAuthenticated: false, session: null, isLoading: true };
      const payload = { 
        accountAddress: 've1...', 
        publicKey: 'pubkey123',
        method: 'wallet' as const,
        session: { sessionId: 'test', expiresAt: Date.now() + 3600000 }
      };
      const result = authReducer(state as any, { type: 'AUTH_SUCCESS', payload });
      expect(result.isAuthenticated).toBe(true);
      expect(result.isLoading).toBe(false);
    });

    it('should handle AUTH_LOGOUT', () => {
      const state = { isAuthenticated: true, session: {}, wallet: {} };
      const result = authReducer(state as any, { type: 'AUTH_LOGOUT' });
      expect(result.isAuthenticated).toBe(false);
      expect(result.session).toBeNull();
    });

    it('should handle AUTH_FAILURE', () => {
      const state = { error: null, isLoading: true };
      const error = { code: 'test_error' as const, message: 'Test error' };
      const result = authReducer(state as any, { type: 'AUTH_FAILURE', payload: error });
      expect(result.error).toEqual(error);
      expect(result.isLoading).toBe(false);
    });
  });
});

describe('Identity Types', () => {
  describe('SCOPE_REQUIREMENTS', () => {
    it('should have requirements for all scopes', () => {
      expect(SCOPE_REQUIREMENTS.basic).toBeDefined();
      expect(SCOPE_REQUIREMENTS.standard).toBeDefined();
      expect(SCOPE_REQUIREMENTS.enhanced).toBeDefined();
      expect(SCOPE_REQUIREMENTS.provider).toBeDefined();
    });

    it('should have increasing score requirements', () => {
      expect(SCOPE_REQUIREMENTS.basic.minimumScore).toBeLessThan(
        SCOPE_REQUIREMENTS.standard.minimumScore!
      );
      expect(SCOPE_REQUIREMENTS.standard.minimumScore!).toBeLessThan(
        SCOPE_REQUIREMENTS.enhanced.minimumScore!
      );
      expect(SCOPE_REQUIREMENTS.enhanced.minimumScore!).toBeLessThan(
        SCOPE_REQUIREMENTS.provider.minimumScore!
      );
    });
  });
});

describe('Session Manager', () => {
  describe('defaultSessionConfig', () => {
    it('should have sensible defaults', () => {
      expect(defaultSessionConfig.tokenLifetimeSeconds).toBe(3600);
      expect(defaultSessionConfig.refreshThresholdSeconds).toBe(300);
      expect(defaultSessionConfig.autoRefresh).toBe(true);
    });
  });

  describe('SessionManager', () => {
    it('should create with default config', () => {
      const manager = new SessionManager();
      expect(manager).toBeDefined();
    });

    it('should accept custom config', () => {
      const manager = new SessionManager({
        apiEndpoint: '/custom/api',
        tokenLifetimeSeconds: 7200,
      });
      expect(manager).toBeDefined();
    });
  });
});
