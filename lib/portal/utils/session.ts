/**
 * Session Manager
 * VE-700: Secure session handling (httpOnly cookies, rotation, logout)
 *
 * CRITICAL: Session tokens are stored in httpOnly cookies and are never
 * accessible to JavaScript. This module interacts with the session API
 * to manage session state.
 */

/**
 * Session configuration
 */
export interface SessionConfig {
  /**
   * Session API endpoint
   */
  apiEndpoint: string;

  /**
   * Session token lifetime in seconds
   */
  tokenLifetimeSeconds: number;

  /**
   * Refresh threshold in seconds (rotate when this much time remains)
   */
  refreshThresholdSeconds: number;

  /**
   * Whether to enable automatic refresh
   */
  autoRefresh: boolean;

  /**
   * CSRF token header name
   */
  csrfHeader: string;
}

/**
 * Session info (public, non-sensitive)
 */
export interface SessionInfo {
  /**
   * Session ID (not the token)
   */
  sessionId: string;

  /**
   * Account address
   */
  accountAddress: string;

  /**
   * Public key (hex)
   */
  publicKey: string;

  /**
   * Auth method used
   */
  authMethod: string;

  /**
   * Creation timestamp
   */
  createdAt: number;

  /**
   * Expiration timestamp
   */
  expiresAt: number;

  /**
   * Whether this is a trusted browser
   */
  isTrustedBrowser: boolean;

  /**
   * Device fingerprint hash
   */
  deviceFingerprint: string;
}

/**
 * Session challenge
 */
export interface SessionChallenge {
  /**
   * Challenge ID
   */
  challengeId: string;

  /**
   * Message to sign
   */
  message: string;

  /**
   * Challenge expiration
   */
  expiresAt: number;
}

/**
 * Create session request
 */
export interface CreateSessionRequest {
  /**
   * Account address
   */
  accountAddress: string;

  /**
   * Signed challenge message
   */
  challenge: string;

  /**
   * Signature bytes
   */
  signature: Uint8Array;

  /**
   * Public key bytes
   */
  publicKey: Uint8Array;
}

/**
 * Create session from SSO request
 */
export interface CreateSessionFromSSORequest {
  /**
   * SSO access token
   */
  ssoToken: string;

  /**
   * Bound account address
   */
  accountAddress: string;

  /**
   * Public key (hex)
   */
  publicKey: string;
}

/**
 * Default session configuration
 */
export const defaultSessionConfig: SessionConfig = {
  apiEndpoint: '/api/session',
  tokenLifetimeSeconds: 3600,
  refreshThresholdSeconds: 300,
  autoRefresh: true,
  csrfHeader: 'X-CSRF-Token',
};

/**
 * Generate device fingerprint (non-identifying hash)
 */
async function generateDeviceFingerprint(): Promise<string> {
  const components = [
    navigator.userAgent,
    navigator.language,
    screen.width,
    screen.height,
    screen.colorDepth,
    new Date().getTimezoneOffset(),
    navigator.hardwareConcurrency || 0,
  ];

  const data = components.join('|');
  const encoder = new TextEncoder();
  const hashBuffer = await crypto.subtle.digest('SHA-256', encoder.encode(data));
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('').slice(0, 32);
}

/**
 * Session manager for secure session handling
 */
export class SessionManager {
  private config: SessionConfig;
  private csrfToken: string | null = null;
  private refreshTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(config: Partial<SessionConfig> = {}) {
    this.config = { ...defaultSessionConfig, ...config };
  }

  /**
   * Get CSRF token for requests
   */
  private async getCSRFToken(): Promise<string> {
    if (this.csrfToken) {
      return this.csrfToken;
    }

    const response = await fetch(`${this.config.apiEndpoint}/csrf`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to get CSRF token');
    }

    const data = await response.json();
    this.csrfToken = data.token;
    return this.csrfToken!;
  }

  /**
   * Make authenticated request with CSRF token
   */
  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const csrfToken = await this.getCSRFToken();

