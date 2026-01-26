/**
 * Salt Generator Tests
 * VE-210: Unit tests for salt generation utilities
 */

import {
  generateSalt,
  generateDeviceBoundSalt,
  generateSessionBoundSalt,
  verifySalt,
  bytesToHex,
  hexToBytes,
  bytesToBase64,
  base64ToBytes,
  concatBytes,
  DEFAULT_SALT_LENGTH,
} from '../utils/salt-generator';

describe('salt-generator', () => {
  describe('generateSalt', () => {
    it('should generate salt with default length', () => {
      const result = generateSalt();

      expect(result.salt).toBeInstanceOf(Uint8Array);
      expect(result.salt.length).toBe(DEFAULT_SALT_LENGTH);
    });

    it('should generate salt with custom length', () => {
      const result = generateSalt({ length: 64 });

      expect(result.salt.length).toBe(64);
    });

    it('should include timestamp in result', () => {
      const before = new Date();
      const result = generateSalt();
      const after = new Date();

      expect(result.generatedAt).toBeInstanceOf(Date);
      expect(result.generatedAt.getTime()).toBeGreaterThanOrEqual(before.getTime());
      expect(result.generatedAt.getTime()).toBeLessThanOrEqual(after.getTime());
    });

    it('should include hex representation', () => {
      const result = generateSalt();

      expect(typeof result.hex).toBe('string');
      expect(result.hex.length).toBe(DEFAULT_SALT_LENGTH * 2);
      expect(/^[0-9a-f]+$/.test(result.hex)).toBe(true);
    });

    it('should include base64 representation', () => {
      const result = generateSalt();

      expect(typeof result.base64).toBe('string');
      expect(result.base64.length).toBeGreaterThan(0);
    });

    it('should generate unique salts', () => {
      const salts = new Set<string>();
      for (let i = 0; i < 100; i++) {
        const result = generateSalt();
        salts.add(result.hex);
      }

      expect(salts.size).toBe(100);
    });

    it('should prepend prefix when provided', () => {
      const prefix = new Uint8Array([0x01, 0x02, 0x03, 0x04]);
      const result = generateSalt({ prefix, length: 32 });

      expect(result.salt.length).toBe(prefix.length + 32);
      expect(result.salt.slice(0, prefix.length)).toEqual(prefix);
    });

    it('should incorporate timestamp binding by default', () => {
      // Generate two salts at different times
      const result1 = generateSalt({ timestampBinding: true });

      // Sleep briefly
      const start = Date.now();
      while (Date.now() - start < 10) {
        // busy wait
      }

      const result2 = generateSalt({ timestampBinding: true });

      // They should be different (extremely high probability)
      expect(result1.hex).not.toBe(result2.hex);
    });
  });

  describe('generateDeviceBoundSalt', () => {
    it('should generate salt bound to device fingerprint', async () => {
      const fingerprint = 'test-device-fingerprint-123';
      const result = await generateDeviceBoundSalt(fingerprint);

      expect(result.salt).toBeInstanceOf(Uint8Array);
      expect(result.salt.length).toBe(DEFAULT_SALT_LENGTH);
    });

    it('should produce different salts for different fingerprints', async () => {
      const result1 = await generateDeviceBoundSalt('fingerprint-1');
      const result2 = await generateDeviceBoundSalt('fingerprint-2');

      expect(result1.hex).not.toBe(result2.hex);
    });

    it('should produce different salts for same fingerprint (random component)', async () => {
      const fingerprint = 'same-fingerprint';
      const result1 = await generateDeviceBoundSalt(fingerprint);
      const result2 = await generateDeviceBoundSalt(fingerprint);

      expect(result1.hex).not.toBe(result2.hex);
    });
  });

  describe('generateSessionBoundSalt', () => {
    it('should generate salt with session prefix', async () => {
      const sessionId = 'session-123-abc';
      const result = await generateSessionBoundSalt(sessionId);

      expect(result.salt).toBeInstanceOf(Uint8Array);
      // Should be prefix (8 bytes) + base salt length
      expect(result.salt.length).toBe(8 + DEFAULT_SALT_LENGTH);
    });

    it('should include session hash in prefix', async () => {
      const result1 = await generateSessionBoundSalt('session-1');
      const result2 = await generateSessionBoundSalt('session-2');

      // First 8 bytes should differ based on session ID
      const prefix1 = result1.salt.slice(0, 8);
      const prefix2 = result2.salt.slice(0, 8);

      expect(bytesToHex(prefix1)).not.toBe(bytesToHex(prefix2));
    });
  });

  describe('verifySalt', () => {
    it('should return true for valid salt', () => {
      const { salt } = generateSalt();
      expect(verifySalt(salt)).toBe(true);
    });

    it('should return false for salt that is too short', () => {
      const shortSalt = new Uint8Array([1, 2, 3, 4, 5]);
      expect(verifySalt(shortSalt)).toBe(false);
    });

    it('should return false for all-zero salt', () => {
      const zeroSalt = new Uint8Array(32);
      expect(verifySalt(zeroSalt)).toBe(false);
    });

    it('should return false for all-same-value salt', () => {
      const sameSalt = new Uint8Array(32).fill(0xab);
      expect(verifySalt(sameSalt)).toBe(false);
    });

    it('should accept salt at minimum length (16 bytes)', () => {
      const salt = new Uint8Array(16);
      // Fill with different values
      for (let i = 0; i < salt.length; i++) {
        salt[i] = i;
      }
      expect(verifySalt(salt)).toBe(true);
    });
  });

  describe('bytesToHex / hexToBytes', () => {
    it('should round-trip correctly', () => {
      const original = new Uint8Array([0x00, 0x0f, 0xf0, 0xff, 0xab, 0xcd]);
      const hex = bytesToHex(original);
      const converted = hexToBytes(hex);

      expect(converted).toEqual(original);
    });

    it('should produce lowercase hex', () => {
      const bytes = new Uint8Array([0xab, 0xcd, 0xef]);
      const hex = bytesToHex(bytes);

      expect(hex).toBe('abcdef');
    });

    it('should handle empty arrays', () => {
      const empty = new Uint8Array(0);
      const hex = bytesToHex(empty);

      expect(hex).toBe('');
      expect(hexToBytes(hex)).toEqual(empty);
    });

    it('should pad single-digit hex values', () => {
      const bytes = new Uint8Array([0x01, 0x02, 0x03]);
      const hex = bytesToHex(bytes);

      expect(hex).toBe('010203');
    });
  });

  describe('bytesToBase64 / base64ToBytes', () => {
    it('should round-trip correctly', () => {
      const original = new Uint8Array([0x00, 0x0f, 0xf0, 0xff, 0xab, 0xcd]);
      const base64 = bytesToBase64(original);
      const converted = base64ToBytes(base64);

      expect(converted).toEqual(original);
    });

    it('should handle empty arrays', () => {
      const empty = new Uint8Array(0);
      const base64 = bytesToBase64(empty);

      expect(base64ToBytes(base64)).toEqual(empty);
    });

    it('should produce valid base64 string', () => {
      const bytes = new Uint8Array([0x48, 0x65, 0x6c, 0x6c, 0x6f]); // "Hello"
      const base64 = bytesToBase64(bytes);

      expect(base64).toBe('SGVsbG8=');
    });
  });

  describe('concatBytes', () => {
    it('should concatenate multiple arrays', () => {
      const a = new Uint8Array([1, 2, 3]);
      const b = new Uint8Array([4, 5]);
      const c = new Uint8Array([6, 7, 8, 9]);

      const result = concatBytes(a, b, c);

      expect(result).toEqual(new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9]));
    });

    it('should handle empty arrays', () => {
      const a = new Uint8Array([1, 2]);
      const empty = new Uint8Array(0);
      const b = new Uint8Array([3, 4]);

      const result = concatBytes(a, empty, b);

      expect(result).toEqual(new Uint8Array([1, 2, 3, 4]));
    });

    it('should handle single array', () => {
      const a = new Uint8Array([1, 2, 3]);
      const result = concatBytes(a);

      expect(result).toEqual(a);
    });

    it('should handle no arrays', () => {
      const result = concatBytes();
      expect(result).toEqual(new Uint8Array(0));
    });
  });
});
