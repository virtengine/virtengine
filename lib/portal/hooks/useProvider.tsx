/**
 * useProvider Hook
 * VE-704: Provider console (offerings, pricing, capacity, domain verification)
 */

import {
  useState,
  useCallback,
  useEffect,
  useContext,
  createContext,
  useRef,
} from "react";
import type { ReactNode } from "react";
import type {
  ProviderState,
  ProviderProfile,
  ProviderRegistration,
  DomainVerification,
  DomainChallenge,
  OfferingDraft,
  ProviderOffering,
  IncomingOrder,
  BidRecord,
  AllocationRecord,
  UsageRecord,
  SettlementSummary,
} from "../types/provider";
import { initialProviderState } from "../types/provider";
import type { QueryClient } from "../types/chain";
import {
  ProviderDomainVerificationError,
  normalizeProviderDomain,
  requireProviderDomainVerifier,
  validateProviderDomainChallenge,
  validateProviderDomainVerification,
  type ProviderDomainChallenge,
  type ProviderDomainVerificationEvidence,
  type ProviderDomainVerifier,
} from "../components/provider/domain-verification";

/**
 * Provider context value
 */
export interface ProviderContextValue {
  state: ProviderState;
  actions: ProviderActions;
}

/**
 * Provider actions
 */
export interface ProviderActions {
  refresh: () => Promise<void>;
  startRegistration: () => void;
  updateRegistrationData: (data: Partial<ProviderRegistration>) => void;
  startDomainVerification: (
    domain: string,
    method: "dns_txt" | "http_file",
  ) => Promise<ProviderDomainChallenge>;
  checkDomainVerification: (
    domain: string,
  ) => Promise<ProviderDomainVerificationEvidence>;
  submitRegistration: () => Promise<void>;
  createOffering: (draft: OfferingDraft) => Promise<ProviderOffering>;
  updateOffering: (
    offeringId: string,
    updates: Partial<OfferingDraft>,
  ) => Promise<ProviderOffering>;
  publishOffering: (offeringId: string) => Promise<void>;
  pauseOffering: (offeringId: string) => Promise<void>;
  getIncomingOrders: () => Promise<void>;
  getActiveBids: () => Promise<void>;
  getActiveAllocations: () => Promise<void>;
  getUsageRecords: (allocationId?: string) => Promise<void>;
  getSettlementSummary: () => Promise<void>;
  clearError: () => void;
}

const ProviderContext = createContext<ProviderContextValue | null>(null);

export interface ProviderProviderProps {
  children: ReactNode;
  queryClient: QueryClient;
  chainId: string;
  accountAddress: string | null;
  getAuthHeader?: () => Promise<string>;
  domainVerifier?: ProviderDomainVerifier;
}

