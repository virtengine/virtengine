import { create } from 'zustand';

export type Theme = 'light' | 'dark' | 'system';
export type SidebarView = 'customer' | 'provider';

export interface UIState {
  theme: Theme;
  sidebarOpen: boolean;
  sidebarView: SidebarView;
  isWalletModalOpen: boolean;
  notifications: Notification[];
  toasts: Toast[];
}

export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  read: boolean;
  createdAt: Date;
}

export interface Toast {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message?: string;
  duration?: number;
}

export interface UIActions {
  setTheme: (theme: Theme) => void;
  toggleSidebar: () => void;
  setSidebarView: (view: SidebarView) => void;
  openWalletModal: () => void;
  closeWalletModal: () => void;
  addNotification: (notification: Omit<Notification, 'id' | 'read' | 'createdAt'>) => void;
  markNotificationRead: (id: string) => void;
  clearNotifications: () => void;
  showToast: (toast: Omit<Toast, 'id'>) => void;
  dismissToast: (id: string) => void;
}

export type UIStore = UIState & UIActions;

const initialState: UIState = {
  theme: 'system',
  sidebarOpen: true,
  sidebarView: 'customer',
  isWalletModalOpen: false,
  notifications: [],
  toasts: [],
};

let toastIdCounter = 0;
let notificationIdCounter = 0;

export const useUIStore = create<UIStore>()((set, get) => ({
  ...initialState,

  setTheme: (theme: Theme) => {
    set({ theme });
  },

  toggleSidebar: () => {
    set((state) => ({ sidebarOpen: !state.sidebarOpen }));
  },

  setSidebarView: (view: SidebarView) => {
    set({ sidebarView: view });
  },

  openWalletModal: () => {
    set({ isWalletModalOpen: true });
  },

  closeWalletModal: () => {
    set({ isWalletModalOpen: false });
  },

  addNotification: (notification) => {
    const id = `notification-${++notificationIdCounter}`;
    set((state) => ({
      notifications: [
        {
          ...notification,
          id,
          read: false,
          createdAt: new Date(),
        },
        ...state.notifications,
      ],
    }));
  },

  markNotificationRead: (id: string) => {
    set((state) => ({
      notifications: state.notifications.map((n) => (n.id === id ? { ...n, read: true } : n)),
    }));
  },

  clearNotifications: () => {
    set({ notifications: [] });
  },

  showToast: (toast) => {
    const id = `toast-${++toastIdCounter}`;
    set((state) => ({
      toasts: [...state.toasts, { ...toast, id }],
    }));

    // Auto-dismiss after duration
    const duration = toast.duration ?? 5000;
    if (duration > 0) {
      setTimeout(() => {
        get().dismissToast(id);
      }, duration);
    }
  },

  dismissToast: (id: string) => {
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    }));
  },
}));

// Selectors
export const selectUnreadNotificationCount = (state: UIStore) =>
  state.notifications.filter((n) => !n.read).length;
