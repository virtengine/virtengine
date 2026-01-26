/**
 * Authentication Types
 * VE-700: Portal authentication, wallets, and session model
 *
 * @packageDocumentation
 */

/**
 * Authentication state
 */
export interface AuthState {
  /**
   * Whether authentication is in progress
   */
  isLoading: boolean;

  /**
   * Whether user is authenticated
   */
  isAuthenticated: boolean;

  /**
   * Current authentication error
   */
  error: AuthError | null;

  /**
   * Authenticated account address
   */
  accountAddress: string | null;

  /**
   * Account public key (hex encoded)
   */
  publicKey: string | null;

  /**
   * Authentication method used
   */
  authMethod: AuthMethod | null;

  /**
   * Session information
   */
  session: SessionInfo | null;
}

/**
 * Authentication method
 */
export type AuthMethod = 'wallet' | 'sso';

/**
 * Session information (non-sensitive)
 */
export interface SessionInfo {
  /**
   * Session ID (not the token itself)
   */
  sessionId: string;

  /**
   * When the session was created
   */
  createdAt: number;

  /**
   * When the session expires
   */
  expiresAt: number;

  /**
   * Whether session is trusted browser
   */
  isTrustedBrowser: boolean;

  /**
   * Device fingerprint (non-identifying hash)
   */
  deviceFingerprint: string;
}

/**
 * Authentication actions
 */
export interface AuthActions {
  /**
   * Login with wallet credentials
   */
  loginWithWallet: (credentials: WalletCredentials) => Promise<void>;

  /**
   * Login with SSO
   */
  loginWithSSO: (credentials: SSOCredentials) => Promise<void>;

  /**
   * Logout and invalidate session
   */
  logout: () => Promise<void>;

  /**
   * Refresh session token
   */
  refreshSession: () => Promise<void>;

  /**
   * Sign a message with the current wallet
   */
  signMessage: (message: Uint8Array) => Promise<Uint8Array>;

  /**
   * Sign a transaction
   */
  signTransaction: (txBytes: Uint8Array) => Promise<Uint8Array>;

  /**
   * Clear any authentication errors
   */
  clearError: () => void;
}

/**
 * Wallet credentials for login
 * NOTE: These are never stored or logged
 */
export interface WalletCredentials {
  /**
   * Wallet type
   */
  type: 'mnemonic' | 'keypair' | 'hardware' | 'extension';

  /**
   * Mnemonic phrase (for type: 'mnemonic')
   * SENSITIVE: Never log or persist
   */
  mnemonic?: string;

  /**
   * Private key bytes (for type: 'keypair')
   * SENSITIVE: Never log or persist
   */
  privateKey?: Uint8Array;

  /**
   * Hardware wallet config (for type: 'hardware')
   */
  hardwareConfig?: HardwareWalletConfig;

  /**
   * Extension wallet ID (for type: 'extension')
   */
  extensionId?: string;

  /**
   * HD derivation path
   * @default "m/44'/118'/0'/0/0"
   */
  hdPath?: string;
}

/**
 * Hardware wallet configuration
 */
export interface HardwareWalletConfig {
  /**
   * Hardware wallet type
   */
  type: 'ledger' | 'trezor';

  /**
   * Transport type
   */
  transport: 'usb' | 'bluetooth' | 'webusb';

  /**
   * Account index
   */
  accountIndex?: number;
}

/**
 * SSO credentials for login
 */
export interface SSOCredentials {
  /**
   * Authorization code from SSO provider
   */
  authorizationCode: string;

  /**
   * State parameter for CSRF protection
   */
  state: string;

  /**
   * Code verifier for PKCE
   */
  codeVerifier: string;

  /**
   * Nonce for replay protection
   */
  nonce: string;
}

/**
 * Session token (stored in httpOnly cookie, never exposed to JS)
 * This type represents the structure but the actual token is never accessible
 */
export interface SessionToken {
  /**
   * Token version for rotation
   */
  version: number;

  /**
   * Session ID
   */
  sessionId: string;

  /**
   * Account address
   */
  accountAddress: string;

  /**
   * Issued at timestamp
   */
  iat: number;

  /**
   * Expiration timestamp
   */
  exp: number;

  /**
   * Token signature (computed server-side)
   */
  signature: string;
}

/**
 * Authentication error
 */
export interface AuthError {
  /**
   * Error code
   */
  code: AuthErrorCode;

  /**
   * Human-readable message
   */
  message: string;

  /**
   * Additional details (non-sensitive)
   */
  details?: Record<string, unknown>;
}

/**
 * Authentication error codes
 */
export type AuthErrorCode =
  | 'invalid_credentials'
  | 'invalid_mnemonic'
  | 'invalid_signature'
  | 'session_expired'
  | 'session_invalid'
  | 'sso_error'
  | 'sso_binding_failed'
  | 'network_error'
  | 'hardware_wallet_error'
  | 'extension_not_found'
  | 'user_cancelled'
  | 'rate_limited'
  | 'unknown';

/**
 * Initial authentication state
 */
export const initialAuthState: AuthState = {
  isLoading: false,
  isAuthenticated: false,
  error: null,
  accountAddress: null,
  publicKey: null,
  authMethod: null,
  session: null,
};

/**
 * Authentication event types
 */
export type AuthEvent =
  | { type: 'AUTH_START' }
  | { type: 'AUTH_SUCCESS'; payload: { accountAddress: string; publicKey: string; method: AuthMethod; session: SessionInfo } }
  | { type: 'AUTH_FAILURE'; payload: AuthError }
  | { type: 'AUTH_LOGOUT' }
  | { type: 'SESSION_REFRESH'; payload: SessionInfo }
  | { type: 'SESSION_EXPIRED' }
  | { type: 'CLEAR_ERROR' };

/**
 * Reduce authentication state
 */
export function authReducer(state: AuthState, event: AuthEvent): AuthState {
  switch (event.type) {
    case 'AUTH_START':
      return {
        ...state,
        isLoading: true,
        error: null,
      };

    case 'AUTH_SUCCESS':
      return {
        ...state,
        isLoading: false,
        isAuthenticated: true,
        error: null,
        accountAddress: event.payload.accountAddress,
        publicKey: event.payload.publicKey,
        authMethod: event.payload.method,
        session: event.payload.session,
      };

    case 'AUTH_FAILURE':
      return {
        ...state,
        isLoading: false,
        isAuthenticated: false,
        error: event.payload,
      };

    case 'AUTH_LOGOUT':
      return initialAuthState;

    case 'SESSION_REFRESH':
      return {
        ...state,
        session: event.payload,
      };

    case 'SESSION_EXPIRED':
      return {
        ...initialAuthState,
        error: {
          code: 'session_expired',
          message: 'Your session has expired. Please log in again.',
        },
      };

    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };

    default:
      return state;
  }
}
