// @ts-nocheck
/**
 * Order Timeline Component
 * VE-703: Display order event timeline
 */
import * as React from 'react';
import type { OrderEvent } from '../../types/marketplace';

export interface OrderTimelineProps {
  events: OrderEvent[];
  className?: string;
}

const eventColors: Record<string, string> = {
  created: '#3b82f6',
  matched: '#8b5cf6',
  active: '#16a34a',
  completed: '#16a34a',
  cancelled: '#6b7280',
  failed: '#dc2626',
};

export function OrderTimeline({ events, className }: OrderTimelineProps): JSX.Element {
  if (events.length === 0) {
    return (
      <div className={className}>
        <p style={{ margin: 0, fontSize: '14px', color: '#666' }}>
          No events recorded yet.
        </p>
      </div>
    );
  }

  return (
    <div className={className}>
      <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
        {events.map((event, index) => (
          <li
            key={index}
            style={{
              display: 'flex',
              gap: '12px',
              paddingBottom: index === events.length - 1 ? 0 : '16px',
              position: 'relative',
            }}
          >
            {/* Timeline line */}
            {index !== events.length - 1 && (
              <div
                style={{
                  position: 'absolute',
                  left: '11px',
                  top: '24px',
                  bottom: 0,
                  width: '2px',
                  backgroundColor: '#e5e7eb',
                }}
              />
            )}
            {/* Timeline dot */}
            <div
              style={{
                width: '24px',
                height: '24px',
                borderRadius: '50%',
                backgroundColor: eventColors[event.type] ?? '#6b7280',
                flexShrink: 0,
              }}
            />
            {/* Event content */}
            <div style={{ flex: 1 }}>
              <p style={{ margin: 0, fontWeight: 500 }}>{event.type}</p>
              {event.message && (
                <p style={{ margin: '4px 0 0', fontSize: '14px', color: '#666' }}>
                  {event.message}
                </p>
              )}
              <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#9ca3af' }}>
                {new Date(event.timestamp).toLocaleString()}
              </p>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
