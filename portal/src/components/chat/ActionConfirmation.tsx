/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import type { ChatAction } from '@/lib/portal-adapter';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';

interface ActionConfirmationProps {
  action: ChatAction;
  onConfirm: () => void;
  onCancel: () => void;
  isPending?: boolean;
}

export function ActionConfirmation({
  action,
  onConfirm,
  onCancel,
  isPending,
}: ActionConfirmationProps) {
  return (
    <Card className="border border-warning/40 bg-warning/10 p-4">
      <div className="flex flex-col gap-3">
        <div>
          <p className="text-sm font-semibold text-warning-foreground">{action.title}</p>
          <p className="text-xs text-muted-foreground">{action.summary}</p>
        </div>
        {action.impact?.resources && action.impact.resources.length > 0 && (
          <ul className="text-xs text-muted-foreground">
            {action.impact.resources.map((resource) => (
              <li key={resource.id}>
                - {resource.id} {resource.label ? `(${resource.label})` : ''}
              </li>
            ))}
          </ul>
        )}
        <div className="flex items-center justify-end gap-2">
          <Button variant="ghost" size="sm" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
          <Button variant="destructive" size="sm" onClick={onConfirm} loading={isPending}>
            Confirm
          </Button>
        </div>
      </div>
    </Card>
  );
}
