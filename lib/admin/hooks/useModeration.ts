/**
 * Moderation Hook
 * VE-706: Content moderation actions
 */
import * as React from 'react';
import type {
  ModerationState,
  ModerationItem,
  ModerationQueueFilter,
  ModerationDecision,
  ModerationStats,
} from '../types/moderation';

/**
 * Moderation context value
 */
interface ModerationContextValue {
  /**
   * Moderation state
   */
  state: ModerationState;

  /**
   * Load moderation queue
   */
  loadQueue: (filter?: ModerationQueueFilter) => Promise<void>;

  /**
   * Select moderation item
   */
  selectItem: (item: ModerationItem | null) => void;

  /**
   * Assign item to self
   */
  assignToSelf: (itemId: string) => Promise<void>;

  /**
   * Submit decision
   */
  submitDecision: (itemId: string, decision: ModerationDecision) => Promise<void>;

  /**
   * Escalate item
   */
  escalate: (itemId: string, reason: string) => Promise<void>;

  /**
   * Add comment
   */
  addComment: (itemId: string, comment: string) => Promise<void>;
}

const ModerationContext = React.createContext<ModerationContextValue | null>(null);

/**
 * Initial moderation state
 */
const initialState: ModerationState = {
  queue: [],
  selectedItem: null,
  stats: {
    pending: 0,
    byType: {
      provider_application: 0,
      offering: 0,
      user_report: 0,
      content_flag: 0,
    },
    byPriority: { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0 },
    avgResolutionTime: 0,
  },
  isLoading: false,
  error: null,
};

/**
 * Moderation action
 */
type ModerationAction =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_QUEUE'; payload: ModerationItem[] }
  | { type: 'SET_SELECTED'; payload: ModerationItem | null }
  | { type: 'SET_STATS'; payload: ModerationStats }
  | { type: 'UPDATE_ITEM'; payload: ModerationItem }
  | { type: 'REMOVE_ITEM'; payload: string }
  | { type: 'SET_ERROR'; payload: Error | null };

/**
 * Moderation reducer
 */
function moderationReducer(
  state: ModerationState,
  action: ModerationAction
): ModerationState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_QUEUE':
      return { ...state, queue: action.payload, isLoading: false };
    case 'SET_SELECTED':
      return { ...state, selectedItem: action.payload };
    case 'SET_STATS':
      return { ...state, stats: action.payload };
    case 'UPDATE_ITEM': {
      const updated = action.payload;
      return {
        ...state,
        queue: state.queue.map((item) =>
          item.id === updated.id ? updated : item
        ),
        selectedItem:
          state.selectedItem?.id === updated.id ? updated : state.selectedItem,
      };
    }
    case 'REMOVE_ITEM':
      return {
        ...state,
        queue: state.queue.filter((item) => item.id !== action.payload),
        selectedItem:
          state.selectedItem?.id === action.payload ? null : state.selectedItem,
      };
    case 'SET_ERROR':
      return { ...state, error: action.payload, isLoading: false };
    default:
      return state;
  }
}

/**
 * Moderation provider props
 */
interface ModerationProviderProps {
  children: React.ReactNode;
}

/**
 * Moderation provider
 */
export function ModerationProvider({ children }: ModerationProviderProps): JSX.Element {
  const [state, dispatch] = React.useReducer(moderationReducer, initialState);

  /**
   * Load moderation queue
   */
  const loadQueue = React.useCallback(
    async (filter?: ModerationQueueFilter) => {
      dispatch({ type: 'SET_LOADING', payload: true });

      try {
        const params = new URLSearchParams();
        if (filter?.type) params.append('type', filter.type);
        if (filter?.status) params.append('status', filter.status);
        if (filter?.priority) params.append('priority', filter.priority.toString());
        if (filter?.assignedTo) params.append('assignedTo', filter.assignedTo);
        if (filter?.unassignedOnly) params.append('unassignedOnly', 'true');
        if (filter?.page) params.append('page', filter.page.toString());
        if (filter?.pageSize) params.append('pageSize', filter.pageSize.toString());

        const response = await fetch(`/api/admin/moderation?${params.toString()}`, {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error('Failed to load moderation queue');
        }

        const data = await response.json();
        dispatch({ type: 'SET_QUEUE', payload: data.items });
        dispatch({ type: 'SET_STATS', payload: data.stats });
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
   * Select moderation item
   */
  const selectItem = React.useCallback((item: ModerationItem | null) => {
    dispatch({ type: 'SET_SELECTED', payload: item });
  }, []);

  /**
   * Assign item to self
   */
  const assignToSelf = React.useCallback(async (itemId: string) => {
    const response = await fetch(`/api/admin/moderation/${itemId}/assign`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to assign item');
    }

    const updated = await response.json();
    dispatch({ type: 'UPDATE_ITEM', payload: updated });
  }, []);

  /**
   * Submit decision
   */
  const submitDecision = React.useCallback(
    async (itemId: string, decision: ModerationDecision) => {
      const response = await fetch(`/api/admin/moderation/${itemId}/decide`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(decision),
      });

      if (!response.ok) {
        throw new Error('Failed to submit decision');
      }

      // Remove from queue as it's no longer pending
      dispatch({ type: 'REMOVE_ITEM', payload: itemId });
    },
    []
  );

  /**
   * Escalate item
   */
  const escalate = React.useCallback(
    async (itemId: string, reason: string) => {
      const response = await fetch(`/api/admin/moderation/${itemId}/escalate`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ reason }),
      });

      if (!response.ok) {
        throw new Error('Failed to escalate item');
      }

      const updated = await response.json();
      dispatch({ type: 'UPDATE_ITEM', payload: updated });
    },
    []
  );

  /**
   * Add comment
   */
  const addComment = React.useCallback(
    async (itemId: string, comment: string) => {
      const response = await fetch(`/api/admin/moderation/${itemId}/comment`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ comment }),
      });

      if (!response.ok) {
        throw new Error('Failed to add comment');
      }

      const updated = await response.json();
      dispatch({ type: 'UPDATE_ITEM', payload: updated });
    },
    []
  );

  const value: ModerationContextValue = {
    state,
    loadQueue,
    selectItem,
    assignToSelf,
    submitDecision,
    escalate,
    addComment,
  };

  return (
    <ModerationContext.Provider value={value}>
      {children}
    </ModerationContext.Provider>
  );
}

/**
 * Use moderation hook
 */
export function useModeration(): ModerationContextValue {
  const context = React.useContext(ModerationContext);
  if (!context) {
    throw new Error('useModeration must be used within a ModerationProvider');
  }
  return context;
}
