/**
 * Admin Types
 * VE-706: Admin portal types
 */

/**
 * Admin role
 */
export type AdminRole = 'super_admin' | 'admin' | 'moderator' | 'support';

/**
 * Admin user
 */
export interface AdminUser {
  /**
   * Admin user ID
   */
  id: string;

  /**
   * Account address
   */
  accountAddress: string;

  /**
   * Display name
   */
  displayName: string;

  /**
   * Admin role
   */
  role: AdminRole;

  /**
   * Permissions granted
   */
  permissions: AdminPermission[];

  /**
   * Whether MFA is required
   */
  mfaRequired: boolean;

  /**
   * Whether MFA is enabled
   */
  mfaEnabled: boolean;

  /**
   * Last login timestamp
   */
  lastLoginAt: number;

  /**
   * Created timestamp
   */
  createdAt: number;
}

/**
 * Admin permission
 */
export type AdminPermission =
  | 'users.read'
  | 'users.write'
  | 'users.suspend'
  | 'providers.read'
  | 'providers.write'
  | 'providers.approve'
  | 'providers.suspend'
  | 'offerings.read'
  | 'offerings.write'
  | 'offerings.approve'
  | 'offerings.remove'
  | 'moderation.read'
  | 'moderation.write'
  | 'support.read'
  | 'support.write'
  | 'audit.read'
  | 'config.read'
  | 'config.write';

/**
 * Admin state
 */
export interface AdminState {
  /**
   * Current admin user
   */
  adminUser: AdminUser | null;

  /**
   * Whether admin is authenticated
   */
  isAuthenticated: boolean;

  /**
   * Whether admin session requires MFA
   */
  requiresMFA: boolean;

  /**
   * Whether loading
   */
  isLoading: boolean;

  /**
   * Error
   */
  error: Error | null;

  /**
   * Dashboard stats
   */
  stats: DashboardStats | null;
}

/**
 * Dashboard stats
 */
export interface DashboardStats {
  /**
   * Total users
   */
  totalUsers: number;

  /**
   * Active users (last 24h)
   */
  activeUsers24h: number;

  /**
   * Total providers
   */
  totalProviders: number;

  /**
   * Active providers
   */
  activeProviders: number;

  /**
   * Pending provider applications
   */
  pendingProviders: number;

  /**
   * Total offerings
   */
  totalOfferings: number;

  /**
   * Active offerings
   */
  activeOfferings: number;

  /**
   * Pending moderation items
   */
  pendingModeration: number;

  /**
   * Open support tickets
   */
  openTickets: number;

  /**
   * Total transactions (24h)
   */
  transactions24h: number;

  /**
   * Transaction volume (24h)
   */
  volume24h: string;
}

/**
 * Audit log entry
 */
export interface AuditLogEntry {
  /**
   * Entry ID
   */
  id: string;

  /**
   * Admin user ID who performed action
   */
  adminUserId: string;

  /**
   * Admin display name
   */
  adminDisplayName: string;

  /**
   * Action performed
   */
  action: string;

  /**
   * Resource type affected
   */
  resourceType: string;

  /**
   * Resource ID affected
   */
  resourceId: string;

  /**
   * Details (non-sensitive)
   */
  details: Record<string, unknown>;

  /**
   * IP address (masked)
   */
  ipAddress: string;

  /**
   * Timestamp
   */
  timestamp: number;
}

/**
 * Audit log filter
 */
export interface AuditLogFilter {
  /**
   * Filter by admin user
   */
  adminUserId?: string;

  /**
   * Filter by action
   */
  action?: string;

  /**
   * Filter by resource type
   */
  resourceType?: string;

  /**
   * Filter by resource ID
   */
  resourceId?: string;

  /**
   * Start timestamp
   */
  startTime?: number;

  /**
   * End timestamp
   */
  endTime?: number;

  /**
   * Page number
   */
  page?: number;

  /**
   * Page size
   */
  pageSize?: number;
}
