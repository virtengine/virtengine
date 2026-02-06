/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useState } from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { cn } from '@/lib/utils';
import { formatRelativeTime } from '@/lib/utils';
import {
  useMetricsStore,
  selectFiringAlerts,
  selectRecentAlertEvents,
} from '@/stores/metricsStore';
import { AlertConfigDialog } from '@/components/metrics/AlertConfigDialog';
import { ALERT_STATUS_VARIANT } from '@virtengine/portal/types/metrics';

export default function AlertsPage() {
  const fetchMetrics = useMetricsStore((s) => s.fetchMetrics);
  const alerts = useMetricsStore((s) => s.alerts);
  const alertEvents = useMetricsStore(selectRecentAlertEvents);
  const firingAlerts = useMetricsStore(selectFiringAlerts);
  const deleteAlert = useMetricsStore((s) => s.deleteAlert);
  const acknowledgeEvent = useMetricsStore((s) => s.acknowledgeAlertEvent);
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  useEffect(() => {
    void fetchMetrics();
  }, [fetchMetrics]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Alerts</h1>
          <p className="text-sm text-muted-foreground">
            Configure and manage metric threshold alerts
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Alert
        </Button>
      </div>

      {/* Summary */}
      <div className="flex gap-3">
        <Badge variant={firingAlerts.length > 0 ? 'destructive' : 'success'} dot>
          {firingAlerts.length} firing
        </Badge>
        <Badge variant="secondary">
          {alerts.filter((a) => a.status === 'active').length} active
        </Badge>
        <Badge variant="secondary">
          {alerts.filter((a) => a.status === 'resolved').length} resolved
        </Badge>
      </div>

      {/* Alert list */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base">Alert Rules</CardTitle>
        </CardHeader>
        <CardContent>
          {alerts.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              No alerts configured. Create one to get started.
            </p>
          ) : (
            <div className="space-y-3">
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className={cn(
                    'flex items-center justify-between rounded-lg border p-3',
                    alert.status === 'firing' &&
                      'border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/10'
                  )}
                >
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{alert.name}</span>
                      <Badge variant={ALERT_STATUS_VARIANT[alert.status]} size="sm">
                        {alert.status}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">
                      {alert.metric.toUpperCase()}{' '}
                      {alert.condition === 'gt' ? '>' : alert.condition === 'lt' ? '<' : '='}{' '}
                      {alert.threshold}% for {alert.duration}s
                      {alert.deploymentId && ` • ${alert.deploymentId}`}
                    </p>
                    {alert.lastFired && (
                      <p className="text-xs text-muted-foreground">
                        Last fired: {formatRelativeTime(new Date(alert.lastFired))}
                      </p>
                    )}
                  </div>
                  <Button size="icon-sm" variant="ghost" onClick={() => deleteAlert(alert.id)}>
                    <Trash2 className="h-4 w-4 text-muted-foreground" />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Alert history */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base">Alert History</CardTitle>
        </CardHeader>
        <CardContent>
          {alertEvents.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">No alert events yet.</p>
          ) : (
            <div className="space-y-2">
              {alertEvents.map((event) => (
                <div
                  key={event.id}
                  className="flex items-center justify-between border-b py-2 last:border-0"
                >
                  <div className="flex items-center gap-3">
                    <span
                      className={cn(
                        'text-lg',
                        event.status === 'firing' ? 'text-red-500' : 'text-green-500'
                      )}
                    >
                      {event.status === 'firing' ? '▲' : '▼'}
                    </span>
                    <div>
                      <p className="text-sm font-medium">{event.alertName}</p>
                      <p className="text-xs text-muted-foreground">
                        Value: {event.value.toFixed(1)}% •{' '}
                        {formatRelativeTime(new Date(event.timestamp))}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {event.acknowledged ? (
                      <Badge variant="secondary" size="sm">
                        Acknowledged
                      </Badge>
                    ) : (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => acknowledgeEvent(event.id)}
                      >
                        Acknowledge
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <AlertConfigDialog open={showCreateDialog} onOpenChange={setShowCreateDialog} />
    </div>
  );
}
