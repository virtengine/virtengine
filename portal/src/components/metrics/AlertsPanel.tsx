/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Alerts panel showing firing alerts and recent alert events.
 */

'use client';

import { Bell, Check } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Alert as AlertMessage, AlertDescription } from '@/components/ui/Alert';
import { cn } from '@/lib/utils';
import { formatRelativeTime } from '@/lib/utils';
import {
  useMetricsStore,
  selectFiringAlerts,
  selectRecentAlertEvents,
} from '@/stores/metricsStore';
import type { Alert, AlertEvent } from '@virtengine/portal/types/metrics';

export function AlertsPanel() {
  const firingAlerts = useMetricsStore(selectFiringAlerts);
  const recentEvents = useMetricsStore(selectRecentAlertEvents);
  const acknowledgeEvent = useMetricsStore((s) => s.acknowledgeAlertEvent);
  const alertMutationPending = useMetricsStore((s) => s.alertMutationPending);
  const alertMutationsAvailable = useMetricsStore((s) => s.alertMutationsAvailable);
  const error = useMetricsStore((s) => s.error);

  const handleAcknowledgeEvent = async (eventId: string) => {
    try {
      await acknowledgeEvent(eventId);
    } catch {
      return;
    }
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">Alerts</CardTitle>
        {alertMutationsAvailable && (
          <Badge variant={firingAlerts.length > 0 ? 'destructive' : 'secondary'}>
            {firingAlerts.length} active
          </Badge>
        )}
      </CardHeader>
      <CardContent>
        {!alertMutationsAvailable ? (
          <AlertMessage variant="warning">
            <AlertDescription>Alert changes are unavailable.</AlertDescription>
          </AlertMessage>
        ) : (
          <>
            {error && (
              <AlertMessage variant="destructive" className="mb-4">
                <AlertDescription>{error}</AlertDescription>
              </AlertMessage>
            )}
            {alertMutationPending && (
              <p role="status" className="mb-4 text-sm text-muted-foreground">
                Committing alert change…
              </p>
            )}

            {firingAlerts.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground">
                <Bell className="mx-auto mb-2 h-8 w-8 opacity-50" />
                <p className="text-sm">No active alerts</p>
              </div>
            ) : (
              <div className="space-y-3">
                {firingAlerts.map((alert) => (
                  <FiringAlertRow key={alert.id} alert={alert} />
                ))}
              </div>
            )}

            {recentEvents.length > 0 && (
              <div className="mt-4 border-t pt-4">
                <h3 className="mb-2 text-sm font-medium">Recent Events</h3>
                <div className="space-y-2">
                  {recentEvents.map((event) => (
                    <AlertEventRow
                      key={event.id}
                      event={event}
                      mutationDisabled={alertMutationPending}
                      onAcknowledge={() => handleAcknowledgeEvent(event.id)}
                    />
                  ))}
                </div>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}

function FiringAlertRow({ alert }: { alert: Alert }) {
  return (
    <div className="flex items-center justify-between rounded-lg bg-red-50 p-3 dark:bg-red-900/20">
      <div>
        <p className="font-medium text-red-700 dark:text-red-400">{alert.name}</p>
        <p className="text-sm text-red-600 dark:text-red-300">
          {alert.metric} {alert.condition === 'gt' ? '>' : alert.condition === 'lt' ? '<' : '='}{' '}
          {alert.threshold}%
        </p>
      </div>
    </div>
  );
}

function AlertEventRow({
  event,
  mutationDisabled,
  onAcknowledge,
}: {
  event: AlertEvent;
  mutationDisabled: boolean;
  onAcknowledge: () => Promise<void>;
}) {
  return (
    <div className="flex items-center justify-between text-sm">
      <div className="flex items-center gap-2">
        <span className={cn(event.status === 'firing' ? 'text-red-500' : 'text-green-500')}>
          {event.status === 'firing' ? '▲' : '▼'}
        </span>
        <span>{event.alertName}</span>
      </div>
      <div className="flex items-center gap-2">
        <span className="text-muted-foreground">
          {formatRelativeTime(new Date(event.timestamp))}
        </span>
        {!event.acknowledged && (
          <Button
            size="icon-sm"
            variant="ghost"
            aria-label={`Acknowledge ${event.alertName}`}
            disabled={mutationDisabled}
            onClick={() => void onAcknowledge()}
          >
            <Check className="h-3 w-3" />
          </Button>
        )}
      </div>
    </div>
  );
}
