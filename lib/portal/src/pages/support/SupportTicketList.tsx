import * as React from 'react';

export interface SupportTicketListItem {
  id: string;
  number: string;
  subject: string;
  status: string;
  priority: string;
  updatedAt: Date;
}

export interface SupportTicketListProps {
  tickets: SupportTicketListItem[];
  onSelect?: (ticketId: string) => void;
  className?: string;
}

export function SupportTicketList({ tickets, onSelect, className }: SupportTicketListProps) {
  return (
    <div className={className}>
      {tickets.map((ticket) => (
        <button
          key={ticket.id}
          type="button"
          onClick={() => onSelect?.(ticket.id)}
          style={{ display: 'block', width: '100%', textAlign: 'left' }}
        >
          <strong>{ticket.subject}</strong>
          <div>
            <span>{ticket.number}</span> · <span>{ticket.priority}</span> · <span>{ticket.status}</span>
          </div>
        </button>
      ))}
    </div>
  );
}
