/**
 * useIdentity Hook
 * VE-701: VEID onboarding, identity score display, re-verification prompts
 *
 * Provides access to identity status, score, and verification workflows.
 */

import {
  useState,
  useCallback,
  useEffect,
  useContext,
  createContext,
} from "react";
import type { ReactNode } from "react";
import type {
  IdentityState,
  IdentityStatus,
  IdentityScore,
  VerificationScope,
  UploadRecord,
  VerificationRecord,
  IdentityGatingError,
  RemediationPath,
  RemediationStep,
  ScopeRequirement,
  MarketplaceAction,
} from "../types/identity";
import { initialIdentityState, getTierFromScore } from "../types/identity";
import type { QueryClient } from "../types/chain";

/**
 * Identity context value
 */
interface IdentityContextValue {
  state: IdentityState;
  actions: IdentityActions;
}

/**
 * Identity actions
 */
interface IdentityActions {
  /**
   * Refresh identity data from chain
   */
  refresh: () => Promise<void>;

  /**
   * Check if user meets requirements for an action
   */
  checkRequirements: (action: MarketplaceAction) => IdentityGatingError | null;

  /**
   * Get remediation path for identity issues
   */
  getRemediationPath: (gatingError: IdentityGatingError) => RemediationPath;

  /**
   * Get scope requirements for an action
   */
  getScopeRequirements: (action: MarketplaceAction) => ScopeRequirement;

  /**
   * Clear identity error
   */
  clearError: () => void;
}

/**
 * Identity context
 */
const IdentityContext = createContext<IdentityContextValue | null>(null);

/**
 * Identity provider props
 */
export interface IdentityProviderProps {
  children: ReactNode;
  queryClient: QueryClient;
  accountAddress: string | null;
}

/**
 * Scope requirements by action
 */
const SCOPE_REQUIREMENTS: Record<MarketplaceAction, ScopeRequirement> = {
  browse_offerings: {
    action: "browse_offerings",
    minScore: 0,
    requiredScopes: [],
    optionalScopes: [],
    mfaRequired: false,
  },
  view_offering_details: {
    action: "view_offering_details",
    minScore: 0,
    requiredScopes: [],
    optionalScopes: [],
    mfaRequired: false,
  },
  place_order: {
    action: "place_order",
    minScore: 30,
    requiredScopes: ["email"],
    optionalScopes: ["id_document"],
    mfaRequired: false,
  },
  place_high_value_order: {
    action: "place_high_value_order",
    minScore: 60,
    requiredScopes: ["email", "id_document", "selfie"],
    optionalScopes: ["domain"],
    mfaRequired: true,
  },
  register_provider: {
    action: "register_provider",
    minScore: 70,
    requiredScopes: ["email", "id_document", "selfie", "domain"],
    optionalScopes: ["sso"],
    mfaRequired: true,
  },
  create_offering: {
    action: "create_offering",
    minScore: 70,
    requiredScopes: ["email", "id_document", "selfie", "domain"],
    optionalScopes: [],
    mfaRequired: true,
  },
  submit_hpc_job: {
    action: "submit_hpc_job",
    minScore: 50,
    requiredScopes: ["email", "id_document"],
    optionalScopes: ["selfie"],
    mfaRequired: true,
  },
  access_outputs: {
    action: "access_outputs",
    minScore: 30,
    requiredScopes: ["email"],
    optionalScopes: [],
    mfaRequired: false,
  },
};

/**
 * Identity provider component
 */
