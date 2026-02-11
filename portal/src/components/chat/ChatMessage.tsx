/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import type { ChatMessage as ChatMessageType } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';

interface ChatMessageProps {
  message: ChatMessageType;
}

const roleStyles: Record<string, string> = {
  user: 'bg-primary text-primary-foreground self-end',
  assistant: 'bg-card text-foreground border border-border',
  system: 'bg-muted text-muted-foreground border border-border',
  tool: 'bg-secondary text-secondary-foreground border border-border',
};

export function ChatMessage({ message }: ChatMessageProps) {
  const style = roleStyles[message.role] ?? roleStyles.assistant;

  return (
    <div
      className={cn(
        'flex max-w-[85%] flex-col gap-1',
        message.role === 'user' ? 'self-end' : 'self-start'
      )}
    >
      <div className={cn('rounded-lg px-4 py-3 text-sm leading-relaxed shadow-sm', style)}>
        <p className="whitespace-pre-wrap">{message.content}</p>
      </div>
      <span className="text-[11px] text-muted-foreground">
        {new Date(message.createdAt).toLocaleTimeString()}
      </span>
    </div>
  );
}
