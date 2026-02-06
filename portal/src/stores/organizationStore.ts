/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Organization store — Zustand store for current organization context.
 * Persists selected organization across sessions.
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface OrganizationStoreState {
  currentOrgId: string | null;
}

export interface OrganizationStoreActions {
  setCurrentOrg: (orgId: string | null) => void;
  clearOrg: () => void;
}

export type OrganizationStore = OrganizationStoreState & OrganizationStoreActions;

export const useOrganizationStore = create<OrganizationStore>()(
  persist(
    (set) => ({
      currentOrgId: null,

      setCurrentOrg: (orgId: string | null) => {
        set({ currentOrgId: orgId });
      },

      clearOrg: () => {
        set({ currentOrgId: null });
      },
    }),
    {
      name: 've-organization-storage',
    }
  )
);
