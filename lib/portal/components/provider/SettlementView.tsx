/**
 * Settlement View Component
 * VE-705: View escrow settlement details
 */
import * as React from 'react';

export interface Settlement {
  id: string;
  leaseId: string;
  amount: { value: number; denom: string };
  status: 'pending' | 'processing' | 'completed' | 'failed';
  createdAt: number;
  completedAt?: number;
  txHash?: string;
}

export interface SettlementViewProps {
  settlements: Settlement[];
  onSettlementClick?: (settlementId: string) => void;
  onRetry?: (settlementId: string) => void;
  showFilters?: boolean;
  className?: string;
}

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  processing: '#3b82f6',
  completed: '#16a34a',
  failed: '#dc2626',
};

export function SettlementView({
  settlements,
  onSettlementClick,
  onRetry,
  showFilters = true,
  className,
}: SettlementViewProps): JSX.Element {
  const [filter, setFilter] = React.useState<string>('all');

  const filteredSettlements = filter === 'all' 
    ? settlements 
    : settlements.filter(s => s.status === filter);

  const totalPending = settlements
    .filter(s => s.status === 'pending' || s.status === 'processing')
    .reduce((sum, s) => sum + s.amount.value, 0);

  const totalCompleted = settlements
    .filter(s => s.status === 'completed')
    .reduce((sum, s) => sum + s.amount.value, 0);

  return (
    <div className={className} style={{ padding: '24px' }}>
      <h3 style={{ margin: '0 0 24px', fontSize: '20px', fontWeight: 600 }}>
        Settlements
      </h3>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px', marginBottom: '24px' }}>
        <div style={{ padding: '16px', backgroundColor: '#fef3c7', borderRadius: '8px' }}>
          <p style={{ margin: 0, fontSize: '12px', color: '#92400e' }}>Pending</p>
          <p style={{ margin: '4px 0 0', fontSize: '24px', fontWeight: 600, color: '#92400e' }}>{totalPending}</p>
        </div>
        <div style={{ padding: '16px', backgroundColor: '#dcfce7', borderRadius: '8px' }}>
          <p style={{ margin: 0, fontSize: '12px', color: '#166534' }}>Completed</p>
          <p style={{ margin: '4px 0 0', fontSize: '24px', fontWeight: 600, color: '#166534' }}>{totalCompleted}</p>
        </div>
      </div>

      {showFilters && (
        <div style={{ marginBottom: '16px' }}>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            style={{ padding: '8px 12px', border: '1px solid #d1d5db', borderRadius: '4px' }}
          >
            <option value="all">All Settlements</option>
            <option value="pending">Pending</option>
            <option value="processing">Processing</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
          </select>
        </div>
      )}

      {filteredSettlements.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>No settlements found.</p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {filteredSettlements.map((settlement) => (
            <li
              key={settlement.id}
              onClick={() => onSettlementClick?.(settlement.id)}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '16px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                cursor: onSettlementClick ? 'pointer' : 'default',
              }}
            >
              <div>
                <p style={{ margin: 0, fontWeight: 500 }}>Lease: {settlement.leaseId.slice(0, 16)}...</p>
                <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                  {new Date(settlement.createdAt).toLocaleString()}
                </p>
                {settlement.txHash && (
                  <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#3b82f6' }}>
                    Tx: {settlement.txHash.slice(0, 16)}...
                  </p>
                )}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                <div style={{ textAlign: 'right' }}>
                  <p style={{ margin: 0, fontWeight: 500 }}>
                    {settlement.amount.value} {settlement.amount.denom}
                  </p>
                  <span style={{
                    display: 'inline-block',
                    marginTop: '4px',
                    padding: '4px 12px',
                    borderRadius: '4px',
                    fontSize: '12px',
                    fontWeight: 500,
                    color: 'white',
                    backgroundColor: statusColors[settlement.status] ?? '#6b7280',
                  }}>
                    {settlement.status}
                  </span>
                </div>
                {settlement.status === 'failed' && onRetry && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onRetry(settlement.id);
                    }}
                    style={{
                      padding: '6px 12px',
                      fontSize: '12px',
                      color: 'white',
                      backgroundColor: '#3b82f6',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                    }}
                  >
                    Retry
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
