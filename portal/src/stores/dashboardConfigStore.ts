/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Dashboard configuration Zustand store.
 * Manages custom dashboard layouts, widget placement, and persistence.
 */

import { create } from 'zustand';
import type {
  DashboardConfig,
  WidgetType,
  WidgetConfig,
  WidgetPosition,
} from '@virtengine/portal/types/metrics';
import {
  DashboardConfigMutationError,
  submitDashboardConfigMutation,
  type DashboardConfigMutationAdapter,
  type DashboardConfigMutationCommand,
} from './dashboard-config-mutation';

// =============================================================================
// Store Interface
// =============================================================================

export interface DashboardConfigState {
  dashboards: readonly DashboardConfig[];
  activeDashboardId: string | null;
  isEditing: boolean;
  dashboardMutationPending: boolean;
  dashboardMutationsAvailable: boolean;
  error: string | null;
}

export interface DashboardConfigActions {
  createDashboard: (name: string) => Promise<void>;
  deleteDashboard: (id: string) => Promise<void>;
  setActiveDashboard: (id: string) => void;
  toggleEditing: () => void;
  addWidget: (
    dashboardId: string,
    type: WidgetType,
    title: string,
    config: WidgetConfig
  ) => Promise<void>;
  removeWidget: (dashboardId: string, widgetId: string) => Promise<void>;
  updateWidgetPosition: (
    dashboardId: string,
    widgetId: string,
    position: WidgetPosition
  ) => Promise<void>;
  updateWidgetConfig: (
    dashboardId: string,
    widgetId: string,
    config: WidgetConfig
  ) => Promise<void>;
  renameDashboard: (id: string, name: string) => Promise<void>;
}

export type DashboardConfigStore = DashboardConfigState & DashboardConfigActions;

// =============================================================================
// Default Dashboard
// =============================================================================

const deepFreeze = <T>(value: T): T => {
  if (value && typeof value === 'object' && !Object.isFrozen(value)) {
    Object.freeze(value);
    for (const child of Object.values(value as Record<string, unknown>)) deepFreeze(child);
  }
  return value;
};

export const DEFAULT_DASHBOARD: Readonly<DashboardConfig> = deepFreeze({
  id: 'dashboard-default',
  name: 'Overview',
  isDefault: true,
  layout: [
    {
      id: 'w-cpu',
      type: 'metric-card',
      title: 'CPU Usage',
      config: { metric: 'cpu' },
      position: { x: 0, y: 0, w: 3, h: 2 },
    },
    {
      id: 'w-mem',
      type: 'metric-card',
      title: 'Memory Usage',
      config: { metric: 'memory' },
      position: { x: 3, y: 0, w: 3, h: 2 },
    },
    {
      id: 'w-stor',
      type: 'metric-card',
      title: 'Storage Usage',
      config: { metric: 'storage' },
      position: { x: 6, y: 0, w: 3, h: 2 },
    },
    {
      id: 'w-deploy',
      type: 'metric-card',
      title: 'Deployments',
      config: { metric: 'deployments' },
      position: { x: 9, y: 0, w: 3, h: 2 },
    },
    {
      id: 'w-cpu-chart',
      type: 'time-series-chart',
      title: 'CPU Over Time',
      config: { metric: 'cpu', timeRange: '24h' },
      position: { x: 0, y: 2, w: 6, h: 4 },
    },
    {
      id: 'w-mem-chart',
      type: 'time-series-chart',
      title: 'Memory Over Time',
      config: { metric: 'memory', timeRange: '24h' },
      position: { x: 6, y: 2, w: 6, h: 4 },
    },
    {
      id: 'w-alerts',
      type: 'alert-list',
      title: 'Active Alerts',
      config: {},
      position: { x: 0, y: 6, w: 6, h: 3 },
    },
    {
      id: 'w-providers',
      type: 'table',
      title: 'Provider Breakdown',
      config: {},
      position: { x: 6, y: 6, w: 6, h: 3 },
    },
  ],
  createdAt: 0,
  updatedAt: 0,
});

const DEFAULT_DASHBOARDS: readonly DashboardConfig[] = Object.freeze([DEFAULT_DASHBOARD]);

let mutationAdapter: DashboardConfigMutationAdapter | null = null;
let mutationSubject = '';
let mutationGeneration = 0;
let acceptedRevision = 0;
let mutationPending = false;
let activeMutation: AbortController | null = null;

