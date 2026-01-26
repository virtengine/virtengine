/**
 * Support Ticket Types
 * VE-707: Support request system with encryption
 *
 * CRITICAL: Ticket content may contain sensitive information.
 * All ticket content is encrypted end-to-end.
 */

/**
 * Ticket category
 */
export type TicketCategory =
  | 'account'
  | 'identity'
  | 'billing'
  | 'provider'
  | 'marketplace'
  | 'technical'
  | 'security'
  | 'other';

/**
 * Ticket priority
 */
export type TicketPriority = 'low' | 'medium' | 'high' | 'urgent';

/**
 * Ticket status
 */
export type TicketStatus =
  | 'open'
  | 'in_progress'
  | 'waiting_customer'
  | 'waiting_internal'
  | 'resolved'
  | 'closed';

/**
 * Support ticket
 */
export interface SupportTicket {
  /**
   * Ticket ID
   */
  id: string;

  /**
   * Ticket number (human-readable)
   */
  ticketNumber: string;

  /**
   * Submitter account address
   */
  submitterAddress: string;

  /**
   * Category
   */
  category: TicketCategory;

  /**
   * Priority
   */
  priority: TicketPriority;

  /**
   * Status
   */
  status: TicketStatus;

  /**
   * Subject (not encrypted)
   */
  subject: string;

  /**
   * Encrypted ticket content
   * CRITICAL: Decrypted only by authorized support agents
   */
  encryptedContent: EncryptedTicketContent;

  /**
   * Assigned support agent ID
   */
  assignedTo?: string;

  /**
   * Related entity (provider, order, etc.)
   */
  relatedEntity?: {
    type: string;
    id: string;
  };

  /**
   * Tags
   */
  tags: string[];

  /**
   * Created timestamp
   */
  createdAt: number;

  /**
   * Updated timestamp
   */
  updatedAt: number;

  /**
   * Resolved timestamp
   */
  resolvedAt?: number;

  /**
   * First response timestamp
   */
  firstResponseAt?: number;
}

/**
 * Encrypted ticket content
 * CRITICAL: Never log decrypted content
 */
export interface EncryptedTicketContent {
  /**
   * Encrypted description
   */
  description: string;

  /**
   * Encryption IV
   */
  iv: string;

  /**
   * Ephemeral public key (for key exchange)
   */
  ephemeralPublicKey: string;

  /**
   * Algorithm used
   */
  algorithm: string;
}

/**
 * Decrypted ticket content
 * CRITICAL: Handle with care, never log
 */
export interface DecryptedTicketContent {
  /**
   * Ticket description
   */
  description: string;

  /**
   * Attachments
   */
  attachments?: TicketAttachment[];
}

/**
 * Ticket attachment
 * CRITICAL: Attachments are also encrypted
 */
export interface TicketAttachment {
  /**
   * Attachment ID
   */
  id: string;

  /**
   * Filename
   */
  filename: string;

  /**
   * MIME type
   */
  mimeType: string;

  /**
   * Size in bytes
   */
  size: number;

  /**
   * Encrypted content reference
   */
  encryptedRef: string;
}

/**
 * Ticket message
 */
export interface TicketMessage {
  /**
   * Message ID
   */
  id: string;

  /**
   * Ticket ID
   */
  ticketId: string;

  /**
   * Sender address (customer or agent)
   */
  senderAddress: string;

  /**
   * Sender name
   */
  senderName: string;

  /**
   * Whether sender is support agent
   */
  isAgent: boolean;

  /**
   * Encrypted message content
   */
  encryptedContent: string;

  /**
   * Encryption IV
   */
  iv: string;

  /**
   * Ephemeral public key
   */
  ephemeralPublicKey: string;

  /**
   * Attachments
   */
  attachments?: TicketAttachment[];

  /**
   * Created timestamp
   */
  createdAt: number;
}

/**
 * Create ticket request
 */
export interface CreateTicketRequest {
  /**
   * Category
   */
  category: TicketCategory;

  /**
   * Priority
   */
  priority: TicketPriority;

  /**
   * Subject
   */
  subject: string;

  /**
   * Description (will be encrypted)
   */
  description: string;

  /**
   * Related entity
   */
  relatedEntity?: {
    type: string;
    id: string;
  };

  /**
   * Attachments (will be encrypted)
   */
  attachments?: File[];
}

/**
 * Ticket filter
 */
export interface TicketFilter {
  /**
   * Filter by status
   */
  status?: TicketStatus | TicketStatus[];

  /**
   * Filter by category
   */
  category?: TicketCategory;

  /**
   * Filter by priority
   */
  priority?: TicketPriority;

  /**
   * Filter by assignee
   */
  assignedTo?: string;

  /**
   * Include unassigned only
   */
  unassignedOnly?: boolean;

  /**
   * Filter by submitter
   */
  submitterAddress?: string;

  /**
   * Search query
   */
  query?: string;

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
 * Support state
 */
export interface SupportState {
  /**
   * Tickets list
   */
  tickets: SupportTicket[];

  /**
   * Currently selected ticket
   */
  selectedTicket: SupportTicket | null;

  /**
   * Messages for selected ticket
   */
  messages: TicketMessage[];

  /**
   * Decrypted content (in-memory only)
   */
  decryptedContent: DecryptedTicketContent | null;

  /**
   * Support stats
   */
  stats: SupportStats;

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
 * Support stats
 */
export interface SupportStats {
  /**
   * Total open tickets
   */
  openTickets: number;

  /**
   * Tickets by status
   */
  byStatus: Record<TicketStatus, number>;

  /**
   * Tickets by category
   */
  byCategory: Record<TicketCategory, number>;

  /**
   * Tickets by priority
   */
  byPriority: Record<TicketPriority, number>;

  /**
   * Average first response time (seconds)
   */
  avgFirstResponseTime: number;

  /**
   * Average resolution time (seconds)
   */
  avgResolutionTime: number;
}
