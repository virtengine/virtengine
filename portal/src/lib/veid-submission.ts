import nacl from 'tweetnacl';
import { env } from '@/config/env';
import type { WalletContextValue } from '@/lib/portal-adapter';
import { signAndBroadcastAmino, type WalletSigner } from '@/lib/api/chain';
import type {
  ClientKeyProvider,
  UserKeyProvider,
  TxBroadcaster,
  UploadScopeMessage,
} from '@/lib/capture-adapter';

function base64ToBytes(base64: string): Uint8Array {
  if (typeof atob !== 'undefined') {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }
  return new Uint8Array(Buffer.from(base64, 'base64'));
}

function bytesToBase64(bytes: Uint8Array): string {
  if (typeof btoa !== 'undefined') {
    return btoa(String.fromCharCode(...bytes));
  }
  return Buffer.from(bytes).toString('base64');
}

function getCaptureClientConfig() {
  const clientId = env.captureClientId;
  const clientVersion = env.captureClientVersion;
  const privateKey = env.captureClientPrivateKey;

  if (!clientId || !privateKey) {
    throw new Error('Capture client credentials are not configured');
  }

  const keyBytes = base64ToBytes(privateKey);
  if (keyBytes.length !== 32 && keyBytes.length !== 64) {
    throw new Error('Capture client private key must be 32 or 64 bytes');
  }
  const keyPair =
    keyBytes.length === 32
      ? nacl.sign.keyPair.fromSeed(keyBytes)
      : nacl.sign.keyPair.fromSecretKey(keyBytes);

  return { clientId, clientVersion, keyPair };
}

export function createPortalClientKeyProvider(): ClientKeyProvider {
  const config = getCaptureClientConfig();

  return {
    getClientId: () => Promise.resolve(config.clientId),
    getClientVersion: () => Promise.resolve(config.clientVersion),
    sign: (data: Uint8Array) => Promise.resolve(nacl.sign.detached(data, config.keyPair.secretKey)),
    getPublicKey: () => Promise.resolve(config.keyPair.publicKey),
    getKeyType: () => Promise.resolve('ed25519' as const),
  };
}

export function createWalletUserKeyProvider(wallet: WalletContextValue): UserKeyProvider {
  return {
    getAccountAddress: () => {
      if (wallet.status !== 'connected') {
        throw new Error('Wallet is not connected');
      }
      const account = wallet.accounts[wallet.activeAccountIndex];
      if (!account) {
        throw new Error('No active wallet account');
      }
      return Promise.resolve(account.address);
    },
    sign: async (data: Uint8Array) => {
      if (wallet.status !== 'connected') {
        throw new Error('Wallet is not connected');
      }
      const account = wallet.accounts[wallet.activeAccountIndex];
      if (!account) {
        throw new Error('No active wallet account');
      }
      if (!wallet.signArbitrary) {
        throw new Error('Wallet does not support arbitrary signing');
      }
      const response = await wallet.signArbitrary(data);
      return base64ToBytes(response.signature);
    },
    getPublicKey: () => {
      if (wallet.status !== 'connected') {
        throw new Error('Wallet is not connected');
      }
      const account = wallet.accounts[wallet.activeAccountIndex];
      if (!account) {
        throw new Error('No active wallet account');
      }
      return Promise.resolve(account.pubKey);
    },
    getKeyType: () => {
      if (wallet.status !== 'connected') {
        throw new Error('Wallet is not connected');
      }
      const account = wallet.accounts[wallet.activeAccountIndex];
      if (!account) {
        throw new Error('No active wallet account');
      }
      return Promise.resolve(account.algo as 'ed25519' | 'secp256k1');
    },
  };
}

function encodeBytes(value: Uint8Array | undefined): string {
  return value ? bytesToBase64(value) : '';
}

function encodeEnvelope(message: UploadScopeMessage) {
  const envelope = message.value.encryptedPayload;
  return {
    version: envelope.version,
    algorithmId: envelope.algorithmId,
    algorithmVersion: envelope.algorithmVersion,
    recipientKeyIds: envelope.recipientKeyIds,
    recipientPublicKeys: (envelope.recipientPublicKeys ?? []).map((key) => bytesToBase64(key)),
    encryptedKeys: (envelope.encryptedKeys ?? []).map((key) => bytesToBase64(key)),
    nonce: bytesToBase64(envelope.nonce),
    ciphertext: bytesToBase64(envelope.ciphertext),
    senderSignature: bytesToBase64(envelope.senderSignature),
    senderPubKey: bytesToBase64(envelope.senderPubKey),
    metadata: envelope.metadata ?? {},
  };
}

function toAminoUploadScopeMessage(message: UploadScopeMessage): UploadScopeMessage {
  return {
    typeUrl: message.typeUrl,
    value: {
      ...message.value,
      encryptedPayload: encodeEnvelope(message),
      salt: encodeBytes(message.value.salt),
      clientSignature: encodeBytes(message.value.clientSignature),
      userSignature: encodeBytes(message.value.userSignature),
      payloadHash: encodeBytes(message.value.payloadHash),
    },
  } as UploadScopeMessage;
}

export function createPortalTxBroadcaster(wallet: WalletSigner): TxBroadcaster {
  return {
    broadcast: async (msg: UploadScopeMessage, memo?: string) => {
      const aminoMsg = toAminoUploadScopeMessage(msg);
      const result = await signAndBroadcastAmino(wallet, [aminoMsg], memo ?? 'Submit VEID scope');
      return {
        txHash: result.txHash,
        code: result.code,
        rawLog: result.rawLog,
        gasUsed: result.gasUsed,
        gasWanted: result.gasWanted,
      };
    },
  };
}
