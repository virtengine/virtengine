/**
 * Pricing Editor Component
 * VE-705: Configure provider pricing tiers
 */
import * as React from 'react';

export interface PricingTier {
  id: string;
  name: string;
  amount: number;
  denom: string;
  period: 'hour' | 'day' | 'month';
  minCommitment?: number;
}

export interface PricingEditorProps {
  tiers: PricingTier[];
  onTierAdd?: (tier: Omit<PricingTier, 'id'>) => void;
  onTierUpdate?: (tier: PricingTier) => void;
  onTierRemove?: (tierId: string) => void;
  availableDenoms?: string[];
  className?: string;
}

export function PricingEditor({
  tiers,
  onTierAdd,
  onTierUpdate,
  onTierRemove,
  availableDenoms = ['uvirt', 'uusd'],
  className,
}: PricingEditorProps): JSX.Element {
  return (
    <div className={className} style={{ padding: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h3 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>
          Pricing Tiers
        </h3>
        <button
          onClick={() => onTierAdd?.({ name: 'New Tier', amount: 0, denom: 'uvirt', period: 'hour' })}
          style={{
            padding: '8px 16px',
            fontSize: '14px',
            color: 'white',
            backgroundColor: '#3b82f6',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Add Tier
        </button>
      </div>

      {tiers.length === 0 ? (
        <p style={{ textAlign: 'center', color: '#666', padding: '24px' }}>
          No pricing tiers configured.
        </p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {tiers.map((tier) => (
            <li
              key={tier.id}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '16px',
                marginBottom: '8px',
                border: '1px solid #e5e7eb',
                borderRadius: '8px',
              }}
            >
              <div>
                <p style={{ margin: 0, fontWeight: 500 }}>{tier.name}</p>
                <p style={{ margin: '4px 0 0', fontSize: '14px', color: '#666' }}>
                  {tier.amount} {tier.denom} / {tier.period}
                </p>
              </div>
              <div style={{ display: 'flex', gap: '8px' }}>
                <button
                  onClick={() => onTierUpdate?.(tier)}
                  style={{
                    padding: '6px 12px',
                    fontSize: '12px',
                    color: '#374151',
                    backgroundColor: '#f3f4f6',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Edit
                </button>
                <button
                  onClick={() => onTierRemove?.(tier.id)}
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
                  Remove
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
