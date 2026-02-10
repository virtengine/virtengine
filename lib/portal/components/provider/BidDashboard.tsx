/**
 * Bid Dashboard Component
 * VE-705: View and manage provider bids
 */
import * as React from 'react';

export interface Bid {
  id: string;
  orderId: string;
  price: { amount: number; denom: string };
  status: 'pending' | 'accepted' | 'rejected' | 'closed';
  createdAt: number;
  expiresAt?: number;
}

export interface BidDashboardProps {
  bids: Bid[];
  onBidClick?: (bidId: string) => void;
  onBidWithdraw?: (bidId: string) => void;
  showFilters?: boolean;
  className?: string;
}

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  accepted: '#16a34a',
  rejected: '#dc2626',
  closed: '#6b7280',
};

export function BidDashboard({
  bids,
  onBidClick,
  onBidWithdraw,
  showFilters = true,
  className,
}: BidDashboardProps): JSX.Element {
  const [filter, setFilter] = React.useState<string>('all');

  const filteredBids = filter === 'all' ? bids : bids.filter(b => b.status === filter);

  return (
    <div className={className} style={{ padding: '24px' }}>
      <h3 style={{ margin: '0 0 24px', fontSize: '20px', fontWeight: 600 }}>
        Bid Dashboard
      </h3>

      {showFilters && (
        <div style={{ marginBottom: '16px' }}>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            style={{ padding: '8px 12px', border: '1px solid #d1d5db', borderRadius: '4px' }}
          >
            <option value="all">All Bids</option>
            <option value="pending">Pending</option>
            <option value="accepted">Accepted</option>
            <option value="rejected">Rejected</option>
            <option value="closed">Closed</option>
          </select>
        </div>
      )}

      {filteredBids.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>No bids found.</p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {filteredBids.map((bid) => (
            <li
              key={bid.id}
              onClick={() => onBidClick?.(bid.id)}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '16px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                cursor: onBidClick ? 'pointer' : 'default',
              }}
            >
              <div>
                <p style={{ margin: 0, fontWeight: 500 }}>Order: {bid.orderId.slice(0, 12)}...</p>
                <p style={{ margin: '4px 0 0', fontSize: '14px', color: '#666' }}>
                  Price: {bid.price.amount} {bid.price.denom}
                </p>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <span style={{
                  padding: '4px 12px',
                  borderRadius: '4px',
                  fontSize: '12px',
                  fontWeight: 500,
                  color: 'white',
                  backgroundColor: statusColors[bid.status] ?? '#6b7280',
                }}>
                  {bid.status}
                </span>
                {bid.status === 'pending' && onBidWithdraw && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onBidWithdraw(bid.id);
                    }}
                    style={{
                      padding: '6px 12px',
                      fontSize: '12px',
                      color: '#dc2626',
                      backgroundColor: '#fef2f2',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                    }}
                  >
                    Withdraw
                  </button>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
