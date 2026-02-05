import * as React from 'react';

export type SupportCategory =
  | 'account'
  | 'identity'
  | 'billing'
  | 'provider'
  | 'marketplace'
  | 'technical'
  | 'security'
  | 'other';

export type SupportPriority = 'low' | 'normal' | 'high' | 'urgent';

export interface SupportTicketDraft {
  subject: string;
  description: string;
  category: SupportCategory;
  priority: SupportPriority;
  relatedEntityId?: string;
}

export interface SupportTicketCreateFormProps {
  value: SupportTicketDraft;
  onChange: (next: SupportTicketDraft) => void;
  onSubmit: (ticket: SupportTicketDraft) => void;
  disabled?: boolean;
  className?: string;
}

export function SupportTicketCreateForm({
  value,
  onChange,
  onSubmit,
  disabled,
  className,
}: SupportTicketCreateFormProps) {
  return (
    <form
      className={className}
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit(value);
      }}
    >
      <div>
        <label htmlFor="support-subject">Subject</label>
        <input
          id="support-subject"
          type="text"
          value={value.subject}
          onChange={(event) => onChange({ ...value, subject: event.target.value })}
          disabled={disabled}
        />
      </div>
      <div>
        <label htmlFor="support-category">Category</label>
        <select
          id="support-category"
          value={value.category}
          onChange={(event) => onChange({ ...value, category: event.target.value as SupportCategory })}
          disabled={disabled}
        >
          <option value="technical">Technical</option>
          <option value="billing">Billing</option>
          <option value="provider">Provider</option>
          <option value="marketplace">Marketplace</option>
          <option value="identity">Identity</option>
          <option value="security">Security</option>
          <option value="account">Account</option>
          <option value="other">Other</option>
        </select>
      </div>
      <div>
        <label htmlFor="support-priority">Priority</label>
        <select
          id="support-priority"
          value={value.priority}
          onChange={(event) => onChange({ ...value, priority: event.target.value as SupportPriority })}
          disabled={disabled}
        >
          <option value="low">Low</option>
          <option value="normal">Normal</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
      </div>
      <div>
        <label htmlFor="support-related">Related Entity (optional)</label>
        <input
          id="support-related"
          type="text"
          value={value.relatedEntityId ?? ''}
          onChange={(event) => onChange({ ...value, relatedEntityId: event.target.value })}
          disabled={disabled}
        />
      </div>
      <div>
        <label htmlFor="support-description">Description</label>
        <textarea
          id="support-description"
          rows={4}
          value={value.description}
          onChange={(event) => onChange({ ...value, description: event.target.value })}
          disabled={disabled}
        />
      </div>
      <button type="submit" disabled={disabled || !value.subject || !value.description}>
        Submit ticket
      </button>
    </form>
  );
}
