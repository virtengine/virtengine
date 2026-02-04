/**
 * Workload Library Component
 * VE-705: Browse workload templates
 */
import * as React from 'react';
import type { WorkloadTemplate } from '../../types/hpc';

export interface WorkloadLibraryProps {
  templates: WorkloadTemplate[];
  onTemplateSelect?: (templateId: string) => void;
  showCategories?: boolean;
  showSearch?: boolean;
  className?: string;
}

export function WorkloadLibrary({
  templates,
  onTemplateSelect,
  showCategories = true,
  showSearch = true,
  className,
}: WorkloadLibraryProps): JSX.Element {
  const [search, setSearch] = React.useState('');

  const filteredTemplates = templates.filter(t =>
    t.name.toLowerCase().includes(search.toLowerCase()) ||
    t.description?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className={className}>
      {showSearch && (
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search templates..."
          style={{
            width: '100%',
            padding: '10px 12px',
            marginBottom: '16px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '14px',
          }}
        />
      )}
      
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '16px' }}>
        {filteredTemplates.map((template) => (
          <div
            key={template.id}
            onClick={() => onTemplateSelect?.(template.id)}
            style={{
              padding: '16px',
              border: '1px solid #e5e7eb',
              borderRadius: '8px',
              cursor: 'pointer',
              transition: 'border-color 0.2s',
            }}
          >
            <h4 style={{ margin: '0 0 8px', fontSize: '16px', fontWeight: 600 }}>{template.name}</h4>
            <p style={{ margin: '0 0 12px', fontSize: '14px', color: '#666', lineHeight: 1.5 }}>
              {template.description ?? 'No description'}
            </p>
            {template.category && (
              <span style={{
                padding: '2px 8px',
                fontSize: '12px',
                borderRadius: '4px',
                backgroundColor: '#f3f4f6',
                color: '#374151',
              }}>
                {template.category}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
