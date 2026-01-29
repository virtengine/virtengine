/**
 * Portal Library Tests
 * VE-700 - VE-705: Testing portal utilities and hooks
 */
import { describe, it, expect, beforeAll, afterAll } from 'vitest';

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
const originalSessionStorage = (globalThis as any).sessionStorage;
const mockSessionStorage = (() => {
  const store = new Map<string, string>();
  return {
    getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
    setItem: (key: string, value: string) => {
      store.set(key, value);
    },
    removeItem: (key: string) => {
      store.delete(key);
    },
    clear: () => {
      store.clear();
    },
  };
})();

beforeAll(() => {
  Object.defineProperty(globalThis, 'crypto', {
    value: mockCrypto,
    configurable: true,
  });
  Object.defineProperty(globalThis, 'sessionStorage', {
    value: mockSessionStorage,
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

  if (originalSessionStorage) {
    Object.defineProperty(globalThis, 'sessionStorage', {
      value: originalSessionStorage,
      configurable: true,
    });
  } else {
    // @ts-expect-error - cleanup mocked global
    delete globalThis.sessionStorage;
  }
});

describe('Validation Utilities', () => {
  const {
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
  } = require('../utils/validation');

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

describe('Security Utilities', () => {
  const {
    sanitizePlainText,
    sanitizeDigits,
    sanitizeJsonInput,
  } = require('../utils/security');

  it('should sanitize plain text inputs', () => {
    const value = '<script>alert(1)</script>';
    const result = sanitizePlainText(value);
    expect(result).toContain('&lt;script&gt;');
  });

  it('should sanitize numeric inputs', () => {
    expect(sanitizeDigits('12ab34', 6)).toBe('1234');
  });

  it('should sanitize JSON inputs and strip dangerous keys', () => {
    const input = '{"__proto__":{"polluted":true},"safeKey":"ok"}';
    const result = sanitizeJsonInput(input);
    expect(result).toHaveProperty('safeKey');
    expect(result).not.toHaveProperty('__proto__');
  });
});

describe('OAuth Helpers', () => {
  const {
    createOAuthRequest,
    persistOAuthRequest,
    consumeOAuthRequest,
    buildAuthorizationUrl,
  } = require('../utils/oidc');

  it('should create and consume OAuth request', async () => {
    const request = await createOAuthRequest(60 * 1000);
    persistOAuthRequest(request, 'test_oauth');
    const consumed = consumeOAuthRequest(request.state, 'test_oauth');
    expect(consumed.state).toBe(request.state);
    expect(consumed.codeVerifier).toBe(request.codeVerifier);
  });

  it('should build authorization URL', async () => {
    const request = await createOAuthRequest();
    const url = buildAuthorizationUrl({
      provider: 'oidc',
      authorizationEndpoint: 'https://idp.example.com/auth',
      tokenEndpoint: 'https://idp.example.com/token',
      clientId: 'client-id',
      redirectUri: 'https://app.example.com/callback',
      scopes: ['openid', 'profile'],
      accountBindingEndpoint: 'https://idp.example.com/bind',
    }, request);
    expect(url).toContain('response_type=code');
    expect(url).toContain(`state=${request.state}`);
  });
});

describe('Format Utilities', () => {
  const {
    formatScore,
    formatTokenAmount,
    formatDuration,
    formatAddress,
    formatBytes,
    formatPercent,
    formatHash,
  } = require('../utils/format');

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
  const { authReducer, initialAuthState } = require('../types/auth');

  describe('authReducer', () => {
    it('should handle AUTH_START', () => {
      const result = authReducer(initialAuthState, { type: 'AUTH_START' });
      expect(result.isLoading).toBe(true);
      expect(result.error).toBeNull();
    });

    it('should handle AUTH_SUCCESS', () => {
      const session = {
        sessionId: 'test-session',
        createdAt: Date.now(),
        expiresAt: Date.now() + 3600,
        isTrustedBrowser: true,
        deviceFingerprint: 'abc123',
      };
      const result = authReducer(initialAuthState, {
        type: 'AUTH_SUCCESS',
        payload: {
          accountAddress: 've1abc',
          publicKey: 'deadbeef',
          method: 'wallet',
          session,
        },
      });
      expect(result.isAuthenticated).toBe(true);
      expect(result.session).toEqual(session);
    });

    it('should handle AUTH_FAILURE', () => {
      const result = authReducer(initialAuthState, {
        type: 'AUTH_FAILURE',
        payload: { code: 'invalid_credentials', message: 'nope' },
      });
      expect(result.isAuthenticated).toBe(false);
      expect(result.error?.code).toBe('invalid_credentials');
    });

    it('should handle AUTH_LOGOUT', () => {
      const loggedInState = {
        ...initialAuthState,
        isAuthenticated: true,
        accountAddress: 've1abc',
      };
      const result = authReducer(loggedInState, { type: 'AUTH_LOGOUT' });
      expect(result.isAuthenticated).toBe(false);
      expect(result.accountAddress).toBeNull();
    });
  });
});

describe('Identity Types', () => {
  const { identityReducer, SCOPE_REQUIREMENTS } = require('../types/identity');

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
  const { SessionManager, defaultSessionConfig } = require('../utils/session');

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