    const response = await fetch(`${this.config.apiEndpoint}${path}`, {
      ...options,
      credentials: 'include', // Include httpOnly cookies
      headers: {
        'Content-Type': 'application/json',
        [this.config.csrfHeader]: csrfToken,
        ...options.headers,
      },
    });

    if (response.status === 403) {
      // CSRF token may be stale, refresh and retry
      this.csrfToken = null;
      const newCsrfToken = await this.getCSRFToken();

      const retryResponse = await fetch(`${this.config.apiEndpoint}${path}`, {
        ...options,
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          [this.config.csrfHeader]: newCsrfToken,
          ...options.headers,
        },
      });

      if (!retryResponse.ok) {
        throw new Error(`Request failed: ${retryResponse.status}`);
      }

      return retryResponse.json();
    }

    if (!response.ok) {
      throw new Error(`Request failed: ${response.status}`);
    }

    return response.json();
  }

  /**
   * Create session challenge
   */
  async createChallenge(accountAddress: string): Promise<SessionChallenge> {
    const fingerprint = await generateDeviceFingerprint();

    return this.request<SessionChallenge>('/challenge', {
      method: 'POST',
      body: JSON.stringify({
        accountAddress,
        deviceFingerprint: fingerprint,
      }),
    });
  }

  /**
   * Create session with signed challenge
   * The session token is set in an httpOnly cookie by the server
   */
  async createSession(request: CreateSessionRequest): Promise<SessionInfo> {
    const fingerprint = await generateDeviceFingerprint();

    const session = await this.request<SessionInfo>('/create', {
      method: 'POST',
      body: JSON.stringify({
        accountAddress: request.accountAddress,
        challenge: request.challenge,
        signature: btoa(String.fromCharCode(...request.signature)),
        publicKey: btoa(String.fromCharCode(...request.publicKey)),
        deviceFingerprint: fingerprint,
      }),
    });

    // Setup auto-refresh if enabled
    if (this.config.autoRefresh) {
      this.setupAutoRefresh(session.expiresAt);
    }

    return session;
  }

  /**
   * Create session from SSO token
   */
  async createSessionFromSSO(request: CreateSessionFromSSORequest): Promise<SessionInfo> {
    const fingerprint = await generateDeviceFingerprint();

    const session = await this.request<SessionInfo>('/create-sso', {
      method: 'POST',
      body: JSON.stringify({
        ssoToken: request.ssoToken,
        accountAddress: request.accountAddress,
        publicKey: request.publicKey,
        deviceFingerprint: fingerprint,
      }),
    });

    if (this.config.autoRefresh) {
      this.setupAutoRefresh(session.expiresAt);
    }

    return session;
  }

  /**
   * Get current session info
   */
  async getSession(): Promise<SessionInfo | null> {
    try {
      return await this.request<SessionInfo>('/info');
    } catch {
      return null;
    }
  }

  /**
   * Refresh session token
   * The new token is set in an httpOnly cookie by the server
   */
  async refreshSession(): Promise<SessionInfo> {
    const session = await this.request<SessionInfo>('/refresh', {
      method: 'POST',
    });

    if (this.config.autoRefresh) {
      this.setupAutoRefresh(session.expiresAt);
    }

    return session;
  }

  /**
   * Invalidate session (logout)
   */
  async invalidateSession(): Promise<void> {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }

    await this.request('/invalidate', {
      method: 'POST',
    });

    this.csrfToken = null;
  }

  /**
   * Setup automatic session refresh
   */
  private setupAutoRefresh(expiresAt: number): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    const timeUntilRefresh = (expiresAt * 1000) - Date.now() - (this.config.refreshThresholdSeconds * 1000);

    if (timeUntilRefresh > 0) {
      this.refreshTimer = setTimeout(async () => {
        try {
          await this.refreshSession();
        } catch {
          // Session expired or invalid, user will be logged out
        }
      }, timeUntilRefresh);
    }
  }

  /**
   * Cleanup resources
   */
  destroy(): void {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }
  }
}
