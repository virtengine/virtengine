import * as React from 'react';

export interface SupportResponseItem {
  id: string;
  author: string;
  message: string;
  createdAt: Date;
  isAgent?: boolean;
}

export interface SupportTicketDetail {
  id: string;
  number: string;
  subject: string;
  description: string;
  status: string;
  priority: string;
  createdAt: Date;
  responses: SupportResponseItem[];
}

export interface SupportTicketDetailProps {
  ticket: SupportTicketDetail;
  onRespond?: (message: string) => void;
  className?: string;
}

export function SupportTicketDetail({ ticket, onRespond, className }: SupportTicketDetailProps) {
  const [message, setMessage] = React.useState('');

  return (
    <section className={className}>
      <header>
        <h2>{ticket.subject}</h2>
        <div>
          <span>{ticket.number}</span> · <span>{ticket.priority}</span> · <span>{ticket.status}</span>
        </div>
      </header>

      <p>{ticket.description}</p>

      <div>
        {ticket.responses.map((response) => (
          <article key={response.id}>
            <strong>{response.author}</strong>
            <p>{response.message}</p>
          </article>
        ))}
      </div>

      {onRespond && (
        <form
          onSubmit={(event) => {
            event.preventDefault();
            if (!message.trim()) return;
            onRespond(message.trim());
            setMessage('');
          }}
        >
          <textarea
            rows={3}
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            placeholder="Add a response"
          />
          <button type="submit" disabled={!message.trim()}>
            Send response
          </button>
        </form>
      )}
    </section>
  );
}
