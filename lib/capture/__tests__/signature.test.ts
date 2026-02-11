/**
 * Signature Tests
 * VE-210: Unit tests for signature creation utilities
 */

import { describe, it, expect } from 'vitest';
import {
  computeHash,
  createPayloadHash,
  createClientSignature,
  createUserSignature,
  createSignaturePackage,
  serializeSignaturePackage,
  createSessionId,
} from '../utils/signature';
import type { ClientKeyProvider, UserKeyProvider, CaptureMetadata } from '../types/capture';

describe('signature', () => {
  // Mock key providers
  const mockClientKeyProvider: ClientKeyProvider = {
    getClientId: async () => 'test-client-id',
    getClientVersion: async () => '1.0.0',
    sign: async (data: Uint8Array) => {
      // Simple mock signature: just hash the data
      const crypto = globalThis.crypto || require('crypto').webcrypto;
      const hash = await crypto.subtle.digest('SHA-256', data);
      return new Uint8Array(hash);
    },
    getPublicKey: async () => new Uint8Array(32).fill(0x01),
    getKeyType: async () => 'ed25519',
  };

  const mockUserKeyProvider: UserKeyProvider = {
    getAccountAddress: async () => 've1abc123xyz',
    sign: async (data: Uint8Array) => {
      const crypto = globalThis.crypto || require('crypto').webcrypto;
      const hash = await crypto.subtle.digest('SHA-256', data);
      return new Uint8Array(hash);
    },
    getPublicKey: async () => new Uint8Array(32).fill(0x02),
    getKeyType: async () => 'ed25519',
  };

  // Mock metadata
  const mockMetadata: CaptureMetadata = {
    deviceFingerprint: 'mock-fingerprint-123',
    clientVersion: '1.0.0',
    capturedAt: '2026-01-24T12:00:00.000Z',
    documentType: 'id_card',
    qualityScore: 85,
    sessionId: 'session-123',
  };

  // Mock image blob
  function createMockImageBlob(): Blob {
    const data = new Uint8Array(1000);
    for (let i = 0; i < data.length; i++) {
      data[i] = i % 256;
    }
    return new Blob([data], { type: 'image/jpeg' });
  }

  describe('computeHash', () => {
    it('should compute SHA-256 hash by default', async () => {
      const data = new TextEncoder().encode('Hello, World!');
      const hash = await computeHash(data);

      expect(hash).toBeInstanceOf(Uint8Array);
      expect(hash.length).toBe(32); // SHA-256 produces 32 bytes
    });

    it('should compute SHA-384 hash when specified', async () => {
      const data = new TextEncoder().encode('Hello, World!');
      const hash = await computeHash(data, 'SHA-384');

      expect(hash.length).toBe(48); // SHA-384 produces 48 bytes
    });

    it('should compute SHA-512 hash when specified', async () => {
      const data = new TextEncoder().encode('Hello, World!');
      const hash = await computeHash(data, 'SHA-512');

      expect(hash.length).toBe(64); // SHA-512 produces 64 bytes
    });

    it('should produce consistent hashes for same input', async () => {
      const data = new TextEncoder().encode('Consistent input');
      const hash1 = await computeHash(data);
      const hash2 = await computeHash(data);

      expect(hash1).toEqual(hash2);
    });

    it('should produce different hashes for different input', async () => {
      const data1 = new TextEncoder().encode('Input 1');
      const data2 = new TextEncoder().encode('Input 2');
      const hash1 = await computeHash(data1);
      const hash2 = await computeHash(data2);

      expect(hash1).not.toEqual(hash2);
    });
  });

  describe('createPayloadHash', () => {
    it('should create hash from image blob', async () => {
      const blob = createMockImageBlob();
      const hash = await createPayloadHash(blob, mockMetadata);

      expect(hash).toBeInstanceOf(Uint8Array);
      expect(hash.length).toBe(32);
    });

    it('should include metadata in hash by default', async () => {
      const blob = createMockImageBlob();

      const hashWith = await createPayloadHash(blob, mockMetadata, { includeMetadata: true });
      const hashWithout = await createPayloadHash(blob, mockMetadata, { includeMetadata: false });

      expect(hashWith).not.toEqual(hashWithout);
    });

    it('should produce different hashes for different metadata', async () => {
      const blob = createMockImageBlob();

      const metadata1 = { ...mockMetadata, qualityScore: 80 };
      const metadata2 = { ...mockMetadata, qualityScore: 90 };

      const hash1 = await createPayloadHash(blob, metadata1);
      const hash2 = await createPayloadHash(blob, metadata2);

      expect(hash1).not.toEqual(hash2);
    });

    it('should support different hash algorithms', async () => {
      const blob = createMockImageBlob();

      const hash256 = await createPayloadHash(blob, mockMetadata, { hashAlgorithm: 'SHA-256' });
      const hash512 = await createPayloadHash(blob, mockMetadata, { hashAlgorithm: 'SHA-512' });

      expect(hash256.length).toBe(32);
      expect(hash512.length).toBe(64);
    });
  });

  describe('createClientSignature', () => {
    it('should create signature with client ID and version', async () => {
      const salt = new Uint8Array(32).fill(0xab);
      const payloadHash = new Uint8Array(32).fill(0xcd);

      const result = await createClientSignature(mockClientKeyProvider, salt, payloadHash);

      expect(result.signature).toBeInstanceOf(Uint8Array);
      expect(result.clientId).toBe('test-client-id');
      expect(result.clientVersion).toBe('1.0.0');
    });

    it('should sign over salt and payload hash', async () => {
      const salt1 = new Uint8Array(32).fill(0xaa);
      const salt2 = new Uint8Array(32).fill(0xbb);
      const payloadHash = new Uint8Array(32).fill(0xcc);

      const result1 = await createClientSignature(mockClientKeyProvider, salt1, payloadHash);
      const result2 = await createClientSignature(mockClientKeyProvider, salt2, payloadHash);

      // Different salts should produce different signatures
      expect(result1.signature).not.toEqual(result2.signature);
    });
  });

  describe('createUserSignature', () => {
    it('should create signature with user address', async () => {
      const salt = new Uint8Array(32).fill(0xab);
      const payloadHash = new Uint8Array(32).fill(0xcd);
      const clientSignature = new Uint8Array(32).fill(0xef);

      const result = await createUserSignature(
        mockUserKeyProvider,
        salt,
        payloadHash,
        clientSignature
      );

      expect(result.signature).toBeInstanceOf(Uint8Array);
      expect(result.userAddress).toBe('ve1abc123xyz');
    });

    it('should include client signature in signed data', async () => {
      const salt = new Uint8Array(32).fill(0xab);
      const payloadHash = new Uint8Array(32).fill(0xcd);
      const clientSig1 = new Uint8Array(32).fill(0x11);
      const clientSig2 = new Uint8Array(32).fill(0x22);

      const result1 = await createUserSignature(mockUserKeyProvider, salt, payloadHash, clientSig1);
      const result2 = await createUserSignature(mockUserKeyProvider, salt, payloadHash, clientSig2);

      // Different client signatures should produce different user signatures
      expect(result1.signature).not.toEqual(result2.signature);
    });
  });

  describe('createSignaturePackage', () => {
    it('should create complete signature package', async () => {
      const blob = createMockImageBlob();
      const salt = new Uint8Array(32).fill(0xab);

      const pkg = await createSignaturePackage(
        blob,
        mockMetadata,
        salt,
        mockClientKeyProvider,
        mockUserKeyProvider
      );

      expect(pkg.salt).toEqual(salt);
      expect(pkg.payloadHash).toBeInstanceOf(Uint8Array);
      expect(pkg.clientSignature).toBeInstanceOf(Uint8Array);
      expect(pkg.userSignature).toBeInstanceOf(Uint8Array);
      expect(pkg.clientId).toBe('test-client-id');
      expect(pkg.clientVersion).toBe('1.0.0');
      expect(pkg.userAddress).toBe('ve1abc123xyz');
      expect(pkg.signedAt).toBeDefined();
    });

    it('should include valid ISO timestamp', async () => {
      const blob = createMockImageBlob();
      const salt = new Uint8Array(32).fill(0xab);

      const pkg = await createSignaturePackage(
        blob,
        mockMetadata,
        salt,
        mockClientKeyProvider,
        mockUserKeyProvider
      );

      // Should be valid ISO date
      const date = new Date(pkg.signedAt);
      expect(date.toString()).not.toBe('Invalid Date');
    });
  });

  describe('serializeSignaturePackage', () => {
    it('should serialize package to JSON string', async () => {
      const blob = createMockImageBlob();
      const salt = new Uint8Array(32).fill(0xab);

      const pkg = await createSignaturePackage(
        blob,
        mockMetadata,
        salt,
        mockClientKeyProvider,
        mockUserKeyProvider
      );

      const serialized = serializeSignaturePackage(pkg);

      expect(typeof serialized).toBe('string');

      // Should be valid JSON
      const parsed = JSON.parse(serialized);
      expect(parsed.clientId).toBe('test-client-id');
      expect(parsed.userAddress).toBe('ve1abc123xyz');
    });

    it('should convert byte arrays to hex', async () => {
      const blob = createMockImageBlob();
      const salt = new Uint8Array(32).fill(0xab);

      const pkg = await createSignaturePackage(
        blob,
        mockMetadata,
        salt,
        mockClientKeyProvider,
        mockUserKeyProvider
      );

      const serialized = serializeSignaturePackage(pkg);
      const parsed = JSON.parse(serialized);

      // Salt should be hex string
      expect(typeof parsed.salt).toBe('string');
      expect(parsed.salt).toBe('ab'.repeat(32));
    });
  });

  describe('createSessionId', () => {
    it('should create session ID with cap_ prefix', () => {
      const sessionId = createSessionId();

      expect(sessionId.startsWith('cap_')).toBe(true);
    });

    it('should create unique session IDs', () => {
      const ids = new Set<string>();
      for (let i = 0; i < 100; i++) {
        ids.add(createSessionId());
      }

      expect(ids.size).toBe(100);
    });

    it('should include timestamp component', () => {
      const sessionId = createSessionId();
      const parts = sessionId.split('_');

      expect(parts.length).toBe(3);
      // Second part is timestamp in base36
      expect(parts[1].length).toBeGreaterThan(0);
    });

    it('should include random component', () => {
      const sessionId = createSessionId();
      const parts = sessionId.split('_');

      // Third part is random hex
      expect(parts[2].length).toBe(16);
      expect(/^[0-9a-f]+$/.test(parts[2])).toBe(true);
    });
  });
});