export const configureDashboardConfigMutations = (
  adapter: DashboardConfigMutationAdapter | null,
  subject = ''
) => {
  activeMutation?.abort();
  activeMutation = null;
  mutationPending = false;
  mutationGeneration += 1;
  acceptedRevision = 0;
  mutationAdapter = adapter;
  mutationSubject = subject.trim();
  useDashboardConfigStore.setState({
    dashboards: DEFAULT_DASHBOARDS,
    activeDashboardId: DEFAULT_DASHBOARD.id,
    isEditing: false,
    dashboardMutationPending: false,
    dashboardMutationsAvailable: Boolean(adapter && mutationSubject),
    error: null,
  });
};

// =============================================================================
// Store Implementation
// =============================================================================

export const useDashboardConfigStore = create<DashboardConfigStore>()((set, get) => {
  const mutate = async (request: DashboardConfigMutationCommand) => {
    if (mutationPending) throw new DashboardConfigMutationError('request_changed');
    const adapter = mutationAdapter;
    const subject = mutationSubject;
    if (!adapter || !subject) {
      set({ error: 'Dashboard persistence is unavailable.' });
      throw new DashboardConfigMutationError('unavailable');
    }
    mutationPending = true;
    const controller = new AbortController();
    activeMutation = controller;
    const generation = mutationGeneration;
    set({ dashboardMutationPending: true, error: null });
    try {
      const result = await submitDashboardConfigMutation({
        adapter,
        request: { ...request, subject },
        signal: controller.signal,
        isCurrent: () =>
          generation === mutationGeneration &&
          adapter === mutationAdapter &&
          subject === mutationSubject,
      });
      if (generation !== mutationGeneration) {
        throw new DashboardConfigMutationError('request_changed');
      }
      if (result.revision <= acceptedRevision) {
        throw new DashboardConfigMutationError('invalid_committed_result');
      }
      acceptedRevision = result.revision;
      const dashboards = Object.freeze([DEFAULT_DASHBOARD, ...result.dashboards]);
      const selectedId = request.action === 'create' ? result.affectedId : get().activeDashboardId;
      set({
        dashboards,
        activeDashboardId: dashboards.some((dashboard) => dashboard.id === selectedId)
          ? selectedId
          : DEFAULT_DASHBOARD.id,
      });
    } catch (error) {
      if (generation === mutationGeneration) set({ error: 'Dashboard change was not committed.' });
      throw error;
    } finally {
      if (generation === mutationGeneration) {
        mutationPending = false;
        activeMutation = null;
        set({ dashboardMutationPending: false });
      }
    }
  };

  return {
    dashboards: DEFAULT_DASHBOARDS,
    activeDashboardId: DEFAULT_DASHBOARD.id,
    isEditing: false,
    dashboardMutationPending: false,
    dashboardMutationsAvailable: false,
    error: null,

    createDashboard: async (name) => mutate({ action: 'create', name }),

    deleteDashboard: async (dashboardId) => {
      if (dashboardId === DEFAULT_DASHBOARD.id) {
        throw new DashboardConfigMutationError('invalid_request');
      }
      await mutate({ action: 'delete', dashboardId });
    },

    setActiveDashboard: (id) => {
      if (get().dashboards.some((dashboard) => dashboard.id === id)) set({ activeDashboardId: id });
    },

    toggleEditing: () => {
      set((state) => ({ isEditing: !state.isEditing }));
    },

    addWidget: async (dashboardId, type, title, config) =>
      mutate({
        action: 'add',
        dashboardId,
        widget: { type, title, config, position: { x: 0, y: 0, w: 6, h: 3 } },
      }),

    removeWidget: async (dashboardId, widgetId) =>
      mutate({ action: 'remove', dashboardId, widgetId }),

    updateWidgetPosition: async (dashboardId, widgetId, position) =>
      mutate({ action: 'update-position', dashboardId, widgetId, position }),

    updateWidgetConfig: async (dashboardId, widgetId, config) =>
      mutate({ action: 'update-config', dashboardId, widgetId, config }),

    renameDashboard: async (dashboardId, name) => mutate({ action: 'rename', dashboardId, name }),
  };
});

// =============================================================================
// Selectors
// =============================================================================

export const selectActiveDashboard = (state: DashboardConfigStore): DashboardConfig | undefined =>
  state.activeDashboardId
    ? state.dashboards.find((d) => d.id === state.activeDashboardId)
    : state.dashboards[0];

export const selectDashboardNames = (
  state: DashboardConfigStore
): ReadonlyArray<{ id: string; name: string; isDefault: boolean }> =>
  state.dashboards.map((d) => ({ id: d.id, name: d.name, isDefault: d.isDefault }));
