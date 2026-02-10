'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { MessageCircle, Sparkles, X } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Textarea } from '@/components/ui/Textarea';
import { useChain, useWallet } from '@/lib/portal-adapter';
import { getChainInfo } from '@/config';
import { useWalletTransaction } from '@/hooks/useWalletTransaction';
import { initializeChatSession, setChatRuntime, useChatStore } from '@/stores/chatStore';
import { ChatMessage } from './ChatMessage';
import { ActionConfirmation } from './ActionConfirmation';

export function ChatPanel() {
  const {
    isOpen,
    messages,
    pendingActions,
    isStreaming,
    error,
    toggleOpen,
    closeChat,
    sendMessage,
    confirmAction,
    cancelAction,
    clearError,
  } = useChatStore();

  const { queryClient } = useChain();
  const wallet = useWallet();
  const walletTx = useWalletTransaction();
  const chainInfo = useMemo(() => getChainInfo(), []);

  const [input, setInput] = useState('');
  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    initializeChatSession();
  }, []);

  useEffect(() => {
    const address = wallet.accounts[wallet.activeAccountIndex]?.address ?? null;
    setChatRuntime({
      queryClient,
      walletAddress: address,
      chainId: wallet.chainId ?? chainInfo.chainId,
      tokenDenom: chainInfo.stakeCurrency.coinMinimalDenom,
      sendTransaction: walletTx.sendTransaction,
      estimateFee: walletTx.estimateFee,
    });
  }, [
    queryClient,
    wallet.accounts,
    wallet.activeAccountIndex,
    wallet.chainId,
    chainInfo.chainId,
    chainInfo.stakeCurrency.coinMinimalDenom,
    walletTx.sendTransaction,
    walletTx.estimateFee,
  ]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [messages, pendingActions, isStreaming]);

  const visibleMessages = messages.filter((message) => message.role !== 'tool');

  const handleSend = async () => {
    if (!input.trim()) return;
    await sendMessage(input.trim());
    setInput('');
  };

  return (
    <>
      <Button
        variant="default"
        size="icon"
        className={
          'fixed bottom-6 right-6 z-40 h-12 w-12 rounded-full bg-gradient-to-r from-sky-500 to-indigo-500 text-white shadow-2xl transition-opacity ' +
          (isOpen ? 'pointer-events-none opacity-0' : 'opacity-100')
        }
        onClick={toggleOpen}
        aria-label="Open AI chat"
      >
        <MessageCircle className="h-5 w-5" />
      </Button>

      <div
        className={
          'fixed bottom-6 right-6 z-50 flex h-[78vh] w-[min(420px,92vw)] flex-col overflow-hidden rounded-3xl border border-white/10 bg-slate-950/90 shadow-2xl backdrop-blur-xl transition-all duration-300 ' +
          (isOpen ? 'translate-y-0 opacity-100' : 'pointer-events-none translate-y-8 opacity-0')
        }
        role="dialog"
        aria-hidden={!isOpen}
      >
        <div className="flex items-center justify-between border-b border-white/10 px-5 py-4">
          <div>
            <p className="text-xs uppercase tracking-[0.3em] text-white/40">VirtEngine</p>
            <div className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-sky-400" />
              <h2 className="text-lg font-semibold text-white">AI Command Center</h2>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={closeChat}>
            <X className="h-4 w-4 text-white/70" />
          </Button>
        </div>

        <div className="flex-1 space-y-4 overflow-y-auto px-4 py-6">
          {visibleMessages.length === 0 && (
            <div className="rounded-2xl border border-white/10 bg-white/5 p-4 text-sm text-white/70">
              Ask me about your deployments, balances, or governance votes. I can draft chain
              actions and walk you through confirmations.
            </div>
          )}

          {visibleMessages.map((message) => (
            <ChatMessage key={message.id} message={message} />
          ))}

          {pendingActions.map((action) => (
            <ActionConfirmation
              key={action.id}
              action={action}
              isExecuting={action.status === 'executing'}
              onConfirm={confirmAction}
              onCancel={cancelAction}
            />
          ))}

          {error && (
            <div className="rounded-2xl border border-rose-500/40 bg-rose-500/10 p-3 text-sm text-rose-200">
              {error}
              <button
                type="button"
                className="ml-3 text-xs text-rose-100 underline"
                onClick={clearError}
              >
                Dismiss
              </button>
            </div>
          )}
          <div ref={bottomRef} />
        </div>

        <div className="border-t border-white/10 bg-slate-950/90 px-4 py-4">
          <div className="flex items-end gap-3">
            <Textarea
              value={input}
              onChange={(event) => setInput(event.target.value)}
              placeholder="Ask about deployments, balances, orders..."
              rows={2}
              className="min-h-[48px] flex-1 resize-none rounded-2xl border-white/10 bg-white/5 text-sm text-white placeholder:text-white/40"
              onKeyDown={(event) => {
                if (event.key === 'Enter' && !event.shiftKey) {
                  event.preventDefault();
                  void handleSend();
                }
              }}
            />
            <Button
              variant="default"
              className="h-11 rounded-2xl bg-gradient-to-r from-sky-500 to-indigo-500 text-white"
              disabled={isStreaming || !input.trim()}
              onClick={() => void handleSend()}
            >
              {isStreaming ? 'Thinking...' : 'Send'}
            </Button>
          </div>
        </div>
      </div>
    </>
  );
}
