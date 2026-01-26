/**
 * Wallet Utilities
 * VE-700: Wallet-based authentication
 *
 * CRITICAL: Private keys and mnemonics are NEVER stored or logged.
 * They are used only in memory for signing operations.
 */

import type { Wallet, WalletConfig, SigningResult, WalletType } from '../types/wallet';

/**
 * Default wallet configuration
 */
const DEFAULT_CONFIG: WalletConfig = {
  hdPath: "m/44'/118'/0'/0/0",
  prefix: 've',
  algorithm: 'secp256k1',
};

/**
 * Bech32 encoding utilities
 */
const BECH32_CHARSET = 'qpzry9x8gf2tvdw0s3jn54khce6mua7l';

function bech32Polymod(values: number[]): number {
  const GEN = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];
  let chk = 1;
  for (const v of values) {
    const b = chk >> 25;
    chk = ((chk & 0x1ffffff) << 5) ^ v;
    for (let i = 0; i < 5; i++) {
      chk ^= (b >> i) & 1 ? GEN[i] : 0;
    }
  }
  return chk;
}

function bech32HrpExpand(hrp: string): number[] {
  const ret: number[] = [];
  for (let i = 0; i < hrp.length; i++) {
    ret.push(hrp.charCodeAt(i) >> 5);
  }
  ret.push(0);
  for (let i = 0; i < hrp.length; i++) {
    ret.push(hrp.charCodeAt(i) & 31);
  }
  return ret;
}

function bech32Encode(hrp: string, data: number[]): string {
  const combined = [...data];
  const polymod = bech32Polymod([...bech32HrpExpand(hrp), ...combined, 0, 0, 0, 0, 0, 0]) ^ 1;
  for (let i = 0; i < 6; i++) {
    combined.push((polymod >> (5 * (5 - i))) & 31);
  }
  return hrp + '1' + combined.map(d => BECH32_CHARSET[d]).join('');
}

function convertBits(data: Uint8Array, fromBits: number, toBits: number, pad: boolean): number[] {
  let acc = 0;
  let bits = 0;
  const ret: number[] = [];
  const maxv = (1 << toBits) - 1;

  for (const value of data) {
    acc = (acc << fromBits) | value;
    bits += fromBits;
    while (bits >= toBits) {
      bits -= toBits;
      ret.push((acc >> bits) & maxv);
    }
  }

  if (pad && bits > 0) {
    ret.push((acc << (toBits - bits)) & maxv);
  }

  return ret;
}

/**
 * Create bech32 address from public key hash
 */
function createAddress(pubKeyHash: Uint8Array, prefix: string): string {
  const words = convertBits(pubKeyHash, 8, 5, true);
  return bech32Encode(prefix, words);
}

/**
 * Hash public key to get address bytes
 */
async function hashPubKey(pubKey: Uint8Array): Promise<Uint8Array> {
  // SHA-256 then RIPEMD-160 (simplified: just SHA-256 and take first 20 bytes)
  const sha256 = await crypto.subtle.digest('SHA-256', pubKey);
  return new Uint8Array(sha256).slice(0, 20);
}

/**
 * Base wallet adapter
 */
export abstract class WalletAdapter implements Wallet {
  protected config: WalletConfig;
  protected locked: boolean = false;

  constructor(config: Partial<WalletConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  abstract getAddress(): Promise<string>;
  abstract getPublicKey(): Promise<Uint8Array>;
  abstract sign(data: Uint8Array): Promise<SigningResult>;
  abstract getType(): WalletType;

  async signTransaction(txBytes: Uint8Array): Promise<SigningResult> {
    // Sign the hash of the transaction
    const hash = await crypto.subtle.digest('SHA-256', txBytes);
    return this.sign(new Uint8Array(hash));
  }

  async verify(data: Uint8Array, signature: Uint8Array): Promise<boolean> {
    // Verification would use the public key
    // Implementation depends on the algorithm
    throw new Error('Verification not implemented');
  }

  isLocked(): boolean {
    return this.locked;
  }

  lock(): void {
    this.locked = true;
  }
}

/**
 * Mnemonic wallet
 * CRITICAL: Mnemonic is only held in memory during operations
 */
export class MnemonicWallet extends WalletAdapter {
  private privateKey: Uint8Array | null = null;
  private publicKey: Uint8Array | null = null;
  private address: string | null = null;

  private constructor(config: Partial<WalletConfig> = {}) {
    super(config);
  }

