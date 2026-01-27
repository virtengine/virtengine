/**
 * Moderation Types
 * VE-706: Content moderation types
 */

/**
 * Moderation item type
 */
export type ModerationItemType =
  | 'provider_application'
  | 'offering'
  | 'user_report'
  | 'content_flag';

/**
 * Moderation status
 */
export type ModerationStatus =
  | 'pending'
  | 'in_review'
  | 'approved'
  | 'rejected'
  | 'escalated';

/**
 * Moderation item
 */
export interface ModerationItem {
  /**
   * Item ID
   */
  id: string;

  /**
   * Item type
   */
  type: ModerationItemType;

  /**
   * Current status
   */
  status: ModerationStatus;

  /**
   * Priority (1-5, 1 is highest)
   */
  priority: number;

  /**
   * Subject ID (provider, offering, user, etc.)
   */
  subjectId: string;

  /**
   * Subject type
   */
  subjectType: string;

  /**
   * Subject name/title
   */
  subjectName: string;

  /**
   * Submitter address
   */
  submitterAddress: string;

  /**
   * Description
   */
  description: string;

  /**
   * Evidence/attachments
   */
  evidence: ModerationEvidence[];

  /**
   * Assigned moderator ID
   */
  assignedTo?: string;

  /**
   * Created timestamp
   */
  createdAt: number;

  /**
   * Updated timestamp
   */
  updatedAt: number;

  /**
   * History of actions
   */
  history: ModerationAction[];
}

/**
 * Moderation evidence
 */
export interface ModerationEvidence {
  /**
   * Evidence type
   */
  type: 'screenshot' | 'log' | 'document' | 'link';

  /**
   * Evidence URL or reference
   */
  reference: string;

  /**
   * Description
   */
  description?: string;
}

/**
 * Moderation action
 */
export interface ModerationAction {
  /**
   * Action ID
   */
  id: string;

  /**
   * Action type
   */
  action: 'assign' | 'review' | 'approve' | 'reject' | 'escalate' | 'comment';

  /**
   * Admin user ID
   */
  adminUserId: string;

  /**
   * Admin display name
   */
  adminDisplayName: string;

  /**
   * Comment/reason
   */
  comment?: string;

  /**
   * Timestamp
   */
  timestamp: number;
}

/**
 * Moderation decision
 */
export interface ModerationDecision {
  /**
   * Decision (approve/reject)
   */
  decision: 'approve' | 'reject';

  /**
   * Reason
   */
  reason: string;

  /**
   * Additional actions to take
   */
  actions?: ModerationConsequence[];
}

/**
 * Moderation consequence
 */
export interface ModerationConsequence {
  /**
   * Consequence type
   */
  type: 'warning' | 'suspend' | 'ban' | 'remove_content';

  /**
   * Target ID
   */
  targetId: string;

  /**
   * Target type
   */
  targetType: 'user' | 'provider' | 'offering';

  /**
   * Duration (seconds, 0 = permanent)
   */
  durationSeconds?: number;

  /**
   * Reason
   */
  reason: string;
}

/**
 * Moderation queue filter
 */
export interface ModerationQueueFilter {
  /**
   * Filter by type
   */
  type?: ModerationItemType;

  /**
   * Filter by status
   */
  status?: ModerationStatus;

  /**
   * Filter by priority
   */
  priority?: number;

  /**
   * Filter by assigned moderator
   */
  assignedTo?: string;

  /**
   * Include unassigned only
   */
  unassignedOnly?: boolean;

  /**
   * Page number
   */
  page?: number;

  /**
   * Page size
   */
  pageSize?: number;
}

/**
 * Moderation state
 */
export interface ModerationState {
  /**
   * Moderation queue items
   */
  queue: ModerationItem[];

  /**
   * Currently selected item
   */
  selectedItem: ModerationItem | null;

  /**
   * Queue stats
   */
  stats: ModerationStats;

  /**
   * Whether loading
   */
  isLoading: boolean;

  /**
   * Error
   */
  error: Error | null;
}

/**
 * Moderation stats
 */
export interface ModerationStats {
  /**
   * Total pending items
   */
  pending: number;

  /**
   * Items by type
   */
  byType: Record<ModerationItemType, number>;

  /**
   * Items by priority
   */
  byPriority: Record<number, number>;

  /**
   * Average resolution time (seconds)
   */
  avgResolutionTime: number;
}
