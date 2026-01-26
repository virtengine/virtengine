/**
 * Support Hook
 * VE-707: Support ticket management with encryption
 *
 * CRITICAL: All ticket content is encrypted. Never log decrypted content.
 */
import * as React from 'react';
import type {
  SupportState,
  SupportTicket,
  TicketMessage,
  TicketFilter,
  CreateTicketRequest,
  DecryptedTicketContent,
  SupportStats,
} from '../types/support';

/**
 * Support context value
 */
interface SupportContextValue {
  /**
   * Support state
   */
  state: SupportState;

  /**
   * Load tickets
   */
  loadTickets: (filter?: TicketFilter) => Promise<void>;

  /**
   * Select ticket and decrypt content
   */
  selectTicket: (ticket: SupportTicket | null) => Promise<void>;

  /**
   * Create new ticket
   */
  createTicket: (request: CreateTicketRequest) => Promise<string>;

  /**
   * Send reply to ticket
   */
  sendReply: (ticketId: string, message: string, attachments?: File[]) => Promise<void>;

  /**
   * Update ticket status
   */
  updateStatus: (ticketId: string, status: string) => Promise<void>;

  /**
   * Assign ticket
   */
  assignTicket: (ticketId: string, agentId: string) => Promise<void>;

  /**
   * Close ticket
   */
  closeTicket: (ticketId: string, resolution: string) => Promise<void>;
}

const SupportContext = React.createContext<SupportContextValue | null>(null);

/**
 * Initial support state
 */
const initialState: SupportState = {
  tickets: [],
  selectedTicket: null,
  messages: [],
  decryptedContent: null,
  stats: {
    openTickets: 0,
    byStatus: {
      open: 0,
      in_progress: 0,
      waiting_customer: 0,
      waiting_internal: 0,
      resolved: 0,
      closed: 0,
    },
    byCategory: {
      account: 0,
      identity: 0,
      billing: 0,
      provider: 0,
      marketplace: 0,
      technical: 0,
      security: 0,
      other: 0,
    },
    byPriority: { low: 0, medium: 0, high: 0, urgent: 0 },
    avgFirstResponseTime: 0,
    avgResolutionTime: 0,
  },
  isLoading: false,
  error: null,
};

/**
 * Support action
 */
type SupportAction =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_TICKETS'; payload: SupportTicket[] }
  | { type: 'SET_SELECTED'; payload: SupportTicket | null }
  | { type: 'SET_MESSAGES'; payload: TicketMessage[] }
  | { type: 'SET_DECRYPTED'; payload: DecryptedTicketContent | null }
  | { type: 'ADD_MESSAGE'; payload: TicketMessage }
  | { type: 'SET_STATS'; payload: SupportStats }
  | { type: 'UPDATE_TICKET'; payload: SupportTicket }
  | { type: 'SET_ERROR'; payload: Error | null };

/**
 * Support reducer
 */
function supportReducer(state: SupportState, action: SupportAction): SupportState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_TICKETS':
      return { ...state, tickets: action.payload, isLoading: false };
    case 'SET_SELECTED':
      return {
        ...state,
        selectedTicket: action.payload,
        messages: [],
        decryptedContent: null,
      };
    case 'SET_MESSAGES':
      return { ...state, messages: action.payload };
    case 'SET_DECRYPTED':
      return { ...state, decryptedContent: action.payload };
    case 'ADD_MESSAGE':
      return { ...state, messages: [...state.messages, action.payload] };
    case 'SET_STATS':
      return { ...state, stats: action.payload };
    case 'UPDATE_TICKET': {
      const updated = action.payload;
      return {
        ...state,
        tickets: state.tickets.map((t) => (t.id === updated.id ? updated : t)),
        selectedTicket:
          state.selectedTicket?.id === updated.id ? updated : state.selectedTicket,
      };
    }
    case 'SET_ERROR':
      return { ...state, error: action.payload, isLoading: false };
    default:
      return state;
  }
}

/**
 * Support provider props
 */
interface SupportProviderProps {
  children: React.ReactNode;
}

/**
 * Support provider
 */
