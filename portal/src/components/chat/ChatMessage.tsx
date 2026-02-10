'use client';

import type { ChatMessage } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';

interface ChatMessageProps {
  message: ChatMessage;
  compact?: boolean;
}

const roleStyles: Record<string, string> = {
  user: 'bg-gradient-to-r from-sky-500 to-indigo-500 text-white self-end',
  assistant: 'bg-white/10 text-white border border-white/10',
  system: 'bg-amber-500/10 text-amber-200 border border-amber-500/30',
};

const roleLabel: Record<string, string> = {
  user: 'You',
  assistant: 'VirtEngine AI',
  system: 'System',
};

export function ChatMessage({ message, compact }: ChatMessageProps) {
  const styles = roleStyles[message.role] ?? roleStyles.assistant;
  const label = roleLabel[message.role] ?? 'Agent';

  return (
    <div
      className={cn(
        'flex w-full flex-col gap-2',
        message.role === 'user' ? 'items-end' : 'items-start'
      )}
    >
      <span className="text-[11px] uppercase tracking-[0.2em] text-white/40">{label}</span>
      <div
        className={cn(
          'max-w-[92%] rounded-2xl px-4 py-3 text-sm leading-relaxed shadow-lg backdrop-blur',
          styles,
          compact ? 'py-2 text-xs' : ''
        )}
      >
        <p className="whitespace-pre-wrap">{message.content}</p>
      </div>
    </div>
  );
}
