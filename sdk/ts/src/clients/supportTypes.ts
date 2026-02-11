export type SupportCategory =
  | "account"
  | "identity"
  | "billing"
  | "provider"
  | "marketplace"
  | "technical"
  | "security"
  | "other";

export type SupportPriority = "low" | "normal" | "high" | "urgent";

export type SupportStatus =
  | "unspecified"
  | "open"
  | "assigned"
  | "in_progress"
  | "waiting_customer"
  | "waiting_support"
  | "resolved"
  | "closed"
  | "archived";

export interface SupportRequestId {
  submitterAddress: string;
  sequence: number;
}

export interface RelatedEntity {
  type: string;
  id: string;
}

export interface SupportRequest {
  ticketId: string;
  submitterAddress: string;
  providerAddress: string;
  subject: string;
  description: string;
  category: SupportCategory;
  priority: SupportPriority;
  status: SupportStatus;
  createdAt: string;
  updatedAt: string;
  relatedEntity?: RelatedEntity;
}

export interface SupportResponse {
  responseId: string;
  ticketId: string;
  author: string;
  message: string;
  createdAt: string;
  isAgent: boolean;
}

export interface MsgCreateSupportRequest {
  submitterAddress: string;
  providerAddress: string;
  subject: string;
  description: string;
  category: SupportCategory;
  priority: SupportPriority;
  relatedEntity?: RelatedEntity;
  attachmentRef?: string;
}

export interface MsgUpdateSupportRequest {
  ticketId: string;
  status?: SupportStatus;
  priority?: SupportPriority;
  subject?: string;
  description?: string;
}

export interface MsgAddSupportResponse {
  ticketId: string;
  author: string;
  message: string;
  isAgent: boolean;
}

export interface MsgArchiveSupportRequest {
  ticketId: string;
  reason?: string;
}
