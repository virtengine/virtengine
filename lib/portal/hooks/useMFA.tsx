/**
 * useMFA Hook
 * VE-702: MFA enrollment, policy configuration, trusted browser UX
 *
 * Provides MFA enrollment, policy management, and challenge handling.
 */

import { useState, useCallback, useEffect, useContext, createContext } from 'react';
import type { ReactNode } from 'react';
import type {
  MFAState,
  MFAFactor,
  MFAFactorType,
  MFAPolicy,
  MFAEnrollment,
  TrustedBrowser,
  MFAChallenge,
  MFAChallengeResponse,
  MFAAuditEntry,
  MFAError,
  SensitiveTransactionType,
} from '../types/mfa';
import { initialMFAState, getFactorDisplayName } from '../types/mfa';

/**
 * MFA context value
 */
interface MFAContextValue {
  state: MFAState;
  actions: MFAActions;
}

/**
 * MFA actions
 */
interface MFAActions {
  /**
   * Refresh MFA data
   */
  refresh: () => Promise<void>;

  /**
   * Start enrollment for a factor type
   */
  startEnrollment: (type: MFAFactorType) => Promise<MFAEnrollment>;

  /**
   * Complete enrollment with verification
   */
  completeEnrollment: (enrollment: MFAEnrollment, verificationCode: string) => Promise<MFAFactor>;

  /**
   * Cancel enrollment
   */
  cancelEnrollment: () => void;

  /**
   * Remove a factor
   */
  removeFactor: (factorId: string) => Promise<void>;

  /**
   * Set primary factor
   */
  setPrimaryFactor: (factorId: string) => Promise<void>;

  /**
   * Update MFA policy
   */
  updatePolicy: (policy: Partial<MFAPolicy>) => Promise<void>;

  /**
   * Trust current browser
   */
  trustBrowser: (deviceName: string) => Promise<TrustedBrowser>;

  /**
   * Revoke trusted browser
   */
  revokeTrustedBrowser: (browserId: string) => Promise<void>;

  /**
   * Create MFA challenge for sensitive action
   */
  createChallenge: (transactionType: SensitiveTransactionType) => Promise<MFAChallenge>;

  /**
   * Verify MFA challenge
   */
  verifyChallenge: (challengeId: string, factorId: string, code: string) => Promise<MFAChallengeResponse>;

  /**
   * Verify FIDO2 challenge
   */
  verifyFido2Challenge: (challengeId: string, factorId: string, credential: PublicKeyCredential) => Promise<MFAChallengeResponse>;

  /**
   * Check if transaction type requires MFA
   */
  requiresMFA: (transactionType: SensitiveTransactionType) => boolean;

  /**
   * Clear error
   */
  clearError: () => void;
}

/**
 * MFA context
 */
const MFAContext = createContext<MFAContextValue | null>(null);

/**
 * MFA provider props
 */
export interface MFAProviderProps {
  children: ReactNode;
  apiEndpoint: string;
  accountAddress: string | null;
  getAuthHeader: () => Promise<string>;
}

/**
 * MFA provider component
 */
