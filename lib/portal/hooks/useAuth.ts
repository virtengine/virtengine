/**
 * useAuth Hook
 * VE-700: Authentication, wallets, and session management
 *
 * Provides wallet-based and SSO authentication with secure session handling.
 * Sessions use httpOnly cookies and are rotated automatically.
 */

import { useState, useCallback, useEffect, useContext, createContext, useReducer } from 'react';
import type { ReactNode } from 'react';
import type {
  AuthState,
  AuthActions,
  AuthEvent,
  WalletCredentials,
  SSOCredentials,
  SessionInfo,
  AuthError,
} from '../types/auth';
import { initialAuthState, authReducer } from '../types/auth';
import { SessionManager } from '../utils/session';
import { consumeOAuthRequest } from '../utils/oidc';
import { MnemonicWallet, KeypairWallet, WalletAdapter } from '../utils/wallet';
import type { Wallet } from '../types/wallet';

/**
 * Auth context value
 */
interface AuthContextValue {
  state: AuthState;
  actions: AuthActions;
  wallet: Wallet | null;
}

/**
 * Auth context
 */
const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * Auth provider props
 */
export interface AuthProviderProps {
  children: ReactNode;
  sessionManager: SessionManager;
  chainEndpoint: string;
  ssoConfig?: {
    authorizationEndpoint: string;
    tokenEndpoint: string;
    clientId: string;
    redirectUri: string;
    accountBindingEndpoint: string;
    stateStorageKey?: string;
    enforceState?: boolean;
    enforcePKCE?: boolean;
  };
  onSessionExpired?: () => void;
}

/**
 * Auth provider component
 */
