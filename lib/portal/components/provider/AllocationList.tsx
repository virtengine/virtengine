/**
 * Allocation List Component
 * VE-705: Display provider resource allocations
 */
import * as React from 'react';

export interface Allocation {
  id: string;
  leaseId: string;
  deploymentId: string;
  owner: string;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
    gpu?: number;
  };
  status: 'active' | 'pending' | 'closing' | 'closed';
  createdAt: number;
}

export interface AllocationListProps {
  allocations: Allocation[];
  onAllocationClick?: (allocationId: string) => void;
  onAllocationClose?: (allocationId: string) => void;
  showFilters?: boolean;
  className?: string;
}

const statusColors: Record<string, string> = {
  active: '#16a34a',
  pending: '#f59e0b',
  closing: '#f59e0b',
  closed: '#6b7280',
};

export function AllocationList({
  allocations,
  onAllocationClick,
  onAllocationClose,
  showFilters = true,
  className,
}: AllocationListProps): JSX.Element {
  const [filter, setFilter] = React.useState<string>('all');

  const filteredAllocations = filter === 'all' 
    ? allocations 
    : allocations.filter(a => a.status === filter);

  return (
    <div className={className} style={{ padding: '24px' }}>
      <h3 style={{ margin: '0 0 24px', fontSize: '20px', fontWeight: 600 }}>
        Resource Allocations
      </h3>

      {showFilters && (
        <div style={{ marginBottom: '16px' }}>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            style={{ padding: '8px 12px', border: '1px solid #d1d5db', borderRadius: '4px' }}
          >
            <option value="all">All Allocations</option>
            <option value="active">Active</option>
            <option value="pending">Pending</option>
            <option value="closing">Closing</option>
            <option value="closed">Closed</option>
          </select>
        </div>
      )}

      {filteredAllocations.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>No allocations found.</p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {filteredAllocations.map((allocation) => (
            <li
              key={allocation.id}
              onClick={() => onAllocationClick?.(allocation.id)}
              style={{
                padding: '16px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                cursor: onAllocationClick ? 'pointer' : 'default',
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div>
                  <p style={{ margin: 0, fontWeight: 500 }}>Lease: {allocation.leaseId.slice(0, 16)}...</p>
                  <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                    Owner: {allocation.owner.slice(0, 16)}...
                  </p>
                  <p style={{ margin: '8px 0 0', fontSize: '14px', color: '#374151' }}>
                    CPU: {allocation.resources.cpu} | Mem: {allocation.resources.memory}GB | 
                    Storage: {allocation.resources.storage}GB
                    {allocation.resources.gpu ? ` | GPU: ${allocation.resources.gpu}` : ''}
                  </p>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '8px' }}>
                  <span style={{
                    padding: '4px 12px',
                    borderRadius: '4px',
                    fontSize: '12px',
                    fontWeight: 500,
                    color: 'white',
                    backgroundColor: statusColors[allocation.status] ?? '#6b7280',
                  }}>
                    {allocation.status}
                  </span>
                  {allocation.status === 'active' && onAllocationClose && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onAllocationClose(allocation.id);
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
                      Close
                    </button>
                  )}
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
