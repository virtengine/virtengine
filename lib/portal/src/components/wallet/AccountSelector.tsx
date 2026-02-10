'use client';

import { useState, useCallback } from 'react';
import type { WalletAccount } from '../../wallet/types';

export interface AccountSelectorProps {
  accounts: WalletAccount[];
  activeIndex: number;
  onSelect: (index: number) => void;
  className?: string;
}

function truncateAddress(address: string): string {
  if (address.length <= 16) return address;
  return `${address.slice(0, 10)}...${address.slice(-4)}`;
}

export function AccountSelector({
  accounts,
  activeIndex,
  onSelect,
  className,
}: AccountSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const activeAccount = accounts[activeIndex];

  const handleCopyAddress = useCallback(async (address: string, index: number, event: React.MouseEvent) => {
    event.stopPropagation();
    try {
      await navigator.clipboard.writeText(address);
      setCopiedIndex(index);
      setTimeout(() => setCopiedIndex(null), 2000);
    } catch {
      // Clipboard API not available
    }
  }, []);

  const handleSelect = useCallback((index: number) => {
    onSelect(index);
    setIsOpen(false);
  }, [onSelect]);

  if (accounts.length === 0) {
    return null;
  }

  return (
    <div
      className={className}
      style={{ position: 'relative', display: 'inline-block' }}
    >
      {/* Trigger button */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '8px 12px',
          borderRadius: 8,
          border: '1px solid #d1d5db',
          backgroundColor: '#fff',
          cursor: 'pointer',
          minWidth: 180,
        }}
      >
        <span
          style={{
            width: 8,
            height: 8,
            borderRadius: '50%',
            backgroundColor: '#22c55e',
          }}
          aria-hidden="true"
        />
        <span style={{ fontFamily: 'monospace', fontSize: 13, flex: 1, textAlign: 'left' }}>
          {activeAccount ? truncateAddress(activeAccount.address) : 'Select account'}
        </span>
        <span
          style={{
            fontSize: 10,
            color: '#6b7280',
            transform: isOpen ? 'rotate(180deg)' : 'rotate(0deg)',
            transition: 'transform 0.15s',
          }}
          aria-hidden="true"
        >
          ▼
        </span>
      </button>

      {/* Dropdown */}
      {isOpen && (
        <div
          role="listbox"
          aria-label="Select wallet account"
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            right: 0,
            marginTop: 4,
            backgroundColor: '#fff',
            borderRadius: 8,
            border: '1px solid #e5e7eb',
            boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
            zIndex: 50,
            maxHeight: 240,
            overflowY: 'auto',
          }}
        >
          {accounts.map((account, index) => {
            const isActive = index === activeIndex;
            const isCopied = copiedIndex === index;

            return (
              <div
                key={account.address}
                role="option"
                aria-selected={isActive}
                onClick={() => handleSelect(index)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '10px 12px',
                  cursor: 'pointer',
                  backgroundColor: isActive ? '#f0f9ff' : 'transparent',
                  borderLeft: isActive ? '3px solid #3b82f6' : '3px solid transparent',
                }}
                onMouseEnter={(e) => {
                  if (!isActive) {
                    e.currentTarget.style.backgroundColor = '#f9fafb';
                  }
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = isActive ? '#f0f9ff' : 'transparent';
                }}
              >
                {/* Active indicator */}
                <span
                  style={{
                    width: 6,
                    height: 6,
                    borderRadius: '50%',
                    backgroundColor: isActive ? '#22c55e' : '#d1d5db',
                  }}
                  aria-hidden="true"
                />

                {/* Account info */}
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 2 }}>
                    Account {index + 1}
                  </div>
                  <div style={{ fontFamily: 'monospace', fontSize: 13, color: '#111827' }}>
                    {truncateAddress(account.address)}
                  </div>
                </div>

                {/* Copy button */}
                <button
                  type="button"
                  onClick={(e) => handleCopyAddress(account.address, index, e)}
                  title={isCopied ? 'Copied!' : 'Copy address'}
                  style={{
                    padding: 4,
                    borderRadius: 4,
                    border: 'none',
                    backgroundColor: isCopied ? '#dcfce7' : '#f3f4f6',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 12,
                    color: isCopied ? '#16a34a' : '#6b7280',
                  }}
                  aria-label={isCopied ? 'Address copied' : 'Copy address to clipboard'}
                >
                  {isCopied ? '✓' : '⧉'}
                </button>
              </div>
            );
          })}
        </div>
      )}

      {/* Click outside to close */}
      {isOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 40,
          }}
          onClick={() => setIsOpen(false)}
          aria-hidden="true"
        />
      )}
    </div>
  );
}
