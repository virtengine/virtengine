/**
 * Admin Hook
 * VE-706: Admin context and actions
 */
import * as React from 'react';
import type {
  AdminState,
  AdminUser,
  DashboardStats,
  AuditLogEntry,
  AuditLogFilter,
} from '../types/admin';

/**
 * Admin context value
 */
interface AdminContextValue {
  /**
   * Admin state
   */
  state: AdminState;

  /**
   * Load dashboard stats
   */
  loadStats: () => Promise<void>;

  /**
   * Get audit log entries
   */
  getAuditLog: (filter: AuditLogFilter) => Promise<AuditLogEntry[]>;

  /**
   * Check if admin has permission
   */
  hasPermission: (permission: string) => boolean;

  /**
   * Logout admin
   */
  logout: () => Promise<void>;
}

const AdminContext = React.createContext<AdminContextValue | null>(null);

/**
 * Initial admin state
 */
const initialState: AdminState = {
  adminUser: null,
  isAuthenticated: false,
  requiresMFA: false,
  isLoading: true,
  error: null,
  stats: null,
};

/**
 * Admin action
 */
type AdminAction =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ADMIN_USER'; payload: AdminUser | null }
  | { type: 'SET_REQUIRES_MFA'; payload: boolean }
  | { type: 'SET_STATS'; payload: DashboardStats }
  | { type: 'SET_ERROR'; payload: Error | null }
  | { type: 'LOGOUT' };

/**
 * Admin reducer
 */
function adminReducer(state: AdminState, action: AdminAction): AdminState {
  switch (action.type) {
    case 'SET_LOADING':
      return { ...state, isLoading: action.payload };
    case 'SET_ADMIN_USER':
      return {
        ...state,
        adminUser: action.payload,
        isAuthenticated: !!action.payload,
        isLoading: false,
      };
    case 'SET_REQUIRES_MFA':
      return { ...state, requiresMFA: action.payload };
    case 'SET_STATS':
      return { ...state, stats: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload, isLoading: false };
    case 'LOGOUT':
      return { ...initialState, isLoading: false };
    default:
      return state;
  }
}

/**
 * Admin provider props
 */
interface AdminProviderProps {
  children: React.ReactNode;
}

/**
 * Admin provider
 */
export function AdminProvider({ children }: AdminProviderProps): JSX.Element {
  const [state, dispatch] = React.useReducer(adminReducer, initialState);

  // Check admin session on mount
  React.useEffect(() => {
    const checkSession = async () => {
      try {
        const response = await fetch('/api/admin/session', {
          credentials: 'include',
        });

        if (response.ok) {
          const data = await response.json();
          dispatch({ type: 'SET_ADMIN_USER', payload: data.adminUser });

          if (data.requiresMFA && !data.mfaVerified) {
            dispatch({ type: 'SET_REQUIRES_MFA', payload: true });
          }
        } else {
          dispatch({ type: 'SET_ADMIN_USER', payload: null });
        }
      } catch {
        dispatch({ type: 'SET_ADMIN_USER', payload: null });
      }
    };

    checkSession();
  }, []);

  /**
   * Load dashboard stats
   */
  const loadStats = React.useCallback(async () => {
    dispatch({ type: 'SET_LOADING', payload: true });

    try {
      const response = await fetch('/api/admin/stats', {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to load stats');
      }

      const stats = await response.json();
      dispatch({ type: 'SET_STATS', payload: stats });
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error : new Error('Unknown error'),
      });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, []);

  /**
   * Get audit log entries
   */
  const getAuditLog = React.useCallback(
    async (filter: AuditLogFilter): Promise<AuditLogEntry[]> => {
      const params = new URLSearchParams();
      if (filter.adminUserId) params.append('adminUserId', filter.adminUserId);
      if (filter.action) params.append('action', filter.action);
      if (filter.resourceType) params.append('resourceType', filter.resourceType);
      if (filter.resourceId) params.append('resourceId', filter.resourceId);
      if (filter.startTime) params.append('startTime', filter.startTime.toString());
      if (filter.endTime) params.append('endTime', filter.endTime.toString());
      if (filter.page) params.append('page', filter.page.toString());
      if (filter.pageSize) params.append('pageSize', filter.pageSize.toString());

      const response = await fetch(`/api/admin/audit?${params.toString()}`, {
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to fetch audit log');
      }

      const data = await response.json();
      return data.entries;
    },
    []
  );

  /**
   * Check if admin has permission
   */
  const hasPermission = React.useCallback(
    (permission: string): boolean => {
      if (!state.adminUser) return false;
      if (state.adminUser.role === 'super_admin') return true;
      return state.adminUser.permissions.includes(permission as any);
    },
    [state.adminUser]
  );

  /**
   * Logout admin
   */
  const logout = React.useCallback(async () => {
    await fetch('/api/admin/logout', {
      method: 'POST',
      credentials: 'include',
    });

    dispatch({ type: 'LOGOUT' });
  }, []);

  const value: AdminContextValue = {
    state,
    loadStats,
    getAuditLog,
    hasPermission,
    logout,
  };

  return (
    <AdminContext.Provider value={value}>
      {children}
    </AdminContext.Provider>
  );
}

/**
 * Use admin hook
 */
export function useAdmin(): AdminContextValue {
  const context = React.useContext(AdminContext);
  if (!context) {
    throw new Error('useAdmin must be used within an AdminProvider');
  }
  return context;
}
