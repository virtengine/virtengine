/**
 * Wallet Types
 * VE-700: Wallet-based authentication
 *
 * @packageDocumentation
 */

/**
 * Wallet configuration
 */
export interface WalletConfig {
  /**
   * HD derivation path
   * @default "m/44'/118'/0'/0/0"
   */
  hdPath?: string;

  /**
   * Address prefix
   * @default 've'
   */
  prefix?: string;

  /**
   * Key algorithm
   * @default 'secp256k1'
   */
  algorithm?: 'secp256k1' | 'ed25519';
}

/**
 * Signing result
 */
export interface SigningResult {
  /**
   * Signature bytes
   */
  signature: Uint8Array;

  /**
   * Public key bytes
   */
  publicKey: Uint8Array;

  /**
   * Signature algorithm used
   */
  algorithm: string;
}

/**
 * Wallet interface
 */
export interface Wallet {
  /**
   * Get the wallet address
   */
  getAddress(): Promise<string>;

  /**
   * Get the public key
   */
  getPublicKey(): Promise<Uint8Array>;

  /**
   * Sign arbitrary bytes
   */
  sign(data: Uint8Array): Promise<SigningResult>;

  /**
   * Sign a transaction
   */
  signTransaction(txBytes: Uint8Array): Promise<SigningResult>;

  /**
   * Verify a signature
   */
  verify(data: Uint8Array, signature: Uint8Array): Promise<boolean>;

  /**
   * Get wallet type
   */
  getType(): WalletType;

  /**
   * Whether wallet is locked
   */
  isLocked(): boolean;

  /**
   * Lock the wallet (clear sensitive data from memory)
   */
  lock(): void;
}

/**
 * Wallet types
 */
export type WalletType =
  | 'mnemonic'
  | 'keypair'
  | 'hardware'
  | 'extension';

/**
 * Hardware wallet interface
 */
export interface HardwareWallet extends Wallet {
  /**
   * Connect to hardware wallet
   */
  connect(): Promise<void>;

  /**
   * Disconnect from hardware wallet
   */
  disconnect(): Promise<void>;

  /**
   * Whether hardware wallet is connected
   */
  isConnected(): boolean;

  /**
   * Get device info
   */
  getDeviceInfo(): Promise<HardwareDeviceInfo>;
}

/**
 * Hardware device info
 */
export interface HardwareDeviceInfo {
  /**
   * Device model
   */
  model: string;

  /**
   * Firmware version
   */
  firmwareVersion: string;

  /**
   * Device ID
   */
  deviceId: string;
}

/**
 * Extension wallet interface
 */
export interface ExtensionWallet extends Wallet {
  /**
   * Extension ID
   */
  extensionId: string;

  /**
   * Extension name
   */
  extensionName: string;

  /**
   * Request connection to extension
   */
  connect(): Promise<void>;

  /**
   * Whether extension is installed
   */
  isInstalled(): boolean;
}
