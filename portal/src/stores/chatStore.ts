import { create } from 'zustand';
import type {
  ChatAction,
  ChatAgentResult,
  ChatMessage,
  ChatRuntimeContext,
} from '@/lib/portal-adapter';
import { createChatAgent, OpenAIProvider, LocalLLMProvider } from '@/lib/portal-adapter';
import { llmConfig } from '@/lib/config';
import type { TransactionResult } from '@/hooks/useWalletTransaction';

const STORAGE_KEY = 've_chat_session_v1';

interface ChatRuntime extends ChatRuntimeContext {
  sendTransaction?: (msgs: unknown[], options?: { memo?: string }) => Promise<TransactionResult>;
  estimateFee?: (gasLimit: number) => {
    amount: Array<{ denom: string; amount: string }>;
    gas: string;
  };
}

let runtime: ChatRuntime = {};
let activeController: AbortController | null = null;

function createMessageId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function loadSession(): { messages: ChatMessage[]; pendingActions: ChatAction[] } | null {
  if (typeof window === 'undefined') return null;
  const stored = window.sessionStorage.getItem(STORAGE_KEY);
  if (!stored) return null;
  try {
    const parsed = JSON.parse(stored) as {
      messages?: ChatMessage[];
      pendingActions?: ChatAction[];
    };
    return {
      messages: parsed.messages ?? [],
      pendingActions: parsed.pendingActions ?? [],
    };
  } catch {
    return null;
  }
}

function persistSession(state: { messages: ChatMessage[]; pendingActions: ChatAction[] }) {
  if (typeof window === 'undefined') return;
  window.sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
}

function buildProvider() {
  if (llmConfig.provider === 'local') {
    return new LocalLLMProvider({
      endpoint: llmConfig.localEndpoint,
      defaultModel: llmConfig.model,
    });
  }

  return new OpenAIProvider({
    endpoint: llmConfig.endpoint,
    apiKey: llmConfig.apiKey,
    defaultModel: llmConfig.model,
  });
}

export interface ChatStoreState {
  isOpen: boolean;
  messages: ChatMessage[];
  pendingActions: ChatAction[];
  isStreaming: boolean;
  error: string | null;
}

export interface ChatStoreActions {
  toggleOpen: () => void;
  closeChat: () => void;
  sendMessage: (content: string) => Promise<void>;
  confirmAction: (actionId: string) => Promise<void>;
  cancelAction: (actionId: string) => void;
  clearError: () => void;
}

export type ChatStore = ChatStoreState & ChatStoreActions;

const initialState: ChatStoreState = {
  isOpen: false,
  messages: [],
  pendingActions: [],
  isStreaming: false,
  error: null,
};

