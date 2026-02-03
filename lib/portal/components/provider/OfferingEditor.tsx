/**
 * Offering Editor Component
 * VE-705: Edit provider service offerings
 */
import * as React from 'react';

export interface Offering {
  id: string;
  name: string;
  description?: string;
  resources: ResourceSpec;
  pricing?: PricingTier;
}

export interface ResourceSpec {
  cpu?: number;
  memory?: number;
  storage?: number;
  gpu?: number;
}

export interface PricingTier {
  amount: number;
  denom: string;
  period: 'hour' | 'day' | 'month';
}

export interface OfferingEditorProps {
  offering?: Offering;
  onSave?: (offering: Offering) => void;
  onCancel?: () => void;
  className?: string;
}

export function OfferingEditor({
  offering,
  onSave,
  onCancel,
  className,
}: OfferingEditorProps): JSX.Element {
  const [name, setName] = React.useState(offering?.name ?? '');
  const [description, setDescription] = React.useState(offering?.description ?? '');

  return (
    <div className={className} style={{ padding: '24px' }}>
      <h3 style={{ margin: '0 0 24px', fontSize: '20px', fontWeight: 600 }}>
        {offering ? 'Edit Offering' : 'New Offering'}
      </h3>

      <div style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: 500 }}>
          Name
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          style={{
            width: '100%',
            padding: '10px 12px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '14px',
          }}
        />
      </div>

      <div style={{ marginBottom: '16px' }}>
        <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', fontWeight: 500 }}>
          Description
        </label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          style={{
            width: '100%',
            padding: '10px 12px',
            border: '1px solid #d1d5db',
            borderRadius: '4px',
            fontSize: '14px',
            resize: 'vertical',
          }}
        />
      </div>

      <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
        <button
          onClick={onCancel}
          style={{
            padding: '10px 20px',
            fontSize: '14px',
            color: '#374151',
            backgroundColor: '#f3f4f6',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Cancel
        </button>
        <button
          onClick={() => onSave?.({
            id: offering?.id ?? '',
            name,
            description,
            resources: offering?.resources ?? {},
            pricing: offering?.pricing,
          })}
          style={{
            padding: '10px 20px',
            fontSize: '14px',
            color: 'white',
            backgroundColor: '#3b82f6',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Save
        </button>
      </div>
    </div>
  );
}