export function IdentityProvider({
  children,
  queryClient,
  accountAddress,
}: IdentityProviderProps) {
  const [state, setState] = useState<IdentityState>(initialIdentityState);

  /**
   * Fetch identity data from chain
   */
  const fetchIdentityData = useCallback(async () => {
    if (!accountAddress) {
      setState(initialIdentityState);
      return;
    }

    setState((prev) => ({ ...prev, isLoading: true }));

    try {
      // Query identity from chain
      const identityInfo = await queryClient.queryIdentity(accountAddress);

      // Parse status
      const status = identityInfo.status as IdentityStatus;

      // Build score object
      const score: IdentityScore | null =
        identityInfo.score > 0
          ? {
              value: identityInfo.score,
              tier: getTierFromScore(identityInfo.score),
              computedAt: identityInfo.updatedAt,
              modelVersion: identityInfo.modelVersion,
              breakdown: {
                document: Math.floor(identityInfo.score * 0.25),
                facial: Math.floor(identityInfo.score * 0.25),
                metadata: Math.floor(identityInfo.score * 0.25),
                trust: Math.floor(identityInfo.score * 0.25),
              },
              blockHeight: identityInfo.blockHeight,
            }
          : null;

      // Query scopes (would be a separate query in real impl)
      const completedScopes: VerificationScope[] = [];

      // Query upload history (would be a separate query)
      const uploadHistory: UploadRecord[] = [];

      // Query verification records (would be a separate query)
      const verificationRecords: VerificationRecord[] = [];

      setState({
        isLoading: false,
        status,
        score,
        completedScopes,
        uploadHistory,
        verificationRecords,
        error: null,
      });
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: {
          code: "network_error",
          message:
            error instanceof Error ? error.message : "Failed to fetch identity",
          remediations: ["Check your network connection", "Try again later"],
        },
      }));
    }
  }, [accountAddress, queryClient]);

  /**
   * Refresh identity data
   */
  const refresh = useCallback(async () => {
    await fetchIdentityData();
  }, [fetchIdentityData]);

  /**
   * Check if user meets requirements for an action
   */
  const checkRequirements = useCallback(
    (action: MarketplaceAction): IdentityGatingError | null => {
      const requirement = SCOPE_REQUIREMENTS[action];
      if (!requirement) {
        return null;
      }

      const currentScore = state.score?.value ?? 0;
      const completedScopeTypes = state.completedScopes.map((s) => s.type);

      // Check score
      if (currentScore < requirement.minScore) {
        const missingScopes = requirement.requiredScopes.filter(
          (s) => !completedScopeTypes.includes(s as any),
        );

        return {
          action,
          requiredScore: requirement.minScore,
          currentScore,
          requiredScopes: requirement.requiredScopes as any[],
          missingScopes: missingScopes as any[],
          remediation: getRemediationPath({
            action,
            requiredScore: requirement.minScore,
            currentScore,
            requiredScopes: requirement.requiredScopes as any[],
            missingScopes: missingScopes as any[],
            remediation: {} as any,
          }),
        };
      }

      // Check required scopes
      const missingScopes = requirement.requiredScopes.filter(
        (s) => !completedScopeTypes.includes(s as any),
      );

      if (missingScopes.length > 0) {
        return {
          action,
          requiredScore: requirement.minScore,
          currentScore,
          requiredScopes: requirement.requiredScopes as any[],
          missingScopes: missingScopes as any[],
          remediation: getRemediationPath({
            action,
            requiredScore: requirement.minScore,
            currentScore,
            requiredScopes: requirement.requiredScopes as any[],
            missingScopes: missingScopes as any[],
            remediation: {} as any,
          }),
        };
      }

      return null;
    },
    [state.score, state.completedScopes],
  );

  /**
   * Get remediation path for identity issues
   */
  const getRemediationPath = useCallback(
    (gatingError: IdentityGatingError): RemediationPath => {
      const steps: RemediationPath["steps"] = [];
      let order = 1;

      // Add steps for missing scopes
      for (const scope of gatingError.missingScopes) {
        steps.push({
          order: order++,
          title: `Complete ${scope} verification`,
          description: `Verify your ${scope} to increase your identity score`,
          action: { type: "upload_scope" as const, scopeType: scope },
          completed: false,
        });
      }

      // Add MFA step if needed
      const requirement =
        SCOPE_REQUIREMENTS[gatingError.action as MarketplaceAction];
      if (requirement?.mfaRequired) {
        steps.push({
          order: order++,
          title: "Set up MFA",
          description: "Enable multi-factor authentication for this action",
          action: { type: "complete_mfa" as const },
          completed: false,
        });
      }

      // Estimate time (5 minutes per scope + 2 for MFA)
      const estimatedTimeMinutes =
        gatingError.missingScopes.length * 5 +
        (requirement?.mfaRequired ? 2 : 0);

      return {
        steps,
        estimatedTimeMinutes,
        captureClientUrl: "https://capture.virtengine.com",
      };
    },
    [],
  );

  /**
   * Get scope requirements for an action
   */
  const getScopeRequirements = useCallback(
    (action: MarketplaceAction): ScopeRequirement => {
      return (
        SCOPE_REQUIREMENTS[action] ?? {
          action,
          minScore: 0,
          requiredScopes: [],
          optionalScopes: [],
          mfaRequired: false,
        }
      );
    },
    [],
  );

  /**
   * Clear error
   */
  const clearError = useCallback(() => {
    setState((prev) => ({ ...prev, error: null }));
  }, []);

  // Fetch identity data when account changes
  useEffect(() => {
    fetchIdentityData();
  }, [fetchIdentityData]);

  const actions: IdentityActions = {
    refresh,
    checkRequirements,
    getRemediationPath,
    getScopeRequirements,
    clearError,
  };

  return (
    <IdentityContext.Provider value={{ state, actions }}>
      {children}
    </IdentityContext.Provider>
  );
}

/**
 * Use identity hook
 */
export function useIdentity(): IdentityContextValue {
  const context = useContext(IdentityContext);
  if (!context) {
    throw new Error("useIdentity must be used within an IdentityProvider");
  }
  return context;
}
