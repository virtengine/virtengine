/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Organization store — Zustand store for authoritative organization context.
 */

import { create } from 'zustand';

const LEGACY_STORAGE_KEY = 've-organization-storage';

export interface OrganizationSelectionAuthority {
  organizationsAvailable: true;
}

export interface OrganizationStoreState {
  currentOrgId: string | null;
}

export interface OrganizationStoreActions {
  setCurrentOrg: (orgId: string | null, authority?: OrganizationSelectionAuthority) => boolean;
  clearOrg: () => void;
}

export type OrganizationStore = OrganizationStoreState & OrganizationStoreActions;

export function clearLegacyOrganizationState(): void {
  if (typeof window === 'undefined') return;

  try {
    window.localStorage.removeItem(LEGACY_STORAGE_KEY);
  } catch {
    // Storage may be unavailable; it is never read as organization authority.
  }
}

clearLegacyOrganizationState();

export const useOrganizationStore = create<OrganizationStore>()((set) => ({
  currentOrgId: null,

  setCurrentOrg: (orgId, authority) => {
    if (orgId === null) {
      set({ currentOrgId: null });
      return true;
    }

    if (authority?.organizationsAvailable !== true) {
      set({ currentOrgId: null });
      return false;
    }

    set({ currentOrgId: orgId });
    return true;
  },

  clearOrg: () => {
    set({ currentOrgId: null });
  },
}));
