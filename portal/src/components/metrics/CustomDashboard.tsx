/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Custom dashboard view with configurable widget layout.
 * Supports creating, editing, and switching between dashboard configurations.
 */

'use client';

import { useState } from 'react';
import { Plus, Pencil, Trash2, LayoutGrid } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import {
  DEFAULT_DASHBOARD,
  useDashboardConfigStore,
  selectActiveDashboard,
  selectDashboardNames,
} from '@/stores/dashboardConfigStore';
import { DashboardWidget } from './DashboardWidget';
import type { WidgetType } from '@virtengine/portal/types/metrics';

const WIDGET_PRESETS: Array<{ type: WidgetType; title: string; metric?: string }> = [
  { type: 'metric-card', title: 'CPU Usage', metric: 'cpu' },
  { type: 'metric-card', title: 'Memory Usage', metric: 'memory' },
  { type: 'metric-card', title: 'Storage Usage', metric: 'storage' },
  { type: 'time-series-chart', title: 'CPU Over Time', metric: 'cpu' },
  { type: 'time-series-chart', title: 'Memory Over Time', metric: 'memory' },
  { type: 'alert-list', title: 'Active Alerts' },
  { type: 'table', title: 'Provider Breakdown' },
];

export function CustomDashboard() {
  const storedActiveDashboard = useDashboardConfigStore(selectActiveDashboard);
  const storedDashboardNames = useDashboardConfigStore(selectDashboardNames);
  const isEditing = useDashboardConfigStore((s) => s.isEditing);
  const mutationPending = useDashboardConfigStore((s) => s.dashboardMutationPending);
  const mutationsAvailable = useDashboardConfigStore((s) => s.dashboardMutationsAvailable);
  const mutationError = useDashboardConfigStore((s) => s.error);
  const setActiveDashboard = useDashboardConfigStore((s) => s.setActiveDashboard);
  const toggleEditing = useDashboardConfigStore((s) => s.toggleEditing);
  const createDashboard = useDashboardConfigStore((s) => s.createDashboard);
  const deleteDashboard = useDashboardConfigStore((s) => s.deleteDashboard);
  const addWidget = useDashboardConfigStore((s) => s.addWidget);
  const removeWidget = useDashboardConfigStore((s) => s.removeWidget);

  const [newDashboardName, setNewDashboardName] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showAddWidget, setShowAddWidget] = useState(false);

  const activeDashboard = mutationsAvailable ? storedActiveDashboard : DEFAULT_DASHBOARD;
  const dashboardNames = mutationsAvailable
    ? storedDashboardNames
    : [{ id: DEFAULT_DASHBOARD.id, name: DEFAULT_DASHBOARD.name, isDefault: true }];
  const canMutateActive = Boolean(
    mutationsAvailable && activeDashboard && !activeDashboard.isDefault
  );

  async function handleCreateDashboard() {
    const name = newDashboardName.trim();
    if (!name || !mutationsAvailable || mutationPending) return;
    try {
      await createDashboard(name);
      setNewDashboardName('');
      setShowCreate(false);
    } catch {
      // The store exposes the user-facing commit error.
    }
  }

  async function handleAddWidget(preset: (typeof WIDGET_PRESETS)[number]) {
    if (!activeDashboard || !canMutateActive || mutationPending) return;
    try {
      await addWidget(activeDashboard.id, preset.type, preset.title, {
        metric: preset.metric,
        timeRange: '24h',
      });
      setShowAddWidget(false);
    } catch {
      // The store exposes the user-facing commit error.
    }
  }

  async function handleDeleteDashboard() {
    if (!activeDashboard || !canMutateActive || mutationPending) return;
    try {
      await deleteDashboard(activeDashboard.id);
    } catch {
      // The store exposes the user-facing commit error.
    }
  }

  async function handleRemoveWidget(widgetId: string) {
    if (!activeDashboard || !canMutateActive || mutationPending) return;
    try {
      await removeWidget(activeDashboard.id, widgetId);
    } catch {
      // The store exposes the user-facing commit error.
    }
  }

  if (!activeDashboard) {
    return (
      <div className="py-12 text-center text-muted-foreground">
        <LayoutGrid className="mx-auto mb-4 h-12 w-12 opacity-50" />
        <p>No dashboards configured.</p>
        <Button className="mt-4" onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Dashboard
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {!mutationsAvailable && (
        <p role="alert" className="rounded-md border border-warning p-3 text-sm text-warning">
          Dashboard persistence is unavailable. Showing the built-in Overview dashboard.
        </p>
      )}
      {mutationPending && (
        <p role="status" aria-live="polite" className="text-sm text-muted-foreground">
          Saving dashboard change...
        </p>
      )}
      {mutationError && (
        <p role="alert" className="text-sm text-destructive">
          {mutationError}
        </p>
      )}

      {/* Dashboard tabs */}
      <div className="flex items-center justify-between">
        <Tabs
          value={activeDashboard.id}
          onValueChange={(id) => !mutationPending && setActiveDashboard(id)}
        >
          <TabsList>
            {dashboardNames.map((d) => (
              <TabsTrigger key={d.id} value={d.id} disabled={mutationPending}>
                {d.name}
                {d.isDefault && (
                  <Badge variant="secondary" className="ml-2" size="sm">
                    default
                  </Badge>
                )}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>

        <div className="flex items-center gap-2">
          {showCreate ? (
            <div className="flex items-center gap-2">
              <Input
                aria-label="Dashboard name"
                value={newDashboardName}
                onChange={(e) => setNewDashboardName(e.target.value)}
                placeholder="Dashboard name"
                className="h-9 w-40"
                disabled={mutationPending}
                onKeyDown={(e) => e.key === 'Enter' && void handleCreateDashboard()}
              />
              <Button
                size="sm"
                onClick={() => void handleCreateDashboard()}
                disabled={!newDashboardName.trim() || !mutationsAvailable || mutationPending}
              >
                Create
              </Button>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setShowCreate(false)}
                disabled={mutationPending}
              >
                Cancel
              </Button>
            </div>
          ) : (
            <>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setShowCreate(true)}
                disabled={!mutationsAvailable || mutationPending}
              >
                <Plus className="mr-1 h-3 w-3" />
                New
              </Button>
              <Button
                size="sm"
                variant={isEditing ? 'default' : 'outline'}
                onClick={toggleEditing}
                disabled={!canMutateActive || mutationPending}
              >
                <Pencil className="mr-1 h-3 w-3" />
                {isEditing ? 'Done' : 'Edit'}
              </Button>
              {!activeDashboard.isDefault && (
                <Button
                  size="sm"
                  variant="destructive"
                  aria-label={`Delete dashboard ${activeDashboard.name}`}
                  title={`Delete dashboard ${activeDashboard.name}`}
                  onClick={() => void handleDeleteDashboard()}
                  disabled={!canMutateActive || mutationPending}
                >
                  <Trash2 className="h-3 w-3" />
                </Button>
              )}
            </>
          )}
        </div>
      </div>

      {/* Add widget panel */}
      {isEditing && canMutateActive && (
        <div className="rounded-lg border border-dashed p-3">
          {showAddWidget ? (
            <div className="space-y-2">
              <p className="text-sm font-medium">Add Widget</p>
              <div className="flex flex-wrap gap-2">
                {WIDGET_PRESETS.map((preset) => (
                  <Button
                    key={preset.title}
                    size="sm"
                    variant="outline"
                    onClick={() => void handleAddWidget(preset)}
                    disabled={mutationPending}
                  >
                    {preset.title}
                  </Button>
                ))}
              </div>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setShowAddWidget(false)}
                disabled={mutationPending}
              >
                Cancel
              </Button>
            </div>
          ) : (
            <Button
              size="sm"
              variant="ghost"
              className="w-full"
              onClick={() => setShowAddWidget(true)}
              disabled={mutationPending}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Widget
            </Button>
          )}
        </div>
      )}

      {/* Widget grid */}
      {activeDashboard.layout.length === 0 ? (
        <div className="py-12 text-center text-muted-foreground">
          <p>This dashboard has no widgets.</p>
          {!isEditing && canMutateActive && (
            <Button className="mt-4" variant="outline" onClick={toggleEditing}>
              <Plus className="mr-2 h-4 w-4" />
              Add Widgets
            </Button>
          )}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {activeDashboard.layout.map((widget) => (
            <div
              key={widget.id}
              className={
                widget.position.w >= 9
                  ? 'md:col-span-2 lg:col-span-3'
                  : widget.position.w >= 6
                    ? 'md:col-span-2'
                    : ''
              }
            >
              <DashboardWidget
                widget={widget}
                isEditing={isEditing && canMutateActive}
                onRemove={() => void handleRemoveWidget(widget.id)}
                removeDisabled={!canMutateActive || mutationPending}
                removePending={mutationPending}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
