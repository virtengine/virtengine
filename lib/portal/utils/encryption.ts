/**
 * Encryption Utilities
 * VE-700: Encrypt/decrypt payloads for secure communication
 *
 * CRITICAL: Never log or expose plaintext sensitive data
 */

/**
 * Encryption result
 */
export interface EncryptionResult {
  /**
   * Encrypted data (base64)
   */
  ciphertext: string;

  /**
   * Initialization vector (base64)
   */
  iv: string;

  /**
   * Ephemeral public key (base64) - for ECDH
   */
  ephemeralPublicKey?: string;

  /**
   * Encryption algorithm used
   */
  algorithm: string;

  /**
   * Key derivation info (non-sensitive)
   */
  kdfInfo?: string;
}

/**
 * Decryption result
 */
export interface DecryptionResult {
  /**
   * Decrypted data
   */
  plaintext: Uint8Array;

  /**
   * Whether decryption was successful
   */
  success: boolean;
}

/**
 * Encryption options
 */
interface EncryptOptions {
  /**
   * Algorithm to use
   */
  algorithm?: 'AES-GCM' | 'AES-CBC';

  /**
   * Additional authenticated data (for GCM mode)
   */
  aad?: Uint8Array;
}

/**
 * Generate a random IV for encryption
 */
function generateIV(length: number = 12): Uint8Array {
  return crypto.getRandomValues(new Uint8Array(length));
}

/**
 * Derive encryption key from shared secret
 */
async function deriveKey(
  sharedSecret: Uint8Array,
  salt: Uint8Array,
  info: string = 'virtengine-encryption'
): Promise<CryptoKey> {
  // Import shared secret as raw key material
  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    sharedSecret as BufferSource,
    'HKDF',
    false,
    ['deriveKey']
  );

  // Derive AES key using HKDF
  return crypto.subtle.deriveKey(
    {
      name: 'HKDF',
      hash: 'SHA-256',
      salt: salt as BufferSource,
      info: new TextEncoder().encode(info) as BufferSource,
    },
    keyMaterial,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
}

/**
 * Encrypt payload to a recipient's public key
 *
 * Uses ECDH key agreement + AES-GCM encryption
 */
export async function encryptPayload(
  plaintext: Uint8Array,
  recipientPublicKey: Uint8Array,
  options: EncryptOptions = {}
): Promise<EncryptionResult> {
  const { algorithm = 'AES-GCM', aad } = options;

  // Generate ephemeral key pair for ECDH
  const ephemeralKeyPair = await crypto.subtle.generateKey(
    { name: 'ECDH', namedCurve: 'P-256' },
    true,
    ['deriveKey', 'deriveBits']
  );

  // Import recipient's public key
  const recipientKey = await crypto.subtle.importKey(
    'raw',
    recipientPublicKey as BufferSource,
    { name: 'ECDH', namedCurve: 'P-256' },
    false,
    []
  );

  // Perform ECDH to get shared secret
  const sharedBits = await crypto.subtle.deriveBits(
    { name: 'ECDH', public: recipientKey },
    ephemeralKeyPair.privateKey,
    256
  );
  const sharedSecret = new Uint8Array(sharedBits);

  // Generate IV and salt
  const iv = generateIV();
  const salt = crypto.getRandomValues(new Uint8Array(32));

  // Derive encryption key
  const encryptionKey = await deriveKey(sharedSecret, salt);

  // Encrypt the plaintext
  const ciphertext = await crypto.subtle.encrypt(
    {
      name: algorithm,
      iv: iv as BufferSource,
      additionalData: aad ? (aad as BufferSource) : undefined,
    },
    encryptionKey,
    plaintext as BufferSource
  );

  // Export ephemeral public key
  const ephemeralPublicKey = await crypto.subtle.exportKey(
    'raw',
    ephemeralKeyPair.publicKey
  );

  return {
    ciphertext: btoa(String.fromCharCode(...new Uint8Array(ciphertext))),
    iv: btoa(String.fromCharCode(...iv)),
    ephemeralPublicKey: btoa(String.fromCharCode(...new Uint8Array(ephemeralPublicKey))),
    algorithm,
    kdfInfo: btoa(String.fromCharCode(...salt)),
  };
}

/**
 * Decrypt payload using private key
 */
export async function decryptPayload(
  encrypted: EncryptionResult,
  privateKey: CryptoKey,
  options: EncryptOptions = {}
): Promise<DecryptionResult> {
  const { aad } = options;

  try {
    // Decode base64 values
    const ciphertext = Uint8Array.from(atob(encrypted.ciphertext), c => c.charCodeAt(0));
    const iv = Uint8Array.from(atob(encrypted.iv), c => c.charCodeAt(0));
    const ephemeralPublicKey = encrypted.ephemeralPublicKey
      ? Uint8Array.from(atob(encrypted.ephemeralPublicKey), c => c.charCodeAt(0))
      : null;
    const salt = encrypted.kdfInfo
      ? Uint8Array.from(atob(encrypted.kdfInfo), c => c.charCodeAt(0))
      : new Uint8Array(32);

    if (!ephemeralPublicKey) {
      throw new Error('Missing ephemeral public key');
    }

    // Import ephemeral public key
    const ephemeralKey = await crypto.subtle.importKey(
      'raw',
      ephemeralPublicKey as BufferSource,
      { name: 'ECDH', namedCurve: 'P-256' },
      false,
      []
    );

    // Perform ECDH to get shared secret
    const sharedBits = await crypto.subtle.deriveBits(
      { name: 'ECDH', public: ephemeralKey },
      privateKey,
      256
    );
    const sharedSecret = new Uint8Array(sharedBits);

    // Derive decryption key
    const decryptionKey = await deriveKey(sharedSecret, salt);

    // Decrypt
    const plaintext = await crypto.subtle.decrypt(
      {
        name: encrypted.algorithm || 'AES-GCM',
        iv: iv as BufferSource,
        additionalData: aad ? (aad as BufferSource) : undefined,
      },
      decryptionKey,
      ciphertext as BufferSource
    );

    return {
      plaintext: new Uint8Array(plaintext),
      success: true,
    };
  } catch (error) {
    return {
      plaintext: new Uint8Array(0),
      success: false,
    };
  }
}

/**
 * Encrypt data with a symmetric key (for local encryption)
 */
export async function encryptWithKey(
  plaintext: Uint8Array,
  key: CryptoKey
): Promise<{ ciphertext: Uint8Array; iv: Uint8Array }> {
  const iv = generateIV();

  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: iv as BufferSource },
    key,
    plaintext as BufferSource
  );

  return {
    ciphertext: new Uint8Array(ciphertext),
    iv,
  };
}

/**
 * Decrypt data with a symmetric key
 */
export async function decryptWithKey(
  ciphertext: Uint8Array,
  iv: Uint8Array,
  key: CryptoKey
): Promise<Uint8Array> {
  const plaintext = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: iv as BufferSource },
    key,
    ciphertext as BufferSource
  );

  return new Uint8Array(plaintext);
}

/**
 * Generate a random symmetric key
 */
export async function generateSymmetricKey(): Promise<CryptoKey> {
  return crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    true,
    ['encrypt', 'decrypt']
  );
}