export function SupportProvider({ children }: SupportProviderProps): JSX.Element {
  const [state, dispatch] = React.useReducer(supportReducer, initialState);

  /**
   * Load tickets
   */
  const loadTickets = React.useCallback(async (filter?: TicketFilter) => {
    dispatch({ type: 'SET_LOADING', payload: true });

    try {
      const params = new URLSearchParams();
      if (filter?.status) {
        const statuses = Array.isArray(filter.status) ? filter.status : [filter.status];
        statuses.forEach((s) => params.append('status', s));
      }
      if (filter?.category) params.append('category', filter.category);
      if (filter?.priority) params.append('priority', filter.priority);
      if (filter?.assignedTo) params.append('assignedTo', filter.assignedTo);
      if (filter?.unassignedOnly) params.append('unassignedOnly', 'true');
      if (filter?.submitterAddress) params.append('submitter', filter.submitterAddress);
      if (filter?.query) params.append('q', filter.query);
      if (filter?.page) params.append('page', filter.page.toString());
      if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

      const response = await fetch(`/api/admin/support?${params.toString()}`, {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to load tickets');
      }

      const data = await response.json();
      dispatch({ type: 'SET_TICKETS', payload: data.tickets });
      dispatch({ type: 'SET_STATS', payload: data.stats });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error : new Error('Unknown error'),
      });
    }
  }, []);

  /**
   * Select ticket and decrypt content
   * CRITICAL: Decryption happens server-side, we only get decrypted data
   */
  const selectTicket = React.useCallback(
    async (ticket: SupportTicket | null) => {
      dispatch({ type: 'SET_SELECTED', payload: ticket });

      if (!ticket) {
        dispatch({ type: 'SET_DECRYPTED', payload: null });
        return;
      }

      try {
        // Fetch decrypted content (decryption happens server-side)
        const [contentRes, messagesRes] = await Promise.all([
          fetch(`/api/admin/support/${ticket.id}/content`, {
            credentials: 'include',
          }),
          fetch(`/api/admin/support/${ticket.id}/messages`, {
            credentials: 'include',
          }),
        ]);

        if (!contentRes.ok || !messagesRes.ok) {
          throw new Error('Failed to load ticket details');
        }

        const content = await contentRes.json();
        const messages = await messagesRes.json();

        // CRITICAL: Content is now decrypted, handle carefully
        dispatch({ type: 'SET_DECRYPTED', payload: content });
        dispatch({ type: 'SET_MESSAGES', payload: messages.messages });
      } catch (error) {
        dispatch({
          type: 'SET_ERROR',
          payload: error instanceof Error ? error : new Error('Unknown error'),
        });
      }
    },
    []
  );

  /**
   * Create new ticket
   */
  const createTicket = React.useCallback(
    async (request: CreateTicketRequest): Promise<string> => {
      const formData = new FormData();
      formData.append('category', request.category);
      formData.append('priority', request.priority);
      formData.append('subject', request.subject);
      formData.append('description', request.description);

      if (request.relatedEntity) {
        formData.append('relatedEntityType', request.relatedEntity.type);
        formData.append('relatedEntityId', request.relatedEntity.id);
      }

      if (request.attachments) {
        request.attachments.forEach((file) => {
          formData.append('attachments', file);
        });
      }

      const response = await fetch('/api/support/tickets', {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to create ticket');
      }

      const data = await response.json();
      return data.ticketId;
    },
    []
  );

  /**
   * Send reply to ticket
   */
  const sendReply = React.useCallback(
    async (ticketId: string, message: string, attachments?: File[]) => {
      const formData = new FormData();
      formData.append('message', message);

      if (attachments) {
        attachments.forEach((file) => {
          formData.append('attachments', file);
        });
      }

      const response = await fetch(`/api/admin/support/${ticketId}/reply`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to send reply');
      }

      const newMessage = await response.json();
      dispatch({ type: 'ADD_MESSAGE', payload: newMessage });
    },
    []
  );

  /**
   * Update ticket status
   */
  const updateStatus = React.useCallback(
    async (ticketId: string, status: string) => {
      const response = await fetch(`/api/admin/support/${ticketId}/status`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
      });

      if (!response.ok) {
        throw new Error('Failed to update status');
      }

      const updated = await response.json();
      dispatch({ type: 'UPDATE_TICKET', payload: updated });
    },
    []
  );

  /**
   * Assign ticket
   */
  const assignTicket = React.useCallback(
    async (ticketId: string, agentId: string) => {
      const response = await fetch(`/api/admin/support/${ticketId}/assign`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agentId }),
      });

      if (!response.ok) {
        throw new Error('Failed to assign ticket');
      }

      const updated = await response.json();
      dispatch({ type: 'UPDATE_TICKET', payload: updated });
    },
    []
  );

  /**
   * Close ticket
   */
  const closeTicket = React.useCallback(
    async (ticketId: string, resolution: string) => {
      const response = await fetch(`/api/admin/support/${ticketId}/close`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ resolution }),
      });

      if (!response.ok) {
        throw new Error('Failed to close ticket');
      }

      const updated = await response.json();
      dispatch({ type: 'UPDATE_TICKET', payload: updated });
    },
    []
  );

  const value: SupportContextValue = {
    state,
    loadTickets,
    selectTicket,
    createTicket,
    sendReply,
    updateStatus,
    assignTicket,
    closeTicket,
  };

  return (
    <SupportContext.Provider value={value}>
      {children}
    </SupportContext.Provider>
  );
}

/**
 * Use support hook
 */
export function useSupport(): SupportContextValue {
  const context = React.useContext(SupportContext);
  if (!context) {
    throw new Error('useSupport must be used within a SupportProvider');
  }
  return context;
}
