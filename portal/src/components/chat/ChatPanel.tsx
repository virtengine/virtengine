/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { Bot, MessageCircle, X } from 'lucide-react';

import {
  createChatAgent,
  createChatProvider,
  createDefaultChatTools,
  type ChatAction,
  type ChatMessage as ChatMessageType,
} from '@/lib/portal-adapter';
import { buildChatContext } from '@/lib/portal-adapter';
import { chatConfig, chainConfig, env } from '@/config';
import { useWallet, useChain } from '@/lib/portal-adapter';
import { useIdentityStore } from '@/stores/identityStore';
import { useChatStore } from '@/stores/chatStore';
import { ChatMessage } from './ChatMessage';
import { ActionConfirmation } from './ActionConfirmation';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/utils';
import { useWalletTransaction } from '@/hooks/useWalletTransaction';

const createMessage = (role: ChatMessageType['role'], content: string): ChatMessageType => ({
  id: `chat-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  role,
  content,
  createdAt: Date.now(),
});

export function ChatPanel() {
  const {
    isOpen,
    toggle,
    messages,
    addMessage,
    updateMessage,
    addAction,
    setPendingAction,
    setStreaming,
    isStreaming,
    pendingActionId,
    setError,
    error,
  } = useChatStore();
  const pendingAction = useChatStore((state) =>
    state.actions.find((action) => action.id === state.pendingActionId)
  );

  const [input, setInput] = useState('');
  const [isExecuting, setIsExecuting] = useState(false);
  const wallet = useWallet();
  const chain = useChain();
  const identity = useIdentityStore();
  const { sendTransaction, estimateFee } = useWalletTransaction();

  const context = useMemo(
    () =>
      buildChatContext({
        wallet,
        chain: chain.state,
        chainConfig,
        identity: {
          score: identity.veidScore,
          status: identity.isVerified ? 'verified' : 'unverified',
        },
      }),
    [wallet, chain.state, identity.veidScore, identity.isVerified]
  );

  const provider = useMemo(() => createChatProvider(chatConfig), []);
  const tools = useMemo(() => createDefaultChatTools(), []);
  const agent = useMemo(
    () => createChatAgent({ provider, toolHandlers: tools, context }),
    [provider, tools, context]
  );

  const scrollRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, pendingActionId, isStreaming]);

  if (!env.enableChat) {
    return null;
  }

  const handleSend = async () => {
    const trimmed = input.trim();
    if (!trimmed || isStreaming) return;

    const userMessage = createMessage('user', trimmed);
    addMessage(userMessage);
    setInput('');

    const assistantPlaceholder = createMessage('assistant', '');
    addMessage(assistantPlaceholder);

    setStreaming(true);
    try {
      const result = await agent.run([...messages, userMessage], (delta) => {
        updateMessage(assistantPlaceholder.id, assistantPlaceholder.content + delta);
        assistantPlaceholder.content += delta;
      });

      if (!assistantPlaceholder.content) {
        updateMessage(assistantPlaceholder.id, result.assistantMessage.content);
      }

      result.toolMessages.forEach(addMessage);
      result.actions.forEach((action) => {
        addAction(action);
        if (action.requiresConfirmation) {
          setPendingAction(action.id);
        }
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Chat request failed');
      updateMessage(assistantPlaceholder.id, 'Sorry, something went wrong.');
    } finally {
      setStreaming(false);
    }
  };

  const handleExecute = async (action: ChatAction) => {
    setIsExecuting(true);
    try {
      if (action.payload.kind === 'transaction') {
        const fee = estimateFee(200000);
        const result = await sendTransaction(action.payload.msgs, { memo: action.payload.memo });
        addMessage(
          createMessage(
            'system',
            `Transaction submitted: ${result.txHash}. Estimated fee: ${fee.amount[0]?.amount ?? '0'} ${fee.amount[0]?.denom ?? ''}`
          )
        );
      } else if (action.payload.kind === 'provider-action') {
        const result = await agent.executeAction(action);
        addMessage(createMessage('system', result.summary));
      } else {
        addMessage(createMessage('system', 'Action executed.'));
      }
    } catch (err) {
      addMessage(createMessage('system', err instanceof Error ? err.message : 'Action failed.'));
    } finally {
      setIsExecuting(false);
      setPendingAction(undefined);
    }
  };

  return (
    <div>
      <div className="fixed bottom-6 right-6 z-40">
        <Button size="icon" onClick={toggle} variant="secondary" aria-label="Toggle chat">
          {isOpen ? <X className="h-4 w-4" /> : <MessageCircle className="h-4 w-4" />}
        </Button>
      </div>

      <div
        className={cn(
          'fixed right-0 top-0 z-30 flex h-full w-full max-w-md flex-col border-l border-border bg-background shadow-2xl transition-transform duration-300',
          isOpen ? 'translate-x-0' : 'translate-x-full'
        )}
        role="dialog"
        aria-label="VirtEngine chat"
      >
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <div className="flex items-center gap-2">
            <Bot className="h-4 w-4 text-primary" />
            <span className="text-sm font-semibold">VirtEngine AI Agent</span>
          </div>
          <Button variant="ghost" size="icon-sm" onClick={toggle} aria-label="Close chat">
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div ref={scrollRef} className="flex flex-1 flex-col gap-4 overflow-y-auto px-4 py-4">
          {messages.length === 0 && (
            <div className="rounded-lg border border-dashed border-border bg-muted/20 px-4 py-6 text-sm text-muted-foreground">
              Ask about deployments, orders, balances, or request actions. Destructive actions
              require confirmation.
            </div>
          )}
          {messages.map((message) => (
            <ChatMessage key={message.id} message={message} />
          ))}

          {pendingAction && (
            <ActionConfirmation
              action={pendingAction}
              onConfirm={() => void handleExecute(pendingAction)}
              onCancel={() => setPendingAction(undefined)}
              isPending={isExecuting}
            />
          )}

          {error && (
            <div className="rounded-lg border border-destructive/50 bg-destructive/10 px-3 py-2 text-xs text-destructive">
              {error}
            </div>
          )}
        </div>

        <div className="border-t border-border p-4">
          <form
            onSubmit={(event) => {
              event.preventDefault();
              void handleSend();
            }}
            className="flex items-center gap-2"
          >
            <Input
              value={input}
              onChange={(event) => setInput(event.target.value)}
              placeholder="Ask the agent..."
              disabled={isStreaming}
            />
            <Button type="submit" disabled={isStreaming || !input.trim()}>
              Send
            </Button>
          </form>
        </div>
      </div>
    </div>
  );
}
