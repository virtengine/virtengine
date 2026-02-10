'use client';

import type { ChatAction } from '@/lib/portal-adapter';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

interface ActionConfirmationProps {
  action: ChatAction;
  isExecuting?: boolean;
  onConfirm: (actionId: string) => void;
  onCancel: (actionId: string) => void;
}

const severityStyles: Record<string, string> = {
  info: 'border-sky-500/40 bg-sky-500/10 text-sky-100',
  warning: 'border-amber-500/40 bg-amber-500/10 text-amber-100',
  danger: 'border-rose-500/40 bg-rose-500/10 text-rose-100',
};

export function ActionConfirmation({
  action,
  isExecuting,
  onConfirm,
  onCancel,
}: ActionConfirmationProps) {
  const preview = action.preview;

  return (
    <div className="w-full rounded-2xl border border-white/10 bg-slate-950/80 p-4 shadow-xl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.25em] text-white/40">Action required</p>
          <h3 className="mt-2 text-lg font-semibold text-white">{action.title}</h3>
        </div>
        <Badge variant="outline" className="border-white/20 text-white/70">
          {action.impact.toUpperCase()} IMPACT
        </Badge>
      </div>

      <p className="mt-3 text-sm text-white/70">{action.summary}</p>

      {preview && (
        <div
          className={cn(
            'mt-4 rounded-xl border px-3 py-3 text-sm',
            severityStyles[preview.severity] ?? severityStyles.info
          )}
        >
          <p className="font-semibold">{preview.title}</p>
          {preview.description && <p className="mt-1 text-white/70">{preview.description}</p>}
          {preview.items && preview.items.length > 0 && (
            <ul className="mt-3 space-y-1">
              {preview.items.map((item) => (
                <li key={item.label} className="flex justify-between">
                  <span className="text-white/60">{item.label}</span>
                  <span
                    className={cn(
                      'font-medium text-white',
                      item.emphasis === 'muted' && 'text-white/60'
                    )}
                  >
                    {item.value}
                  </span>
                </li>
              ))}
            </ul>
          )}
          {preview.affectedResources && preview.affectedResources.length > 0 && (
            <div className="mt-3 rounded-lg border border-white/10 bg-white/5 p-2">
              <p className="text-xs uppercase tracking-[0.2em] text-white/50">Affected</p>
              <ul className="mt-2 space-y-1 text-xs text-white/70">
                {preview.affectedResources.slice(0, 5).map((resource) => (
                  <li key={resource.id}>{resource.label}</li>
                ))}
                {preview.affectedResources.length > 5 && (
                  <li>+{preview.affectedResources.length - 5} more</li>
                )}
              </ul>
            </div>
          )}
        </div>
      )}

      <div className="mt-4 flex flex-wrap gap-3">
        <Button
          variant="destructive"
          className="flex-1"
          loading={isExecuting}
          onClick={() => onConfirm(action.id)}
        >
          Confirm & Execute
        </Button>
        <Button
          variant="outline"
          className="flex-1 border-white/20 text-white"
          onClick={() => onCancel(action.id)}
        >
          Cancel
        </Button>
      </div>
    </div>
  );
}
