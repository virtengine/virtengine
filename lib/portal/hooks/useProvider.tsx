/**
 * useProvider Hook
 * VE-704: Provider console (offerings, pricing, capacity, domain verification)
 */

import { useState, useCallback, useEffect, useContext, createContext } from 'react';
import type { ReactNode } from 'react';
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
} from '../types/provider';
import { initialProviderState } from '../types/provider';
import type { QueryClient } from '../types/chain';

/**
 * Provider context value
 */
interface ProviderContextValue {
  state: ProviderState;
  actions: ProviderActions;
}

/**
 * Provider actions
 */
interface ProviderActions {
  refresh: () => Promise<void>;
  startRegistration: () => void;
  updateRegistrationData: (data: Partial<ProviderRegistration>) => void;
  startDomainVerification: (domain: string, method: 'dns_txt' | 'http_file') => Promise<DomainChallenge>;
  checkDomainVerification: (domain: string) => Promise<DomainVerification>;
  submitRegistration: () => Promise<void>;
  createOffering: (draft: OfferingDraft) => Promise<ProviderOffering>;
  updateOffering: (offeringId: string, updates: Partial<OfferingDraft>) => Promise<ProviderOffering>;
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
  accountAddress: string | null;
  getAuthHeader: () => Promise<string>;
}

