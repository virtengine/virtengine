/**
 * Usage Reports Component
 * VE-705: View provider usage reports and metrics
 */
import * as React from 'react';

export interface UsageRecord {
  id: string;
  leaseId: string;
  period: { start: number; end: number };
  metrics: {
    cpuMilliSeconds: number;
    memoryByteSeconds: number;
    storageByteSeconds: number;
    gpuSeconds?: number;
  };
  amount: { value: number; denom: string };
  settled: boolean;
}

export interface UsageReportsProps {
  records: UsageRecord[];
  onRecordClick?: (recordId: string) => void;
  onExport?: () => void;
  dateRange?: { start: Date; end: Date };
  onDateRangeChange?: (range: { start: Date; end: Date }) => void;
  className?: string;
}

export function UsageReports({
  records,
  onRecordClick,
  onExport,
  dateRange,
  onDateRangeChange,
  className,
}: UsageReportsProps): JSX.Element {
  const totalAmount = records.reduce((sum, r) => sum + r.amount.value, 0);
  const settledCount = records.filter(r => r.settled).length;

  return (
    <div className={className} style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h3 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>
          Usage Reports
        </h3>
        {onExport && (
          <button
            onClick={onExport}
            style={{
              padding: '8px 16px',
              fontSize: '14px',
              color: '#374151',
              backgroundColor: '#f3f4f6',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Export
          </button>
        )}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '16px', marginBottom: '24px' }}>
        <div style={{ padding: '16px', backgroundColor: '#f3f4f6', borderRadius: '8px' }}>
          <p style={{ margin: 0, fontSize: '12px', color: '#666' }}>Total Records</p>
          <p style={{ margin: '4px 0 0', fontSize: '24px', fontWeight: 600 }}>{records.length}</p>
        </div>
        <div style={{ padding: '16px', backgroundColor: '#f3f4f6', borderRadius: '8px' }}>
          <p style={{ margin: 0, fontSize: '12px', color: '#666' }}>Settled</p>
          <p style={{ margin: '4px 0 0', fontSize: '24px', fontWeight: 600 }}>{settledCount}</p>
        </div>
        <div style={{ padding: '16px', backgroundColor: '#f3f4f6', borderRadius: '8px' }}>
          <p style={{ margin: 0, fontSize: '12px', color: '#666' }}>Total Amount</p>
          <p style={{ margin: '4px 0 0', fontSize: '24px', fontWeight: 600 }}>{totalAmount}</p>
        </div>
      </div>

      {records.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>No usage records found.</p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {records.map((record) => (
            <li
              key={record.id}
              onClick={() => onRecordClick?.(record.id)}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '16px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
                cursor: onRecordClick ? 'pointer' : 'default',
              }}
            >
              <div>
                <p style={{ margin: 0, fontWeight: 500 }}>Lease: {record.leaseId.slice(0, 16)}...</p>
                <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                  {new Date(record.period.start).toLocaleDateString()} - {new Date(record.period.end).toLocaleDateString()}
                </p>
              </div>
              <div style={{ textAlign: 'right' }}>
                <p style={{ margin: 0, fontWeight: 500 }}>
                  {record.amount.value} {record.amount.denom}
                </p>
                <span style={{
                  display: 'inline-block',
                  marginTop: '4px',
                  padding: '2px 8px',
                  borderRadius: '4px',
                  fontSize: '12px',
                  color: record.settled ? '#16a34a' : '#f59e0b',
                  backgroundColor: record.settled ? '#dcfce7' : '#fef3c7',
                }}>
                  {record.settled ? 'Settled' : 'Pending'}
                </span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
