/**
 * Admin Library
 * VE-706, VE-707: Admin & moderation portal with support ticket system
 *
 * CRITICAL: Admin actions are audited. Never log sensitive data.
 */

// Types
export * from './types/admin';
export * from './types/moderation';
export * from './types/support';

// Hooks
export { AdminProvider, useAdmin } from './hooks/useAdmin';
export { ModerationProvider, useModeration } from './hooks/useModeration';
export { SupportProvider, useSupport } from './hooks/useSupport';

// Components
export { AdminDashboard } from './components/AdminDashboard';
export { ModerationQueue } from './components/ModerationQueue';
export { SupportTicketList } from './components/SupportTicketList';
export { SupportTicketDetail } from './components/SupportTicketDetail';
export { AuditLog } from './components/AuditLog';
