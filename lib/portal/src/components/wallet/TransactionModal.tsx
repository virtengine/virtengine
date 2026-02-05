'use client';

export interface TransactionFee {
  amount: string;
  denom: string;
}

export interface TransactionPreview {
  type: string;
  description: string;
  from: string;
  to?: string;
  amount?: string;
  denom?: string;
  fee: TransactionFee;
  memo?: string;
  messages?: Array<{ type: string; value: unknown }>;
}

export interface TransactionModalProps {
  isOpen: boolean;
  onClose: () => void;
  preview: TransactionPreview;
  onConfirm: () => void;
  isLoading?: boolean;
}

function truncateAddress(address: string): string {
  if (address.length <= 20) return address;
  return `${address.slice(0, 12)}...${address.slice(-6)}`;
}

function formatAmount(amount: string, denom: string): string {
  const num = parseFloat(amount);
  if (isNaN(num)) return `${amount} ${denom}`;
  return `${num.toLocaleString(undefined, { maximumFractionDigits: 6 })} ${denom.toUpperCase()}`;
}

export function TransactionModal({
  isOpen,
  onClose,
  preview,
  onConfirm,
  isLoading = false,
}: TransactionModalProps) {
  if (!isOpen) return null;

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget && !isLoading) {
      onClose();
    }
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="tx-modal-title"
      onClick={handleBackdropClick}
      style={{
        position: 'fixed',
        inset: 0,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        zIndex: 100,
        padding: 16,
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: 440,
          backgroundColor: '#fff',
          borderRadius: 12,
          boxShadow: '0 20px 40px rgba(0, 0, 0, 0.2)',
          overflow: 'hidden',
        }}
      >
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '16px 20px',
            borderBottom: '1px solid #e5e7eb',
            backgroundColor: '#f9fafb',
          }}
        >
          <h2
            id="tx-modal-title"
            style={{
              margin: 0,
              fontSize: 18,
              fontWeight: 600,
              color: '#111827',
            }}
          >
            Confirm Transaction
          </h2>
          <button
            type="button"
            onClick={onClose}
            disabled={isLoading}
            aria-label="Close"
            style={{
              padding: 4,
              border: 'none',
              background: 'transparent',
              fontSize: 20,
              cursor: isLoading ? 'not-allowed' : 'pointer',
              color: '#6b7280',
              opacity: isLoading ? 0.5 : 1,
            }}
          >
            ✕
          </button>
        </div>

        {/* Body */}
        <div style={{ padding: 20 }}>
          {/* Transaction type */}
          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Transaction Type</div>
            <div
              style={{
                display: 'inline-block',
                padding: '4px 10px',
                borderRadius: 6,
                backgroundColor: '#e0e7ff',
                color: '#4338ca',
                fontSize: 13,
                fontWeight: 500,
              }}
            >
              {preview.type}
            </div>
          </div>

          {/* Description */}
          {preview.description && (
            <div style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Description</div>
              <div style={{ fontSize: 14, color: '#374151' }}>{preview.description}</div>
            </div>
          )}

          {/* Details grid */}
          <div
            style={{
              backgroundColor: '#f9fafb',
              borderRadius: 8,
              padding: 16,
              marginBottom: 16,
            }}
          >
            {/* From */}
            <div style={{ marginBottom: 12 }}>
              <div style={{ fontSize: 11, color: '#6b7280', textTransform: 'uppercase', marginBottom: 4 }}>
                From
              </div>
              <div style={{ fontFamily: 'monospace', fontSize: 13, color: '#111827' }}>
                {truncateAddress(preview.from)}
              </div>
            </div>

            {/* To */}
            {preview.to && (
              <div style={{ marginBottom: 12 }}>
                <div style={{ fontSize: 11, color: '#6b7280', textTransform: 'uppercase', marginBottom: 4 }}>
                  To
                </div>
                <div style={{ fontFamily: 'monospace', fontSize: 13, color: '#111827' }}>
                  {truncateAddress(preview.to)}
                </div>
              </div>
            )}

            {/* Amount */}
            {preview.amount && preview.denom && (
              <div style={{ marginBottom: 12 }}>
                <div style={{ fontSize: 11, color: '#6b7280', textTransform: 'uppercase', marginBottom: 4 }}>
                  Amount
                </div>
                <div style={{ fontSize: 16, fontWeight: 600, color: '#111827' }}>
                  {formatAmount(preview.amount, preview.denom)}
                </div>
              </div>
            )}

            {/* Memo */}
            {preview.memo && (
              <div>
                <div style={{ fontSize: 11, color: '#6b7280', textTransform: 'uppercase', marginBottom: 4 }}>
                  Memo
                </div>
                <div
                  style={{
                    fontSize: 13,
                    color: '#374151',
                    wordBreak: 'break-word',
                  }}
                >
                  {preview.memo}
                </div>
              </div>
            )}
          </div>

          {/* Fee section */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '12px 16px',
              backgroundColor: '#fef3c7',
              borderRadius: 8,
              marginBottom: 16,
            }}
          >
            <span style={{ fontSize: 13, color: '#92400e', fontWeight: 500 }}>
              Network Fee
            </span>
            <span style={{ fontSize: 14, fontWeight: 600, color: '#78350f' }}>
              {formatAmount(preview.fee.amount, preview.fee.denom)}
            </span>
          </div>

          {/* Message count (if multiple) */}
          {preview.messages && preview.messages.length > 1 && (
            <div
              style={{
                padding: '8px 12px',
                backgroundColor: '#f0fdf4',
                borderRadius: 6,
                marginBottom: 16,
                fontSize: 13,
                color: '#166534',
              }}
            >
              This transaction contains {preview.messages.length} messages
            </div>
          )}
        </div>

        {/* Footer */}
        <div
          style={{
            display: 'flex',
            gap: 12,
            padding: '16px 20px',
            borderTop: '1px solid #e5e7eb',
            backgroundColor: '#f9fafb',
          }}
        >
          <button
            type="button"
            onClick={onClose}
            disabled={isLoading}
            style={{
              flex: 1,
              padding: '12px 16px',
              borderRadius: 8,
              border: '1px solid #d1d5db',
              backgroundColor: '#fff',
              color: '#374151',
              fontSize: 14,
              fontWeight: 500,
              cursor: isLoading ? 'not-allowed' : 'pointer',
              opacity: isLoading ? 0.5 : 1,
            }}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={onConfirm}
            disabled={isLoading}
            style={{
              flex: 1,
              padding: '12px 16px',
              borderRadius: 8,
              border: 'none',
              backgroundColor: isLoading ? '#9ca3af' : '#3b82f6',
              color: '#fff',
              fontSize: 14,
              fontWeight: 500,
              cursor: isLoading ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 8,
            }}
          >
            {isLoading ? (
              <>
                <span
                  style={{
                    display: 'inline-block',
                    width: 16,
                    height: 16,
                    border: '2px solid rgba(255,255,255,0.3)',
                    borderTopColor: '#fff',
                    borderRadius: '50%',
                    animation: 'tx-spinner 0.8s linear infinite',
                  }}
                />
                Signing...
              </>
            ) : (
              'Confirm & Sign'
            )}
          </button>
        </div>

        {/* Spinner animation */}
        {isLoading && (
          <style
            dangerouslySetInnerHTML={{
              __html: `
                @keyframes tx-spinner {
                  from { transform: rotate(0deg); }
                  to { transform: rotate(360deg); }
                }
              `,
            }}
          />
        )}
      </div>
    </div>
  );
}