export function ProviderProvider({
  children,
  queryClient,
  chainId,
  accountAddress,
  getAuthHeader,
  domainVerifier,
}: ProviderProviderProps) {
  const [state, setState] = useState<ProviderState>(initialProviderState);
  const verifierGeneration = useRef(0);
  const registrationGeneration = useRef(0);
  const verifierAuthority = useRef({ domainVerifier, chainId, accountAddress });
  const accountAuthority = useRef({ chainId, accountAddress });
  const accountGeneration = useRef(0);
  const stateAccountGeneration = useRef(0);
  const accountResetPending = useRef(false);
  const challenges = useRef(
    new Map<
      string,
      { challenge: ProviderDomainChallenge; registrationGeneration: number }
    >(),
  );
  const operationSequence = useRef(0);
  const challengeIssuanceInFlight = useRef(new Map<string, number>());
  const verificationsInFlight = useRef(new Map<string, number>());
  const registrationDomain = useRef<string | null>(null);
  if (
    accountAuthority.current.chainId !== chainId ||
    accountAuthority.current.accountAddress !== accountAddress
  ) {
    accountAuthority.current = { chainId, accountAddress };
    accountGeneration.current += 1;
    accountResetPending.current = true;
  }
  if (
    verifierAuthority.current.domainVerifier !== domainVerifier ||
    verifierAuthority.current.chainId !== chainId ||
    verifierAuthority.current.accountAddress !== accountAddress
  ) {
    verifierAuthority.current = { domainVerifier, chainId, accountAddress };
    verifierGeneration.current += 1;
    challenges.current.clear();
    challengeIssuanceInFlight.current.clear();
    verificationsInFlight.current.clear();
  }
  const renderVerifierGeneration = verifierGeneration.current;
  const renderAccountGeneration = accountGeneration.current;
  const effectiveState =
    stateAccountGeneration.current === renderAccountGeneration
      ? state
      : initialProviderState;

  useEffect(() => {
    if (!accountResetPending.current) return;
    accountResetPending.current = false;
    stateAccountGeneration.current = renderAccountGeneration;
    registrationGeneration.current += 1;
    challenges.current.clear();
    challengeIssuanceInFlight.current.clear();
    verificationsInFlight.current.clear();
    setState(initialProviderState);
  }, [renderAccountGeneration]);

  const fetchProviderData = useCallback(async () => {
    if (!accountAddress) {
      setState(initialProviderState);
      return;
    }

    const generation = renderAccountGeneration;
    const requestedAccount = accountAddress;
    setState((prev) => ({ ...prev, isLoading: true }));

    try {
      const providerInfo = await queryClient.queryProvider(accountAddress);
      if (
        generation !== accountGeneration.current ||
        accountAuthority.current.accountAddress !== requestedAccount
      ) {
        return;
      }

      if (providerInfo) {
        const profile: ProviderProfile = {
          address: providerInfo.address,
          name: "",
          description: "",
          website: "",
          verifiedDomains: [],
          status: providerInfo.status as any,
          identityScore: 0,
          reliabilityScore: providerInfo.reliabilityScore,
          registeredAt: providerInfo.registeredAt,
          offeringsCount: 0,
          ordersFulfilled: 0,
          tier: "bronze",
          stakedAmount: "0",
        };

        setState((prev) => ({
          ...prev,
          isLoading: false,
          isRegistered: true,
          profile,
          error: null,
        }));
      } else {
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isRegistered: false,
          error: null,
        }));
      }
    } catch (error) {
      if (
        generation !== accountGeneration.current ||
        accountAuthority.current.accountAddress !== requestedAccount
      ) {
        return;
      }
      setState((prev) => ({
        ...prev,
        isLoading: false,
        isRegistered: false,
        error: null,
      }));
    }
  }, [accountAddress, queryClient, renderAccountGeneration]);

  const assertAccountAuthority = useCallback(() => {
    if (
      !accountAddress ||
      renderAccountGeneration !== accountGeneration.current
    ) {
      throw new ProviderDomainVerificationError("authority_changed");
    }
  }, [accountAddress, renderAccountGeneration]);

  const refresh = useCallback(async () => {
    assertAccountAuthority();
    await fetchProviderData();
  }, [assertAccountAuthority, fetchProviderData]);

  const startRegistration = useCallback(() => {
    assertAccountAuthority();
    registrationGeneration.current += 1;
    registrationDomain.current = null;
    challenges.current.clear();
    challengeIssuanceInFlight.current.clear();
    verificationsInFlight.current.clear();
    setState((prev) => ({
      ...prev,
      registration: {
        step: "identity_check",
        data: {},
        identityVerified: false,
        domainVerified: false,
        domainChallenge: null,
        error: null,
      },
    }));
  }, [assertAccountAuthority]);

  const updateRegistrationData = useCallback(
    (data: Partial<ProviderRegistration>) => {
      assertAccountAuthority();
      if (data.primaryDomain !== undefined) {
        const nextDomain = normalizeProviderDomain(data.primaryDomain);
        if (registrationDomain.current !== nextDomain) {
          registrationDomain.current = nextDomain;
          registrationGeneration.current += 1;
          challenges.current.clear();
          challengeIssuanceInFlight.current.clear();
          verificationsInFlight.current.clear();
        }
      }
      setState((prev) => ({
        ...prev,
        registration: prev.registration
          ? {
              ...prev.registration,
              data: { ...prev.registration.data, ...data },
            }
          : null,
      }));
    },
    [assertAccountAuthority],
  );

  const startDomainVerification = useCallback(
    async (
      domain: string,
      method: "dns_txt" | "http_file",
    ): Promise<ProviderDomainChallenge> => {
      if (
        !accountAddress ||
        renderVerifierGeneration !== verifierGeneration.current
      ) {
        throw new ProviderDomainVerificationError("feature_unavailable");
      }
      const binding = { chainId, accountAddress };
      const verifier = requireProviderDomainVerifier(domainVerifier, binding);
      if (method !== "dns_txt" && method !== "http_file") {
        throw new ProviderDomainVerificationError("invalid_challenge");
      }
      const normalizedDomain = normalizeProviderDomain(domain);
      const primaryDomain = state.registration?.data.primaryDomain;
      if (
        !primaryDomain ||
        normalizeProviderDomain(primaryDomain) !== normalizedDomain
      ) {
        throw new ProviderDomainVerificationError("invalid_domain");
      }
      if (challengeIssuanceInFlight.current.has(normalizedDomain)) {
        throw new ProviderDomainVerificationError("challenge_in_progress");
      }
      const generation = renderVerifierGeneration;
      const currentRegistrationGeneration = registrationGeneration.current;
      const operationId = ++operationSequence.current;
      challengeIssuanceInFlight.current.set(normalizedDomain, operationId);
      let challenge: ProviderDomainChallenge;
      try {
        challenge = validateProviderDomainChallenge(
          await verifier.issueChallenge(normalizedDomain, method),
          binding,
          normalizedDomain,
          method,
        );
        if (
          generation !== verifierGeneration.current ||
          currentRegistrationGeneration !== registrationGeneration.current
        ) {
          throw new ProviderDomainVerificationError("authority_changed");
        }
      } finally {
        if (
          generation === verifierGeneration.current &&
          challengeIssuanceInFlight.current.get(normalizedDomain) ===
            operationId
        ) {
          challengeIssuanceInFlight.current.delete(normalizedDomain);
        }
      }
      challenges.current.set(normalizedDomain, {
        challenge,
        registrationGeneration: currentRegistrationGeneration,
      });

      setState((prev) => ({
        ...prev,
        registration: prev.registration
          ? {
              ...prev.registration,
              domainChallenge: challenge,
            }
          : null,
      }));

      return challenge;
    },
    [
      accountAddress,
      chainId,
      domainVerifier,
      renderVerifierGeneration,
      state.registration,
    ],
  );

  const checkDomainVerification = useCallback(
    async (domain: string): Promise<ProviderDomainVerificationEvidence> => {
      if (
        !accountAddress ||
        renderVerifierGeneration !== verifierGeneration.current
      ) {
        throw new ProviderDomainVerificationError("feature_unavailable");
      }
      const normalizedDomain = normalizeProviderDomain(domain);
      const storedChallenge = challenges.current.get(normalizedDomain);
      const challenge = storedChallenge?.challenge;
      if (
        !challenge ||
        challenge.expiresAt <= Date.now() ||
        storedChallenge.registrationGeneration !==
          registrationGeneration.current ||
        !state.registration?.data.primaryDomain ||
        normalizeProviderDomain(state.registration.data.primaryDomain) !==
          normalizedDomain
      ) {
        throw new ProviderDomainVerificationError("invalid_challenge");
      }
      if (verificationsInFlight.current.has(normalizedDomain)) {
        throw new ProviderDomainVerificationError("verification_in_progress");
      }
      const binding = { chainId, accountAddress };
      const verifier = requireProviderDomainVerifier(domainVerifier, binding);
      const generation = renderVerifierGeneration;
      const operationId = ++operationSequence.current;
      verificationsInFlight.current.set(normalizedDomain, operationId);
      let verification;
      try {
        verification = validateProviderDomainVerification(
          await verifier.verifyChallenge(challenge),
          binding,
          challenge,
        );
        if (generation !== verifierGeneration.current) {
          throw new ProviderDomainVerificationError("authority_changed");
        }
        if (
          challenge.expiresAt <= Date.now() ||
          storedChallenge.registrationGeneration !==
            registrationGeneration.current
        ) {
          throw new ProviderDomainVerificationError("invalid_challenge");
        }
        challenges.current.delete(normalizedDomain);
      } finally {
        if (
          generation === verifierGeneration.current &&
          verificationsInFlight.current.get(normalizedDomain) === operationId
        ) {
          verificationsInFlight.current.delete(normalizedDomain);
        }
      }

      setState((prev) => ({
        ...prev,
        domainVerifications: [
          ...prev.domainVerifications.filter(
            (item) => normalizeProviderDomain(item.domain) !== normalizedDomain,
          ),
          verification,
        ],
        registration: prev.registration
          ? {
              ...prev.registration,
              domainVerified: true,
              step: "stake_deposit",
            }
          : null,
      }));

      return verification;
    },
    [
      accountAddress,
      chainId,
      domainVerifier,
      renderVerifierGeneration,
      state.registration,
    ],
  );

  const submitRegistration = useCallback(async () => {
    assertAccountAuthority();
    // Would submit registration transaction
    setState((prev) => ({
      ...prev,
      registration: null,
      isRegistered: true,
    }));
    await refresh();
  }, [assertAccountAuthority, refresh]);

  const createOffering = useCallback(
    async (draft: OfferingDraft): Promise<ProviderOffering> => {
      assertAccountAuthority();
      // Would create offering transaction
      const offering: ProviderOffering = {
        id: `offering-${Date.now()}`,
        title: draft.title,
        type: draft.type,
        status: draft.autoPublish ? "active" : "draft",
        activeOrders: 0,
        totalOrders: 0,
        capacityUtilization: 0,
        totalRevenue: "0",
        createdAt: Date.now(),
        updatedAt: Date.now(),
      };

      setState((prev) => ({
        ...prev,
        offerings: [...prev.offerings, offering],
      }));

      return offering;
    },
    [assertAccountAuthority],
  );

  const updateOffering = useCallback(
    async (
      offeringId: string,
      updates: Partial<OfferingDraft>,
    ): Promise<ProviderOffering> => {
      assertAccountAuthority();
      const offering = state.offerings.find((o) => o.id === offeringId);
      if (!offering) throw new Error("Offering not found");

      const updated = { ...offering, ...updates, updatedAt: Date.now() };
      setState((prev) => ({
        ...prev,
        offerings: prev.offerings.map((o) =>
          o.id === offeringId ? updated : o,
        ),
      }));

      return updated;
    },
    [assertAccountAuthority, state.offerings],
  );

  const publishOffering = useCallback(
    async (offeringId: string) => {
      assertAccountAuthority();
      setState((prev) => ({
        ...prev,
        offerings: prev.offerings.map((o) =>
          o.id === offeringId
            ? { ...o, status: "active", updatedAt: Date.now() }
            : o,
        ),
      }));
    },
    [assertAccountAuthority],
  );

  const pauseOffering = useCallback(
    async (offeringId: string) => {
      assertAccountAuthority();
      setState((prev) => ({
        ...prev,
        offerings: prev.offerings.map((o) =>
          o.id === offeringId
            ? { ...o, status: "paused", updatedAt: Date.now() }
            : o,
        ),
      }));
    },
    [assertAccountAuthority],
  );

  const getIncomingOrders = useCallback(async () => {
    assertAccountAuthority();
    // Would fetch incoming orders
    setState((prev) => ({ ...prev, incomingOrders: [] }));
  }, [assertAccountAuthority]);

  const getActiveBids = useCallback(async () => {
    assertAccountAuthority();
    setState((prev) => ({ ...prev, activeBids: [] }));
  }, [assertAccountAuthority]);

  const getActiveAllocations = useCallback(async () => {
    assertAccountAuthority();
    setState((prev) => ({ ...prev, activeAllocations: [] }));
  }, [assertAccountAuthority]);

  const getUsageRecords = useCallback(
    async (allocationId?: string) => {
      assertAccountAuthority();
      setState((prev) => ({ ...prev, usageRecords: [] }));
    },
    [assertAccountAuthority],
  );

  const getSettlementSummary = useCallback(async () => {
    assertAccountAuthority();
    const summary: SettlementSummary = {
      periodStart: Date.now() - 30 * 24 * 60 * 60 * 1000,
      periodEnd: Date.now(),
      totalOrders: 0,
      totalRevenue: "0",
      totalSettled: "0",
      pendingSettlement: "0",
      byOffering: [],
      recentSettlements: [],
    };
    setState((prev) => ({ ...prev, settlementSummary: summary }));
  }, [assertAccountAuthority]);

  const clearError = useCallback(() => {
    assertAccountAuthority();
    setState((prev) => ({ ...prev, error: null }));
  }, [assertAccountAuthority]);

  useEffect(() => {
    fetchProviderData();
  }, [fetchProviderData]);

  const actions: ProviderActions = {
    refresh,
    startRegistration,
    updateRegistrationData,
    startDomainVerification,
    checkDomainVerification,
    submitRegistration,
    createOffering,
    updateOffering,
    publishOffering,
    pauseOffering,
    getIncomingOrders,
    getActiveBids,
    getActiveAllocations,
    getUsageRecords,
    getSettlementSummary,
    clearError,
  };

  return (
    <ProviderContext.Provider value={{ state: effectiveState, actions }}>
      {children}
    </ProviderContext.Provider>
  );
}

export function useProvider(): ProviderContextValue {
  const context = useContext(ProviderContext);
  if (!context) {
    throw new Error("useProvider must be used within a ProviderProvider");
  }
  return context;
}
