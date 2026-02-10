// @ts-nocheck
/**
 * Offering Detail Component
 * VE-703: Display detailed offering information
 */
import * as React from 'react';
import type { Offering } from '../../types/marketplace';

export interface OfferingDetailProps {
  offering: Offering;
  onCheckout?: () => void;
  onBack?: () => void;
  className?: string;
}

export function OfferingDetail({
  offering,
  onCheckout,
  onBack,
  className,
}: OfferingDetailProps): JSX.Element {
  return (
    <div className={className}>
      <div style={{ marginBottom: '16px' }}>
        {onBack && (
          <button
            onClick={onBack}
            style={{
              padding: '4px 0',
              fontSize: '14px',
              color: '#3b82f6',
              backgroundColor: 'transparent',
              border: 'none',
              cursor: 'pointer',
              marginBottom: '12px',
            }}
          >
            ← Back to listings
          </button>
        )}
        <h2 style={{ margin: 0, fontSize: '24px', fontWeight: 600 }}>{offering.name}</h2>
        <p style={{ margin: '4px 0 0', fontSize: '14px', color: '#666' }}>
          by {offering.providerName ?? offering.providerAddress?.slice(0, 12)}...
        </p>
      </div>

      <div style={{ marginBottom: '24px' }}>
        <p style={{ margin: 0, lineHeight: 1.6 }}>{offering.description}</p>
      </div>

      <div style={{
        padding: '16px',
        borderRadius: '8px',
        backgroundColor: '#f3f4f6',
        marginBottom: '24px',
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <p style={{ margin: 0, fontSize: '12px', color: '#666' }}>Price</p>
            <p style={{ margin: '4px 0 0', fontSize: '24px', fontWeight: 600 }}>
              {offering.price} {offering.priceDenom ?? 'VE'}
            </p>
          </div>
          {onCheckout && (
            <button
              onClick={onCheckout}
              style={{
                padding: '12px 24px',
                fontSize: '16px',
                fontWeight: 600,
                color: 'white',
                backgroundColor: '#3b82f6',
                border: 'none',
                borderRadius: '8px',
                cursor: 'pointer',
              }}
            >
              Order Now
            </button>
          )}
        </div>
      </div>

      {offering.specs && (
        <div>
          <h3 style={{ margin: '0 0 12px', fontSize: '18px', fontWeight: 600 }}>Specifications</h3>
          <dl style={{ margin: 0 }}>
            {Object.entries(offering.specs).map(([key, value]) => (
              <div key={key} style={{ display: 'flex', padding: '8px 0', borderBottom: '1px solid #e5e7eb' }}>
                <dt style={{ width: '40%', fontWeight: 500, color: '#374151' }}>{key}</dt>
                <dd style={{ margin: 0, color: '#666' }}>{String(value)}</dd>
              </div>
            ))}
          </dl>
        </div>
      )}
    </div>
  );
}