export const useChatStore = create<ChatStore>()((set, get) => ({
  ...initialState,

  toggleOpen: () => set((state) => ({ isOpen: !state.isOpen })),
  closeChat: () => set({ isOpen: false }),

  sendMessage: async (content: string) => {
    const provider = buildProvider();
    const agent = createChatAgent({ provider, model: llmConfig.model });

    if (!runtime.queryClient) {
      set({ error: 'Chain connection is not ready yet.' });
      return;
    }

    const userMessage: ChatMessage = {
      id: createMessageId('user'),
      role: 'user',
      content,
      createdAt: Date.now(),
    };

    const assistantId = createMessageId('assistant');
    const assistantPlaceholder: ChatMessage = {
      id: assistantId,
      role: 'assistant',
      content: '',
      createdAt: Date.now(),
      status: 'streaming',
    };

    set((state) => {
      const nextMessages = [...state.messages, userMessage, assistantPlaceholder];
      persistSession({ messages: nextMessages, pendingActions: state.pendingActions });
      return { messages: nextMessages, isStreaming: true, error: null };
    });

    if (activeController) {
      activeController.abort();
    }
    activeController = new AbortController();

    let streamedContent = '';

    try {
      const result: ChatAgentResult = await agent.processMessage({
        messages: [...get().messages.filter((message) => message.id !== assistantId)],
        runtime,
        signal: activeController.signal,
        onStreamToken: (token) => {
          streamedContent += token;
          set((state) => ({
            messages: state.messages.map((message) =>
              message.id === assistantId
                ? { ...message, content: streamedContent, status: 'streaming' as const }
                : message
            ),
          }));
        },
      });

      set((state) => {
        const nextMessages = state.messages.map((message) =>
          message.id === assistantId
            ? {
                ...result.assistantMessage,
                id: assistantId,
                content: streamedContent || result.assistantMessage.content,
              }
            : message
        );

        const nextActions = [...state.pendingActions, ...result.actions];
        persistSession({ messages: nextMessages, pendingActions: nextActions });

        return {
          messages: nextMessages,
          pendingActions: nextActions,
          isStreaming: false,
        };
      });
    } catch (error) {
      set((state) => {
        const nextMessages = state.messages.map((message) =>
          message.id === assistantId
            ? {
                ...message,
                content: 'Sorry, I ran into an error while processing that request.',
                status: 'error' as const,
              }
            : message
        );
        persistSession({ messages: nextMessages, pendingActions: state.pendingActions });
        return {
          messages: nextMessages,
          isStreaming: false,
          error: error instanceof Error ? error.message : 'Chat request failed',
        };
      });
    }
  },

  confirmAction: async (actionId: string) => {
    const action = get().pendingActions.find((item) => item.id === actionId);
    if (!action) return;

    if (action.requiresWallet && !runtime.sendTransaction) {
      set({ error: 'Wallet transaction handler not available.' });
      return;
    }

    set((state) => ({
      pendingActions: state.pendingActions.map((item) =>
        item.id === actionId ? { ...item, status: 'executing' as const } : item
      ),
    }));

    try {
      let result: TransactionResult | null = null;
      if (runtime.sendTransaction && action.messages.length > 0) {
        result = await runtime.sendTransaction(action.messages);
      }

      const systemMessage: ChatMessage = {
        id: createMessageId('system'),
        role: 'system',
        content: result
          ? `Action executed. Transaction hash: ${result.txHash}`
          : 'Action confirmed and queued for execution.',
        createdAt: Date.now(),
      };

      set((state) => {
        const nextActions = state.pendingActions.map((item) =>
          item.id === actionId ? { ...item, status: 'completed' as const } : item
        );
        const nextMessages = [...state.messages, systemMessage];
        persistSession({ messages: nextMessages, pendingActions: nextActions });
        return { pendingActions: nextActions, messages: nextMessages };
      });
    } catch (error) {
      set((state) => {
        const nextActions = state.pendingActions.map((item) =>
          item.id === actionId
            ? {
                ...item,
                status: 'failed' as const,
                error: error instanceof Error ? error.message : 'Action failed',
              }
            : item
        );
        persistSession({ messages: state.messages, pendingActions: nextActions });
        return {
          pendingActions: nextActions,
          error: error instanceof Error ? error.message : 'Action failed',
        };
      });
    }
  },

  cancelAction: (actionId: string) => {
    set((state) => {
      const nextActions = state.pendingActions.filter((item) => item.id !== actionId);
      const cancelMessage: ChatMessage = {
        id: createMessageId('system'),
        role: 'system',
        content: 'Action cancelled.',
        createdAt: Date.now(),
      };
      const nextMessages = [...state.messages, cancelMessage];
      persistSession({ messages: nextMessages, pendingActions: nextActions });
      return { pendingActions: nextActions, messages: nextMessages };
    });
  },

  clearError: () => set({ error: null }),
}));

export function setChatRuntime(update: Partial<ChatRuntime>) {
  runtime = { ...runtime, ...update };
}

export function initializeChatSession() {
  const stored = loadSession();
  if (!stored) return;
  useChatStore.setState({
    messages: stored.messages,
    pendingActions: stored.pendingActions,
  });
}
