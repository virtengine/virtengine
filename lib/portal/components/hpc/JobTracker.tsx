/**
 * Job Tracker Component
 * VE-705: Track HPC job status
 */
import * as React from 'react';
import type { JobStatus } from '../../types/hpc';

export interface Job {
  id: string;
  name?: string;
  status: JobStatus;
  createdAt: number;
  templateId?: string;
  providerAddress?: string;
  outputs?: any[];
}

export interface JobTrackerProps {
  jobs: Job[];
  onJobClick?: (jobId: string) => void;
  showFilters?: boolean;
  className?: string;
}

const statusColors: Record<string, string> = {
  pending: '#f59e0b',
  running: '#3b82f6',
  completed: '#16a34a',
  failed: '#dc2626',
  cancelled: '#6b7280',
};

export function JobTracker({ jobs, onJobClick, showFilters, className }: JobTrackerProps): JSX.Element {
  const [filter, setFilter] = React.useState<string>('all');

  const filteredJobs = filter === 'all' ? jobs : jobs.filter(j => j.status === filter);

  return (
    <div className={className}>
      {showFilters && (
        <div style={{ marginBottom: '16px' }}>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            style={{ padding: '8px 12px', border: '1px solid #d1d5db', borderRadius: '4px' }}
          >
            <option value="all">All Jobs</option>
            <option value="pending">Pending</option>
            <option value="running">Running</option>
            <option value="completed">Completed</option>
            <option value="failed">Failed</option>
          </select>
        </div>
      )}
      
      {filteredJobs.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>No jobs found.</p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {filteredJobs.map((job) => (
            <li
              key={job.id}
              onClick={() => onJobClick?.(job.id)}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '12px',
                borderBottom: '1px solid #e5e7eb',
                cursor: onJobClick ? 'pointer' : 'default',
              }}
            >
              <div>
                <p style={{ margin: 0, fontWeight: 500 }}>{job.name ?? job.id.slice(0, 12)}</p>
                <p style={{ margin: '4px 0 0', fontSize: '12px', color: '#666' }}>
                  {new Date(job.createdAt).toLocaleDateString()}
                </p>
              </div>
              <span style={{
                padding: '4px 12px',
                borderRadius: '4px',
                fontSize: '12px',
                fontWeight: 500,
                color: 'white',
                backgroundColor: statusColors[job.status] ?? '#6b7280',
              }}>
                {job.status}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
