import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { ChatAction, ChatMessage } from '@/lib/portal-adapter';

export interface ChatState {
  isOpen: boolean;
  isStreaming: boolean;
  messages: ChatMessage[];
  actions: ChatAction[];
  pendingActionId?: string;
  error?: string | null;
}

export interface ChatActions {
  open: () => void;
  close: () => void;
  toggle: () => void;
  setStreaming: (value: boolean) => void;
  addMessage: (message: ChatMessage) => void;
  updateMessage: (id: string, content: string) => void;
  addAction: (action: ChatAction) => void;
  setPendingAction: (id?: string) => void;
  clearError: () => void;
  setError: (error: string) => void;
  reset: () => void;
}

export type ChatStore = ChatState & ChatActions;

const initialState: ChatState = {
  isOpen: false,
  isStreaming: false,
  messages: [],
  actions: [],
  pendingActionId: undefined,
  error: null,
};

export const useChatStore = create<ChatStore>()(
  persist(
    (set, _get) => ({
      ...initialState,
      open: () => set({ isOpen: true }),
      close: () => set({ isOpen: false }),
      toggle: () => set((state) => ({ isOpen: !state.isOpen })),
      setStreaming: (value) => set({ isStreaming: value }),
      addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
      updateMessage: (id, content) =>
        set((state) => ({
          messages: state.messages.map((msg) => (msg.id === id ? { ...msg, content } : msg)),
        })),
      addAction: (action) => set((state) => ({ actions: [...state.actions, action] })),
      setPendingAction: (id) => set({ pendingActionId: id }),
      clearError: () => set({ error: null }),
      setError: (error) => set({ error }),
      reset: () => set(initialState),
    }),
    {
      name: 'portal-chat',
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({
        isOpen: state.isOpen,
        messages: state.messages,
        actions: state.actions,
      }),
    }
  )
);

export const selectPendingAction = (state: ChatStore) =>
  state.actions.find((action) => action.id === state.pendingActionId);