export function ProviderProvider({
  children,
  queryClient,
  accountAddress,
  getAuthHeader,
}: ProviderProviderProps) {
  const [state, setState] = useState<ProviderState>(initialProviderState);

  const fetchProviderData = useCallback(async () => {
    if (!accountAddress) {
      setState(initialProviderState);
      return;
    }

    setState(prev => ({ ...prev, isLoading: true }));

    try {
      const providerInfo = await queryClient.queryProvider(accountAddress);
      
      if (providerInfo) {
        const profile: ProviderProfile = {
          address: providerInfo.address,
          name: '',
          description: '',
          website: '',
          verifiedDomains: [],
          status: providerInfo.status as any,
          identityScore: 0,
          reliabilityScore: providerInfo.reliabilityScore,
          registeredAt: providerInfo.registeredAt,
          offeringsCount: 0,
          ordersFulfilled: 0,
          tier: 'bronze',
          stakedAmount: '0',
        };

        setState(prev => ({
          ...prev,
          isLoading: false,
          isRegistered: true,
          profile,
          error: null,
        }));
      } else {
        setState(prev => ({
          ...prev,
          isLoading: false,
          isRegistered: false,
          error: null,
        }));
      }
    } catch (error) {
      setState(prev => ({
        ...prev,
        isLoading: false,
        isRegistered: false,
        error: null,
      }));
    }
  }, [accountAddress, queryClient]);

  const refresh = useCallback(async () => {
    await fetchProviderData();
  }, [fetchProviderData]);

  const startRegistration = useCallback(() => {
    setState(prev => ({
      ...prev,
      registration: {
        step: 'identity_check',
        data: {},
        identityVerified: false,
        domainVerified: false,
        domainChallenge: null,
        error: null,
      },
    }));
  }, []);

  const updateRegistrationData = useCallback((data: Partial<ProviderRegistration>) => {
    setState(prev => ({
      ...prev,
      registration: prev.registration ? {
        ...prev.registration,
        data: { ...prev.registration.data, ...data },
      } : null,
    }));
  }, []);

  const startDomainVerification = useCallback(async (
    domain: string,
    method: 'dns_txt' | 'http_file'
  ): Promise<DomainChallenge> => {
    const challenge: DomainChallenge = {
      domain,
      method,
      challengeValue: `ve-verification=${Math.random().toString(36).slice(2)}`,
      dnsRecordName: method === 'dns_txt' ? `_virtengine.${domain}` : undefined,
      httpFilePath: method === 'http_file' ? '/.well-known/virtengine-verification' : undefined,
      expiresAt: Date.now() + 24 * 60 * 60 * 1000,
      instructions: method === 'dns_txt'
        ? `Add a TXT record for _virtengine.${domain} with the challenge value`
        : `Create a file at /.well-known/virtengine-verification with the challenge value`,
    };

    setState(prev => ({
      ...prev,
      registration: prev.registration ? {
        ...prev.registration,
        domainChallenge: challenge,
      } : null,
    }));

    return challenge;
  }, []);

  const checkDomainVerification = useCallback(async (domain: string): Promise<DomainVerification> => {
    // Would check if domain is verified
    const verification: DomainVerification = {
      domain,
      status: 'verified',
      method: 'dns_txt',
      verifiedAt: Date.now(),
      expiresAt: Date.now() + 365 * 24 * 60 * 60 * 1000,
    };

    setState(prev => ({
      ...prev,
      domainVerifications: [...prev.domainVerifications.filter(d => d.domain !== domain), verification],
      registration: prev.registration ? {
        ...prev.registration,
        domainVerified: true,
        step: 'stake_deposit',
      } : null,
    }));

    return verification;
  }, []);

  const submitRegistration = useCallback(async () => {
    // Would submit registration transaction
    setState(prev => ({
      ...prev,
      registration: null,
      isRegistered: true,
    }));
    await refresh();
  }, [refresh]);

  const createOffering = useCallback(async (draft: OfferingDraft): Promise<ProviderOffering> => {
    // Would create offering transaction
    const offering: ProviderOffering = {
      id: `offering-${Date.now()}`,
      title: draft.title,
      type: draft.type,
      status: draft.autoPublish ? 'active' : 'draft',
      activeOrders: 0,
      totalOrders: 0,
      capacityUtilization: 0,
      totalRevenue: '0',
      createdAt: Date.now(),
      updatedAt: Date.now(),
    };

    setState(prev => ({
      ...prev,
      offerings: [...prev.offerings, offering],
    }));

    return offering;
  }, []);

  const updateOffering = useCallback(async (
    offeringId: string,
    updates: Partial<OfferingDraft>
  ): Promise<ProviderOffering> => {
    const offering = state.offerings.find(o => o.id === offeringId);
    if (!offering) throw new Error('Offering not found');

    const updated = { ...offering, ...updates, updatedAt: Date.now() };
    setState(prev => ({
      ...prev,
      offerings: prev.offerings.map(o => o.id === offeringId ? updated : o),
    }));

    return updated;
  }, [state.offerings]);

  const publishOffering = useCallback(async (offeringId: string) => {
    setState(prev => ({
      ...prev,
      offerings: prev.offerings.map(o => 
        o.id === offeringId ? { ...o, status: 'active', updatedAt: Date.now() } : o
      ),
    }));
  }, []);

  const pauseOffering = useCallback(async (offeringId: string) => {
    setState(prev => ({
      ...prev,
      offerings: prev.offerings.map(o => 
        o.id === offeringId ? { ...o, status: 'paused', updatedAt: Date.now() } : o
      ),
    }));
  }, []);

  const getIncomingOrders = useCallback(async () => {
    // Would fetch incoming orders
    setState(prev => ({ ...prev, incomingOrders: [] }));
  }, []);

  const getActiveBids = useCallback(async () => {
    setState(prev => ({ ...prev, activeBids: [] }));
  }, []);

  const getActiveAllocations = useCallback(async () => {
    setState(prev => ({ ...prev, activeAllocations: [] }));
  }, []);

  const getUsageRecords = useCallback(async (allocationId?: string) => {
    setState(prev => ({ ...prev, usageRecords: [] }));
  }, []);

  const getSettlementSummary = useCallback(async () => {
    const summary: SettlementSummary = {
      periodStart: Date.now() - 30 * 24 * 60 * 60 * 1000,
      periodEnd: Date.now(),
      totalOrders: 0,
      totalRevenue: '0',
      totalSettled: '0',
      pendingSettlement: '0',
      byOffering: [],
      recentSettlements: [],
    };
    setState(prev => ({ ...prev, settlementSummary: summary }));
  }, []);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

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
    <ProviderContext.Provider value={{ state, actions }}>
      {children}
    </ProviderContext.Provider>
  );
}

export function useProvider(): ProviderContextValue {
  const context = useContext(ProviderContext);
  if (!context) {
    throw new Error('useProvider must be used within a ProviderProvider');
  }
  return context;
}