  /**
   * Create wallet from mnemonic
   * CRITICAL: The mnemonic is used to derive keys and then discarded
   */
  static async fromMnemonic(
    mnemonic: string,
    hdPath?: string
  ): Promise<MnemonicWallet> {
    const wallet = new MnemonicWallet({ hdPath });

    // Validate mnemonic (basic check)
    const words = mnemonic.trim().split(/\s+/);
    if (![12, 15, 18, 21, 24].includes(words.length)) {
      throw new Error('Invalid mnemonic length');
    }

    // Derive keys from mnemonic
    // In a real implementation, this would use BIP39/BIP32 derivation
    // Here we simulate with a deterministic hash
    const seed = await crypto.subtle.digest(
      'SHA-256',
      new TextEncoder().encode(mnemonic + (hdPath || wallet.config.hdPath))
    );

    wallet.privateKey = new Uint8Array(seed);
    
    // Derive public key (simplified - real impl would use secp256k1)
    const pubKeyHash = await crypto.subtle.digest('SHA-256', wallet.privateKey);
    wallet.publicKey = new Uint8Array(pubKeyHash).slice(0, 33);

    // Generate address
    const addressHash = await hashPubKey(wallet.publicKey);
    wallet.address = createAddress(addressHash, wallet.config.prefix || 've');

    return wallet;
  }

  getType(): WalletType {
    return 'mnemonic';
  }

  async getAddress(): Promise<string> {
    if (this.locked || !this.address) {
      throw new Error('Wallet is locked');
    }
    return this.address;
  }

  async getPublicKey(): Promise<Uint8Array> {
    if (this.locked || !this.publicKey) {
      throw new Error('Wallet is locked');
    }
    return this.publicKey;
  }

  async sign(data: Uint8Array): Promise<SigningResult> {
    if (this.locked || !this.privateKey || !this.publicKey) {
      throw new Error('Wallet is locked');
    }

    // Simplified signing (real impl would use secp256k1 or ed25519)
    const key = await crypto.subtle.importKey(
      'raw',
      this.privateKey,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );

    const signature = await crypto.subtle.sign('HMAC', key, data);

    return {
      signature: new Uint8Array(signature),
      publicKey: this.publicKey,
      algorithm: this.config.algorithm || 'secp256k1',
    };
  }

  lock(): void {
    // Clear sensitive data from memory
    if (this.privateKey) {
      this.privateKey.fill(0);
      this.privateKey = null;
    }
    this.locked = true;
  }
}

/**
 * Keypair wallet (from raw private key)
 * CRITICAL: Private key is only held in memory during operations
 */
export class KeypairWallet extends WalletAdapter {
  private privateKey: Uint8Array | null = null;
  private publicKey: Uint8Array | null = null;
  private address: string | null = null;

  private constructor(config: Partial<WalletConfig> = {}) {
    super(config);
  }

  /**
   * Create wallet from private key
   * CRITICAL: The private key is stored in memory only
   */
  static async fromPrivateKey(
    privateKey: Uint8Array,
    config?: Partial<WalletConfig>
  ): Promise<KeypairWallet> {
    const wallet = new KeypairWallet(config);

    if (privateKey.length !== 32) {
      throw new Error('Invalid private key length');
    }

    wallet.privateKey = new Uint8Array(privateKey);

    // Derive public key
    const pubKeyHash = await crypto.subtle.digest('SHA-256', wallet.privateKey);
    wallet.publicKey = new Uint8Array(pubKeyHash).slice(0, 33);

    // Generate address
    const addressHash = await hashPubKey(wallet.publicKey);
    wallet.address = createAddress(addressHash, wallet.config.prefix || 've');

    return wallet;
  }

  getType(): WalletType {
    return 'keypair';
  }

  async getAddress(): Promise<string> {
    if (this.locked || !this.address) {
      throw new Error('Wallet is locked');
    }
    return this.address;
  }

  async getPublicKey(): Promise<Uint8Array> {
    if (this.locked || !this.publicKey) {
      throw new Error('Wallet is locked');
    }
    return this.publicKey;
  }

  async sign(data: Uint8Array): Promise<SigningResult> {
    if (this.locked || !this.privateKey || !this.publicKey) {
      throw new Error('Wallet is locked');
    }

    const key = await crypto.subtle.importKey(
      'raw',
      this.privateKey,
      { name: 'HMAC', hash: 'SHA-256' },
      false,
      ['sign']
    );

    const signature = await crypto.subtle.sign('HMAC', key, data);

    return {
      signature: new Uint8Array(signature),
      publicKey: this.publicKey,
      algorithm: this.config.algorithm || 'secp256k1',
    };
  }

  lock(): void {
    if (this.privateKey) {
      this.privateKey.fill(0);
      this.privateKey = null;
    }
    this.locked = true;
  }
}