export function MFAProvider({
  children,
  apiEndpoint,
  accountAddress,
  getAuthHeader,
}: MFAProviderProps) {
  const [state, setState] = useState<MFAState>(initialMFAState);
  const [currentEnrollment, setCurrentEnrollment] = useState<MFAEnrollment | null>(null);

  /**
   * Make authenticated API request
   */
  const apiRequest = useCallback(async <T,>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> => {
    const authHeader = await getAuthHeader();
    const response = await fetch(`${apiEndpoint}${path}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': authHeader,
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.message || 'Request failed');
    }

    return response.json();
  }, [apiEndpoint, getAuthHeader]);

  /**
   * Fetch MFA data
   */
  const fetchMFAData = useCallback(async () => {
    if (!accountAddress) {
      setState(initialMFAState);
      return;
    }

    setState(prev => ({ ...prev, isLoading: true }));

    try {
      const [factors, policy, browsers, audit] = await Promise.all([
        apiRequest<MFAFactor[]>('/mfa/factors'),
        apiRequest<MFAPolicy>('/mfa/policy'),
        apiRequest<TrustedBrowser[]>('/mfa/trusted-browsers'),
        apiRequest<MFAAuditEntry[]>('/mfa/audit?limit=50'),
      ]);

      setState({
        isLoading: false,
        isEnabled: factors.length > 0,
        enrolledFactors: factors,
        policy,
        trustedBrowsers: browsers,
        activeChallenge: null,
        auditHistory: audit,
        error: null,
      });
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: {
          code: 'network_error',
          message: error instanceof Error ? error.message : 'Failed to fetch MFA data',
        },
      }));
    }
  }, [accountAddress, apiRequest]);

  /**
   * Refresh MFA data
   */
  const refresh = useCallback(async () => {
    await fetchMFAData();
  }, [fetchMFAData]);

  /**
   * Start enrollment for a factor type
   */
  const startEnrollment = useCallback(async (type: MFAFactorType): Promise<MFAEnrollment> => {
    const response = await apiRequest<{
      challengeData: MFAEnrollment['challengeData'];
    }>('/mfa/enroll/start', {
      method: 'POST',
      body: JSON.stringify({ type }),
    });

    const enrollment: MFAEnrollment = {
      type,
      step: 'configure',
      challengeData: response.challengeData,
    };

    setCurrentEnrollment(enrollment);
    return enrollment;
  }, [apiRequest]);

  /**
   * Complete enrollment with verification
   */
  const completeEnrollment = useCallback(async (
    enrollment: MFAEnrollment,
    verificationCode: string
  ): Promise<MFAFactor> => {
    const factor = await apiRequest<MFAFactor>('/mfa/enroll/complete', {
      method: 'POST',
      body: JSON.stringify({
        type: enrollment.type,
        verificationCode,
      }),
    });

    setCurrentEnrollment(null);
    await refresh();

    return factor;
  }, [apiRequest, refresh]);

  /**
   * Cancel enrollment
   */
  const cancelEnrollment = useCallback(() => {
    setCurrentEnrollment(null);
  }, []);

  /**
   * Remove a factor
   */
  const removeFactor = useCallback(async (factorId: string) => {
    await apiRequest(`/mfa/factors/${factorId}`, {
      method: 'DELETE',
    });
    await refresh();
  }, [apiRequest, refresh]);

  /**
   * Set primary factor
   */
  const setPrimaryFactor = useCallback(async (factorId: string) => {
    await apiRequest(`/mfa/factors/${factorId}/primary`, {
      method: 'PUT',
    });
    await refresh();
  }, [apiRequest, refresh]);

  /**
   * Update MFA policy
   */
  const updatePolicy = useCallback(async (policyUpdate: Partial<MFAPolicy>) => {
    await apiRequest('/mfa/policy', {
      method: 'PUT',
      body: JSON.stringify(policyUpdate),
    });
    await refresh();
  }, [apiRequest, refresh]);

  /**
   * Trust current browser
   */
  const trustBrowser = useCallback(async (deviceName: string): Promise<TrustedBrowser> => {
    const browser = await apiRequest<TrustedBrowser>('/mfa/trusted-browsers', {
      method: 'POST',
      body: JSON.stringify({ deviceName }),
    });
    await refresh();
    return browser;
  }, [apiRequest, refresh]);

  /**
   * Revoke trusted browser
   */
  const revokeTrustedBrowser = useCallback(async (browserId: string) => {
    await apiRequest(`/mfa/trusted-browsers/${browserId}`, {
      method: 'DELETE',
    });
    await refresh();
  }, [apiRequest, refresh]);

  /**
   * Create MFA challenge for sensitive action
   */
  const createChallenge = useCallback(async (
    transactionType: SensitiveTransactionType
  ): Promise<MFAChallenge> => {
    const challenge = await apiRequest<MFAChallenge>('/mfa/challenge', {
      method: 'POST',
      body: JSON.stringify({ transactionType }),
    });

    setState(prev => ({ ...prev, activeChallenge: challenge }));
    return challenge;
  }, [apiRequest]);

  /**
   * Verify MFA challenge
   */
  const verifyChallenge = useCallback(async (
    challengeId: string,
    factorId: string,
    code: string
  ): Promise<MFAChallengeResponse> => {
    try {
      const response = await apiRequest<MFAChallengeResponse>('/mfa/challenge/verify', {
        method: 'POST',
        body: JSON.stringify({ challengeId, factorId, code }),
      });

      if (response.verified) {
        setState(prev => ({ ...prev, activeChallenge: null }));
      }

      return response;
    } catch (error) {
      const mfaError: MFAError = {
        code: 'verification_failed',
        message: error instanceof Error ? error.message : 'Verification failed',
      };
      setState(prev => ({ ...prev, error: mfaError }));
      throw error;
    }
  }, [apiRequest]);

  /**
   * Verify FIDO2 challenge
   */
  const verifyFido2Challenge = useCallback(async (
    challengeId: string,
    factorId: string,
    credential: PublicKeyCredential
  ): Promise<MFAChallengeResponse> => {
    const response = credential.response as AuthenticatorAssertionResponse;

    const result = await apiRequest<MFAChallengeResponse>('/mfa/challenge/verify-fido2', {
      method: 'POST',
      body: JSON.stringify({
        challengeId,
        factorId,
        credentialId: credential.id,
        authenticatorData: btoa(String.fromCharCode(...new Uint8Array(response.authenticatorData))),
        clientDataJSON: btoa(String.fromCharCode(...new Uint8Array(response.clientDataJSON))),
        signature: btoa(String.fromCharCode(...new Uint8Array(response.signature))),
      }),
    });

    if (result.verified) {
      setState(prev => ({ ...prev, activeChallenge: null }));
    }

    return result;
  }, [apiRequest]);

  /**
   * Check if transaction type requires MFA
   */
  const requiresMFA = useCallback((transactionType: SensitiveTransactionType): boolean => {
    if (!state.policy) {
      return false;
    }
    return state.policy.sensitiveTransactions.includes(transactionType);
  }, [state.policy]);

  /**
   * Clear error
   */
  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Fetch MFA data when account changes
  useEffect(() => {
    fetchMFAData();
  }, [fetchMFAData]);

  const actions: MFAActions = {
    refresh,
    startEnrollment,
    completeEnrollment,
    cancelEnrollment,
    removeFactor,
    setPrimaryFactor,
    updatePolicy,
    trustBrowser,
    revokeTrustedBrowser,
    createChallenge,
    verifyChallenge,
    verifyFido2Challenge,
    requiresMFA,
    clearError,
  };

  return (
    <MFAContext.Provider value={{ state, actions }}>
      {children}
    </MFAContext.Provider>
  );
}

/**
 * Use MFA hook
 */
export function useMFA(): MFAContextValue {
  const context = useContext(MFAContext);
  if (!context) {
    throw new Error('useMFA must be used within an MFAProvider');
  }
  return context;
}
