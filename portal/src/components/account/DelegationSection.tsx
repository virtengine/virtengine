/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Delegation management section with MFA-gated invite flow.
 */

'use client';

import { useMemo, useState } from 'react';
import { MFAChallenge } from '@/components/mfa';
import { useMFAGate } from '@/features/mfa';

interface DelegateEntry {
  name: string;
  address: string;
  scope: string;
  limit: string;
}

interface DelegationSectionProps {
  delegates: DelegateEntry[];
}

export function DelegationSection({ delegates }: DelegationSectionProps) {
  const [address, setAddress] = useState('');
  const [scope, setScope] = useState('Marketplace + Orders');
  const [success, setSuccess] = useState(false);

  const { gateAction, challengeProps } = useMFAGate();

  const canSubmit = useMemo(() => address.trim().length > 0, [address]);

  const handleInvite = async () => {
    setSuccess(false);
    await gateAction({
      transactionType: 'delegation_change',
      actionDescription: 'Send delegation invite',
      onAuthorized: () => {
        setSuccess(true);
        setAddress('');
      },
    });
  };

  return (
    <section id="delegation" className="scroll-mt-24 rounded-xl border border-border bg-card p-6">
      <div className="mb-6">
        <h2 className="text-xl font-semibold">Delegation</h2>
        <p className="text-sm text-muted-foreground">
          Delegate limited access to teammates or partners with scoped permissions.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
        <div className="space-y-4">
          {delegates.map((delegate) => (
            <div key={delegate.address} className="rounded-lg border border-border bg-muted/30 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium">{delegate.name}</p>
                  <p className="text-xs text-muted-foreground">{delegate.address}</p>
                </div>
                <button
                  type="button"
                  className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                >
                  Edit
                </button>
              </div>
              <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                <span>Scope: {delegate.scope}</span>
                <span>Spend limit: {delegate.limit}</span>
              </div>
            </div>
          ))}
        </div>

        <div className="rounded-lg border border-border bg-muted/30 p-4">
          <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
            Add delegate
          </h3>
          <div className="mt-4 space-y-3">
            <div>
              <label className="text-sm font-medium" htmlFor="delegate-address">
                Wallet address
              </label>
              <input
                id="delegate-address"
                type="text"
                placeholder="ve1..."
                className="mt-2 w-full rounded-lg border border-border bg-background px-4 py-2 text-sm"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="delegate-scope">
                Permission scope
              </label>
              <select
                id="delegate-scope"
                className="mt-2 w-full rounded-lg border border-border bg-background px-4 py-2 text-sm"
                value={scope}
                onChange={(e) => setScope(e.target.value)}
              >
                <option>Marketplace + Orders</option>
                <option>Billing only</option>
                <option>Read-only analytics</option>
                <option>Provider management</option>
              </select>
            </div>
            {success && (
              <div className="rounded-lg border border-success/40 bg-success/10 px-3 py-2 text-xs text-success">
                Delegation invite sent. Pending MFA verification is now complete.
              </div>
            )}
            <button
              type="button"
              className="w-full rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
              onClick={handleInvite}
              disabled={!canSubmit}
            >
              Send delegation invite
            </button>
            <p className="text-xs text-muted-foreground">
              Delegation changes require MFA verification for account security.
            </p>
          </div>
        </div>
      </div>

      <MFAChallenge {...challengeProps} />
    </section>
  );
}
