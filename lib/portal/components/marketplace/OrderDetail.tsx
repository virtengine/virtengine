// @ts-nocheck
/**
 * Order Detail Component
 * VE-703: Display order details
 */
import * as React from 'react';
import type { Order } from '../../types/marketplace';

export interface OrderDetailProps {
  order: Order;
  className?: string;
}

const stateColors: Record<string, string> = {
  pending: '#f59e0b',
  active: '#16a34a',
  completed: '#3b82f6',
  cancelled: '#6b7280',
  failed: '#dc2626',
};

export function OrderDetail({ order, className }: OrderDetailProps): JSX.Element {
  return (
    <div className={className}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginBottom: '24px',
      }}>
        <div>
          <h3 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>
            {order.offeringName ?? order.offeringId}
          </h3>
          <p style={{ margin: '4px 0 0', fontSize: '14px', color: '#666' }}>
            Order #{order.id.slice(0, 8)}
          </p>
        </div>
        <span
          style={{
            padding: '4px 12px',
            borderRadius: '4px',
            fontSize: '14px',
            fontWeight: 500,
            color: 'white',
            backgroundColor: stateColors[order.state] ?? '#6b7280',
          }}
        >
          {order.state}
        </span>
      </div>

      <dl style={{ margin: 0 }}>
        <div style={{ display: 'flex', padding: '8px 0', borderBottom: '1px solid #e5e7eb' }}>
          <dt style={{ width: '40%', fontWeight: 500, color: '#374151' }}>Provider</dt>
          <dd style={{ margin: 0, fontFamily: 'monospace', fontSize: '14px' }}>
            {order.providerAddress?.slice(0, 16)}...
          </dd>
        </div>
        <div style={{ display: 'flex', padding: '8px 0', borderBottom: '1px solid #e5e7eb' }}>
          <dt style={{ width: '40%', fontWeight: 500, color: '#374151' }}>Created</dt>
          <dd style={{ margin: 0 }}>{new Date(order.createdAt).toLocaleString()}</dd>
        </div>
        {order.amount && (
          <div style={{ display: 'flex', padding: '8px 0', borderBottom: '1px solid #e5e7eb' }}>
            <dt style={{ width: '40%', fontWeight: 500, color: '#374151' }}>Amount</dt>
            <dd style={{ margin: 0 }}>{order.amount} {order.denom ?? 'VE'}</dd>
          </div>
        )}
      </dl>
    </div>
  );
}