export function AuthProvider({
  children,
  sessionManager,
  chainEndpoint,
  ssoConfig,
  onSessionExpired,
}: AuthProviderProps) {
  const [state, dispatch] = useReducer(authReducer, initialAuthState);
  const [wallet, setWallet] = useState<Wallet | null>(null);

  /**
   * Create wallet from credentials
   * CRITICAL: Credentials are never stored or logged
   */
  const createWallet = useCallback(async (credentials: WalletCredentials): Promise<Wallet> => {
    switch (credentials.type) {
      case 'mnemonic':
        if (!credentials.mnemonic) {
          throw new Error('Mnemonic is required');
        }
        return MnemonicWallet.fromMnemonic(
          credentials.mnemonic,
          credentials.hdPath
        );

      case 'keypair':
        if (!credentials.privateKey) {
          throw new Error('Private key is required');
        }
        return KeypairWallet.fromPrivateKey(credentials.privateKey);

      case 'hardware':
        // Hardware wallet integration would go here
        throw new Error('Hardware wallet not yet implemented');

      case 'extension':
        // Extension wallet integration would go here
        throw new Error('Extension wallet not yet implemented');

      default:
        throw new Error('Unknown wallet type');
    }
  }, []);

  /**
   * Create signed session token
   */
  const createSessionToken = useCallback(async (
    wallet: Wallet,
    accountAddress: string
  ): Promise<SessionInfo> => {
    // Create session challenge
    const challenge = await sessionManager.createChallenge(accountAddress);

    // Sign the challenge with wallet
    const signResult = await wallet.sign(new TextEncoder().encode(challenge.message));

    // Submit signed challenge to create session
    const session = await sessionManager.createSession({
      accountAddress,
      challenge: challenge.message,
      signature: signResult.signature,
      publicKey: signResult.publicKey,
    });

    return session;
  }, [sessionManager]);

  /**
   * Login with wallet credentials
   */
  const loginWithWallet = useCallback(async (credentials: WalletCredentials) => {
    dispatch({ type: 'AUTH_START' });

    try {
      // Create wallet (credentials cleared after this)
      const newWallet = await createWallet(credentials);
      const accountAddress = await newWallet.getAddress();
      const publicKey = await newWallet.getPublicKey();

      // Create session
      const session = await createSessionToken(newWallet, accountAddress);

      // Store wallet reference
      setWallet(newWallet);

      // Dispatch success
      dispatch({
        type: 'AUTH_SUCCESS',
        payload: {
          accountAddress,
          publicKey: Buffer.from(publicKey).toString('hex'),
          method: 'wallet',
          session,
        },
      });
    } catch (error) {
      const authError: AuthError = {
        code: 'invalid_credentials',
        message: error instanceof Error ? error.message : 'Authentication failed',
      };
      dispatch({ type: 'AUTH_FAILURE', payload: authError });
    }
  }, [createWallet, createSessionToken]);

  /**
   * Login with SSO
   */
  const loginWithSSO = useCallback(async (credentials: SSOCredentials) => {
    if (!ssoConfig) {
      throw new Error('SSO is not configured');
    }

    dispatch({ type: 'AUTH_START' });

    try {
      const storedRequest = consumeOAuthRequest(
        credentials.state,
        ssoConfig.stateStorageKey
      );

      if (ssoConfig.enforceState && !storedRequest) {
        throw new Error('Invalid or missing OAuth state');
      }

      if (storedRequest) {
        if (ssoConfig.enforcePKCE && storedRequest.codeVerifier !== credentials.codeVerifier) {
          throw new Error('Invalid PKCE verifier');
        }
        if (storedRequest.nonce !== credentials.nonce) {
          throw new Error('Invalid nonce');
        }
      }

      // Exchange authorization code for tokens
      const tokenResponse = await fetch(ssoConfig.tokenEndpoint, {
        method: 'POST',
        credentials: 'omit',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: credentials.authorizationCode,
          redirect_uri: ssoConfig.redirectUri,
          client_id: ssoConfig.clientId,
          code_verifier: credentials.codeVerifier,
        }),
      });

      if (!tokenResponse.ok) {
        throw new Error('Token exchange failed');
      }

      const tokens = await tokenResponse.json();

      // Bind SSO identity to blockchain account
      const bindingResponse = await fetch(ssoConfig.accountBindingEndpoint, {
        method: 'POST',
        credentials: 'omit',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${tokens.access_token}`,
        },
        body: JSON.stringify({ nonce: credentials.nonce }),
      });

      if (!bindingResponse.ok) {
        throw new Error('Account binding failed');
      }

      const binding = await bindingResponse.json();

      // Create session from binding
      const session = await sessionManager.createSessionFromSSO({
        ssoToken: tokens.access_token,
        accountAddress: binding.accountAddress,
        publicKey: binding.publicKey,
      });

      dispatch({
        type: 'AUTH_SUCCESS',
        payload: {
          accountAddress: binding.accountAddress,
          publicKey: binding.publicKey,
          method: 'sso',
          session,
        },
      });
    } catch (error) {
      const authError: AuthError = {
        code: 'sso_error',
        message: error instanceof Error ? error.message : 'SSO authentication failed',
      };
      dispatch({ type: 'AUTH_FAILURE', payload: authError });
    }
  }, [ssoConfig, sessionManager]);

  /**
   * Logout and invalidate session
   */
  const logout = useCallback(async () => {
    try {
      await sessionManager.invalidateSession();
    } catch (error) {
      // Log error but continue with logout
      console.error('Session invalidation failed');
    }

    // Clear wallet
    if (wallet) {
      wallet.lock();
      setWallet(null);
    }

    dispatch({ type: 'AUTH_LOGOUT' });
  }, [wallet, sessionManager]);

  /**
   * Refresh session token
   */
  const refreshSession = useCallback(async () => {
    try {
      const newSession = await sessionManager.refreshSession();
      dispatch({ type: 'SESSION_REFRESH', payload: newSession });
    } catch (error) {
      dispatch({ type: 'SESSION_EXPIRED' });
      onSessionExpired?.();
    }
  }, [sessionManager, onSessionExpired]);

  /**
   * Sign a message with current wallet
   */
  const signMessage = useCallback(async (message: Uint8Array): Promise<Uint8Array> => {
    if (!wallet) {
      throw new Error('No wallet connected');
    }
    const result = await wallet.sign(message);
    return result.signature;
  }, [wallet]);

  /**
   * Sign a transaction
   */
  const signTransaction = useCallback(async (txBytes: Uint8Array): Promise<Uint8Array> => {
    if (!wallet) {
      throw new Error('No wallet connected');
    }
    const result = await wallet.signTransaction(txBytes);
    return result.signature;
  }, [wallet]);

  /**
   * Clear error
   */
  const clearError = useCallback(() => {
    dispatch({ type: 'CLEAR_ERROR' });
  }, []);

  // Check for existing session on mount
  useEffect(() => {
    const checkExistingSession = async () => {
      try {
        const existingSession = await sessionManager.getSession();
        if (existingSession) {
          dispatch({
            type: 'AUTH_SUCCESS',
            payload: {
              accountAddress: existingSession.accountAddress,
              publicKey: existingSession.publicKey,
              method: existingSession.authMethod as 'wallet' | 'sso',
              session: {
                sessionId: existingSession.sessionId,
                createdAt: existingSession.createdAt,
                expiresAt: existingSession.expiresAt,
                isTrustedBrowser: existingSession.isTrustedBrowser,
                deviceFingerprint: existingSession.deviceFingerprint,
              },
            },
          });
        }
      } catch (error) {
        // No existing session, stay logged out
      }
    };

    checkExistingSession();
  }, [sessionManager]);

  // Setup automatic session refresh
  useEffect(() => {
    if (!state.isAuthenticated || !state.session) {
      return;
    }

    const refreshThreshold = 5 * 60 * 1000; // 5 minutes before expiry
    const timeUntilRefresh = (state.session.expiresAt * 1000) - Date.now() - refreshThreshold;

    if (timeUntilRefresh <= 0) {
      refreshSession();
      return;
    }

    const timeout = setTimeout(refreshSession, timeUntilRefresh);
    return () => clearTimeout(timeout);
  }, [state.isAuthenticated, state.session, refreshSession]);

  const actions: AuthActions = {
    loginWithWallet,
    loginWithSSO,
    logout,
    refreshSession,
    signMessage,
    signTransaction,
    clearError,
  };

  return (
    <AuthContext.Provider value={{ state, actions, wallet }}>
      {children}
    </AuthContext.Provider>
  );
}

/**
 * Use auth hook
 */
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
