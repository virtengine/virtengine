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
  const activeDashboard = useDashboardConfigStore(selectActiveDashboard);
  const dashboardNames = useDashboardConfigStore(selectDashboardNames);
  const isEditing = useDashboardConfigStore((s) => s.isEditing);
  const setActiveDashboard = useDashboardConfigStore((s) => s.setActiveDashboard);
  const toggleEditing = useDashboardConfigStore((s) => s.toggleEditing);
  const createDashboard = useDashboardConfigStore((s) => s.createDashboard);
  const deleteDashboard = useDashboardConfigStore((s) => s.deleteDashboard);
  const addWidget = useDashboardConfigStore((s) => s.addWidget);
  const removeWidget = useDashboardConfigStore((s) => s.removeWidget);

  const [newDashboardName, setNewDashboardName] = useState('');
  const [showCreate, setShowCreate] = useState(false);
  const [showAddWidget, setShowAddWidget] = useState(false);

  function handleCreateDashboard() {
    if (!newDashboardName.trim()) return;
    createDashboard(newDashboardName.trim());
    setNewDashboardName('');
    setShowCreate(false);
  }

  function handleAddWidget(preset: (typeof WIDGET_PRESETS)[number]) {
    if (!activeDashboard) return;
    addWidget(activeDashboard.id, preset.type, preset.title, {
      metric: preset.metric,
      timeRange: '24h',
    });
    setShowAddWidget(false);
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
      {/* Dashboard tabs */}
      <div className="flex items-center justify-between">
        <Tabs value={activeDashboard.id} onValueChange={setActiveDashboard}>
          <TabsList>
            {dashboardNames.map((d) => (
              <TabsTrigger key={d.id} value={d.id}>
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
                value={newDashboardName}
                onChange={(e) => setNewDashboardName(e.target.value)}
                placeholder="Dashboard name"
                className="h-9 w-40"
                onKeyDown={(e) => e.key === 'Enter' && handleCreateDashboard()}
              />
              <Button size="sm" onClick={handleCreateDashboard}>
                Create
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowCreate(false)}>
                Cancel
              </Button>
            </div>
          ) : (
            <>
              <Button size="sm" variant="outline" onClick={() => setShowCreate(true)}>
                <Plus className="mr-1 h-3 w-3" />
                New
              </Button>
              <Button size="sm" variant={isEditing ? 'default' : 'outline'} onClick={toggleEditing}>
                <Pencil className="mr-1 h-3 w-3" />
                {isEditing ? 'Done' : 'Edit'}
              </Button>
              {!activeDashboard.isDefault && (
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => deleteDashboard(activeDashboard.id)}
                >
                  <Trash2 className="h-3 w-3" />
                </Button>
              )}
            </>
          )}
        </div>
      </div>

      {/* Add widget panel */}
      {isEditing && (
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
                    onClick={() => handleAddWidget(preset)}
                  >
                    {preset.title}
                  </Button>
                ))}
              </div>
              <Button size="sm" variant="ghost" onClick={() => setShowAddWidget(false)}>
                Cancel
              </Button>
            </div>
          ) : (
            <Button
              size="sm"
              variant="ghost"
              className="w-full"
              onClick={() => setShowAddWidget(true)}
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
          {!isEditing && (
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
                widget.position.w >= 6
                  ? 'md:col-span-2'
                  : widget.position.w >= 9
                    ? 'md:col-span-2 lg:col-span-3'
                    : ''
              }
            >
              <DashboardWidget
                widget={widget}
                isEditing={isEditing}
                onRemove={() => removeWidget(activeDashboard.id, widget.id)}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
