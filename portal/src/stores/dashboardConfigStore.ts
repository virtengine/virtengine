/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Dashboard configuration Zustand store.
 * Manages custom dashboard layouts, widget placement, and persistence.
 */

import { create } from 'zustand';
import { generateId } from '@/lib/utils';
import type {
  DashboardConfig,
  DashboardWidget,
  WidgetType,
  WidgetConfig,
  WidgetPosition,
} from '@virtengine/portal/types/metrics';

// =============================================================================
// Store Interface
// =============================================================================

export interface DashboardConfigState {
  dashboards: DashboardConfig[];
  activeDashboardId: string | null;
  isEditing: boolean;
}

export interface DashboardConfigActions {
  createDashboard: (name: string) => string;
  deleteDashboard: (id: string) => void;
  setActiveDashboard: (id: string) => void;
  toggleEditing: () => void;
  addWidget: (dashboardId: string, type: WidgetType, title: string, config: WidgetConfig) => void;
  removeWidget: (dashboardId: string, widgetId: string) => void;
  updateWidgetPosition: (dashboardId: string, widgetId: string, position: WidgetPosition) => void;
  updateWidgetConfig: (dashboardId: string, widgetId: string, config: WidgetConfig) => void;
  renameDashboard: (id: string, name: string) => void;
}

export type DashboardConfigStore = DashboardConfigState & DashboardConfigActions;

// =============================================================================
// Default Dashboard
// =============================================================================

const DEFAULT_DASHBOARD: DashboardConfig = {
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
  createdAt: Date.now(),
  updatedAt: Date.now(),
};

// =============================================================================
// Store Implementation
// =============================================================================

export const useDashboardConfigStore = create<DashboardConfigStore>()((set, get) => ({
  dashboards: [DEFAULT_DASHBOARD],
  activeDashboardId: DEFAULT_DASHBOARD.id,
  isEditing: false,

  createDashboard: (name) => {
    const id = generateId('dashboard');
    const newDashboard: DashboardConfig = {
      id,
      name,
      isDefault: false,
      layout: [],
      createdAt: Date.now(),
      updatedAt: Date.now(),
    };
    set((state) => ({
      dashboards: [...state.dashboards, newDashboard],
      activeDashboardId: id,
    }));
    return id;
  },

  deleteDashboard: (id) => {
    const { dashboards, activeDashboardId } = get();
    const target = dashboards.find((d) => d.id === id);
    if (target?.isDefault) return;
    const filtered = dashboards.filter((d) => d.id !== id);
    set({
      dashboards: filtered,
      activeDashboardId: activeDashboardId === id ? (filtered[0]?.id ?? null) : activeDashboardId,
    });
  },

  setActiveDashboard: (id) => {
    set({ activeDashboardId: id });
  },

  toggleEditing: () => {
    set((state) => ({ isEditing: !state.isEditing }));
  },

  addWidget: (dashboardId, type, title, config) => {
    const widgetId = generateId('widget');
    const widget: DashboardWidget = {
      id: widgetId,
      type,
      title,
      config,
      position: { x: 0, y: 0, w: 6, h: 3 },
    };
    set((state) => ({
      dashboards: state.dashboards.map((d) =>
        d.id === dashboardId ? { ...d, layout: [...d.layout, widget], updatedAt: Date.now() } : d
      ),
    }));
  },

  removeWidget: (dashboardId, widgetId) => {
    set((state) => ({
      dashboards: state.dashboards.map((d) =>
        d.id === dashboardId
          ? {
              ...d,
              layout: d.layout.filter((w) => w.id !== widgetId),
              updatedAt: Date.now(),
            }
          : d
      ),
    }));
  },

  updateWidgetPosition: (dashboardId, widgetId, position) => {
    set((state) => ({
      dashboards: state.dashboards.map((d) =>
        d.id === dashboardId
          ? {
              ...d,
              layout: d.layout.map((w) => (w.id === widgetId ? { ...w, position } : w)),
              updatedAt: Date.now(),
            }
          : d
      ),
    }));
  },

  updateWidgetConfig: (dashboardId, widgetId, config) => {
    set((state) => ({
      dashboards: state.dashboards.map((d) =>
        d.id === dashboardId
          ? {
              ...d,
              layout: d.layout.map((w) => (w.id === widgetId ? { ...w, config } : w)),
              updatedAt: Date.now(),
            }
          : d
      ),
    }));
  },

  renameDashboard: (id, name) => {
    set((state) => ({
      dashboards: state.dashboards.map((d) =>
        d.id === id ? { ...d, name, updatedAt: Date.now() } : d
      ),
    }));
  },
}));

// =============================================================================
// Selectors
// =============================================================================

export const selectActiveDashboard = (state: DashboardConfigStore): DashboardConfig | undefined =>
  state.activeDashboardId
    ? state.dashboards.find((d) => d.id === state.activeDashboardId)
    : state.dashboards[0];

export const selectDashboardNames = (
  state: DashboardConfigStore
): Array<{ id: string; name: string; isDefault: boolean }> =>
  state.dashboards.map((d) => ({ id: d.id, name: d.name, isDefault: d.isDefault }));
