'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { useWallet, useIdentity } from '@/lib/portal-adapter';
import { IdentityRequirements } from '@/components/identity';
import { formatCurrency, formatTokenAmount, generateId, truncateAddress } from '@/lib/utils';
import { formatPriceUSD, useOfferingStore } from '@/stores/offeringStore';

const SIGNING_DELAY_MS = 1200;

type OrderStep = 'configure' | 'review' | 'signing' | 'confirmed';

type ResourceLineItem = {
  key: string;
  resourceType: string;
  unit: string;
  unitPrice: number;
  usdReference?: string;
};

function formatUve(amount: number): string {
  return `${formatTokenAmount(amount, 6, 2)} UVE`;
}

function buildTxHash(): string {
  const random = Math.random().toString(16).slice(2).padEnd(16, '0');
  return `0x${random.slice(0, 16)}`;
}

export default function OrderCreateClient() {
  const params = useParams();
  const router = useRouter();
  const provider = params.provider as string;
  const sequence = Number(params.sequence);

  const { status, accounts, activeAccountIndex } = useWallet();
  const account = accounts[activeAccountIndex];
  const { actions: identityActions } = useIdentity();
  const gatingError = identityActions.checkRequirements('place_order');

  const {
    selectedOffering: offering,
    isLoadingDetail,
    error,
    fetchOffering,
    fetchProvider,
    clearError,
  } = useOfferingStore();

  const [step, setStep] = useState<OrderStep>('configure');
  const [quantities, setQuantities] = useState<Record<string, number>>({});
  const [region, setRegion] = useState('');
  const [orderName, setOrderName] = useState('');
  const [notes, setNotes] = useState('');
  const [txHash, setTxHash] = useState<string | null>(null);
  const [orderId, setOrderId] = useState<string | null>(null);

  useEffect(() => {
    if (provider && Number.isFinite(sequence)) {
      void fetchOffering(provider, sequence);
    }
  }, [fetchOffering, provider, sequence]);

  useEffect(() => {
    if (offering?.id.providerAddress) {
      void fetchProvider(offering.id.providerAddress);
    }
  }, [fetchProvider, offering]);

  useEffect(() => {
    const prices = offering?.prices;
    if (!offering || !prices) return;
    setQuantities((prev) => {
      const next = { ...prev };
      prices.forEach((price, index) => {
        const key = `${price.resourceType}-${price.unit}`;
        if (next[key] === undefined) {
          next[key] = index === 0 ? 1 : 0;
        }
      });
      return next;
    });

    if (!region && offering.regions?.length) {
      setRegion(offering.regions[0]);
    }
  }, [offering, region]);

  const lineItems = useMemo<ResourceLineItem[]>(() => {
    if (!offering?.prices) return [];
    return offering.prices.map((price) => ({
      key: `${price.resourceType}-${price.unit}`,
      resourceType: price.resourceType,
      unit: price.unit,
      unitPrice: Number.parseInt(price.price.amount, 10),
      usdReference: price.usdReference,
    }));
  }, [offering]);

  const totals = useMemo(() => {
    const itemTotals = lineItems.map((item) => {
      const quantity = quantities[item.key] ?? 0;
      const totalUve = Math.round(item.unitPrice * quantity);
      const totalUsd = item.usdReference ? quantity * Number.parseFloat(item.usdReference) : null;
      return {
        ...item,
        quantity,
        totalUve,
        totalUsd,
      };
    });

    const totalUve = itemTotals.reduce((sum, item) => sum + item.totalUve, 0);
    const totalUsd = itemTotals.reduce((sum, item) => sum + (item.totalUsd ?? 0), 0);
    const hasUsd = itemTotals.every((item) => item.usdReference);

    return {
      itemTotals,
      totalUve,
      totalUsd,
      hasUsd,
    };
  }, [lineItems, quantities]);

  const escrowAccount = offering
    ? `escrow-${offering.id.providerAddress.slice(0, 6)}-${offering.id.sequence}`
    : 'escrow-unknown';

  const resourceSelections = totals.itemTotals
    .filter((item) => item.quantity > 0)
    .map((item) => ({
      resourceType: item.resourceType,
      unit: item.unit,
      quantity: item.quantity,
    }));

  const msgPreview = offering
    ? {
        owner: account?.address ?? 'connect-wallet',
        offeringId: `${offering.id.providerAddress}/${offering.id.sequence}`,
        resources: resourceSelections,
        deposit: {
          denom: 'uve',
          amount: totals.totalUve.toString(),
        },
      }
    : null;

  const canProceed = totals.totalUve > 0 && resourceSelections.length > 0;

  const handleSign = async () => {
    setStep('signing');
    await new Promise((resolve) => setTimeout(resolve, SIGNING_DELAY_MS));
    setTxHash(buildTxHash());
    setOrderId(generateId('order'));
    setStep('confirmed');
  };

  if (isLoadingDetail) {
    return (
      <div className="container py-8">
        <div className="rounded-lg border border-border bg-card p-6">Loading order flow...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container py-8">
        <div className="rounded-lg border border-red-200 bg-red-50 p-6 text-sm text-red-600">
          <p className="font-medium">Unable to load offering</p>
          <p className="mt-1">{error}</p>
          <button
            type="button"
            onClick={() => {
              clearError();
              router.push('/marketplace');
            }}
            className="mt-4 rounded-lg bg-red-600 px-4 py-2 text-white"
          >
            Back to Marketplace
          </button>
        </div>
      </div>
    );
  }

  if (!offering) {
    return null;
  }

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link
          href={`/marketplace/${provider}/${sequence}`}
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to Offering
        </Link>
      </div>

      <div className="mb-6">
        <h1 className="text-3xl font-bold">Create Order</h1>
        <p className="mt-1 text-muted-foreground">
          Fixed pricing is active. Bidding is deferred to Phase 3.
        </p>
      </div>

      <div className="grid gap-8 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Offering</p>
                <h2 className="text-2xl font-semibold">{offering.name}</h2>
                <p className="mt-1 text-sm text-muted-foreground">{offering.description}</p>
              </div>
              <div className="text-right">
                <p className="text-sm text-muted-foreground">Provider</p>
                <p className="font-medium">{truncateAddress(offering.id.providerAddress)}</p>
              </div>
            </div>
          </div>

          {step === 'configure' && gatingError && (
            <IdentityRequirements
              action="place_order"
              onStartVerification={() => router.push('/verify')}
            />
          )}

          {step === 'configure' && !gatingError && (
            <>
              <div className="rounded-lg border border-border bg-card p-6">
                <h2 className="text-lg font-semibold">1. Order Configuration</h2>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <div>
                    <label htmlFor="order-name" className="text-sm font-medium">
                      Order name
                    </label>
                    <input
                      id="order-name"
                      type="text"
                      value={orderName}
                      onChange={(event) => setOrderName(event.target.value)}
                      placeholder={`${offering.name} deployment`}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    />
                  </div>
                  <div>
                    <label htmlFor="order-region" className="text-sm font-medium">
                      Region
                    </label>
                    <select
                      id="order-region"
                      value={region}
                      onChange={(event) => setRegion(event.target.value)}
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    >
                      {offering.regions?.map((item) => (
                        <option key={item} value={item}>
                          {item}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="sm:col-span-2">
                    <label htmlFor="order-notes" className="text-sm font-medium">
                      Notes for provider (optional)
                    </label>
                    <textarea
                      id="order-notes"
                      rows={2}
                      value={notes}
                      onChange={(event) => setNotes(event.target.value)}
                      placeholder="Special scheduling or deployment notes..."
                      className="mt-1 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
                    />
                  </div>
                </div>
              </div>

              <div className="rounded-lg border border-border bg-card p-6">
                <h2 className="text-lg font-semibold">2. Resource Selection</h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  Enter fixed units for each priced resource. Total is calculated as sum(units ×
                  unit price).
                </p>

                <div className="mt-4 overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-border text-left">
                        <th className="pb-3 font-medium">Resource</th>
                        <th className="pb-3 font-medium">Unit</th>
                        <th className="pb-3 text-right font-medium">Unit Price</th>
                        <th className="pb-3 text-right font-medium">USD / Unit</th>
                        <th className="pb-3 text-right font-medium">Units</th>
                        <th className="pb-3 text-right font-medium">Line Total</th>
                      </tr>
                    </thead>
                    <tbody>
                      {totals.itemTotals.map((item) => (
                        <tr key={item.key} className="border-b border-border last:border-0">
                          <td className="py-3 capitalize">{item.resourceType}</td>
                          <td className="py-3 text-muted-foreground">{item.unit}</td>
                          <td className="py-3 text-right font-mono">{formatUve(item.unitPrice)}</td>
                          <td className="py-3 text-right text-muted-foreground">
                            {item.usdReference ? formatPriceUSD(item.usdReference) : 'Unavailable'}
                          </td>
                          <td className="py-3 text-right">
                            <input
                              type="number"
                              min={0}
                              step={1}
                              value={item.quantity}
                              onChange={(event) => {
                                const value = Number.parseFloat(event.target.value || '0');
                                setQuantities((prev) => ({
                                  ...prev,
                                  [item.key]: Number.isNaN(value) ? 0 : value,
                                }));
                              }}
                              className="w-24 rounded-lg border border-border bg-background px-2 py-1 text-right text-sm"
                            />
                          </td>
                          <td className="py-3 text-right font-mono">{formatUve(item.totalUve)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {!canProceed && (
                  <p className="mt-3 text-sm text-muted-foreground">
                    Select at least one resource unit to proceed.
                  </p>
                )}
              </div>

              <div className="rounded-lg border border-primary/50 bg-primary/5 p-6">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div>
                    <h2 className="text-lg font-semibold">Fixed Price Summary</h2>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Total deposit equals the calculated fixed price.
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">Total</p>
                    <p className="text-3xl font-bold">{formatUve(totals.totalUve)}</p>
                    <p className="text-sm text-muted-foreground">
                      {totals.hasUsd
                        ? `${formatCurrency(totals.totalUsd)} USD (oracle)`
                        : 'USD estimate unavailable'}
                    </p>
                  </div>
                </div>
              </div>

              <div className="flex flex-wrap gap-4">
                <Link
                  href={`/marketplace/${provider}/${sequence}`}
                  className="flex-1 rounded-lg border border-border px-4 py-3 text-center text-sm hover:bg-accent"
                >
                  Cancel
                </Link>
                <button
                  type="button"
                  disabled={!canProceed}
                  onClick={() => setStep('review')}
                  className="flex-1 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Review & Sign
                </button>
              </div>
            </>
          )}

          {step === 'review' && (
            <>
              <div className="rounded-lg border border-border bg-card p-6">
                <h2 className="text-lg font-semibold">3. Escrow Deposit</h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  Funds are locked in escrow until the provider completes deployment.
                </p>
                <div className="mt-4 grid gap-4 sm:grid-cols-2">
                  <div className="rounded-lg border border-border bg-muted/40 p-4">
                    <p className="text-sm text-muted-foreground">Escrow account</p>
                    <p className="mt-1 font-mono text-sm">{escrowAccount}</p>
                  </div>
                  <div className="rounded-lg border border-border bg-muted/40 p-4">
                    <p className="text-sm text-muted-foreground">Deposit amount</p>
                    <p className="mt-1 text-lg font-semibold">{formatUve(totals.totalUve)}</p>
                    <p className="text-sm text-muted-foreground">
                      {totals.hasUsd
                        ? `${formatCurrency(totals.totalUsd)} USD (oracle)`
                        : 'USD estimate unavailable'}
                    </p>
                  </div>
                </div>
              </div>

              <div className="rounded-lg border border-border bg-card p-6">
                <h2 className="text-lg font-semibold">Transaction Preview</h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  MsgCreateOrder payload for signing.
                </p>
                <pre className="mt-4 overflow-x-auto rounded-lg bg-muted/40 p-4 text-xs">
                  {JSON.stringify(msgPreview, null, 2)}
                </pre>
              </div>

              <div className="rounded-lg border border-border bg-card p-6">
                <h2 className="text-lg font-semibold">Resources</h2>
                <ul className="mt-4 space-y-2 text-sm">
                  {resourceSelections.map((item) => (
                    <li key={`${item.resourceType}-${item.unit}`} className="flex justify-between">
                      <span className="capitalize">
                        {item.resourceType} ({item.unit})
                      </span>
                      <span className="font-mono">{item.quantity}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <div className="flex flex-wrap gap-4">
                <button
                  type="button"
                  onClick={() => setStep('configure')}
                  className="flex-1 rounded-lg border border-border px-4 py-3 text-sm hover:bg-accent"
                >
                  Back
                </button>
                <button
                  type="button"
                  disabled={status !== 'connected'}
                  onClick={() => void handleSign()}
                  className="flex-1 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {status === 'connected' ? 'Sign & Submit' : 'Connect wallet to sign'}
                </button>
              </div>

              <div className="rounded-lg border border-border bg-muted/40 p-4 text-sm text-muted-foreground">
                <p>
                  Signing as: {account ? truncateAddress(account.address) : 'Wallet not connected'}
                </p>
                {orderName && <p>Order name: {orderName}</p>}
                {notes && <p>Notes: {notes}</p>}
              </div>
            </>
          )}

          {step === 'signing' && (
            <div className="rounded-lg border border-border bg-card p-6 text-center">
              <div className="mx-auto h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent" />
              <h2 className="mt-4 text-lg font-semibold">Signing transaction</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Confirm the MsgCreateOrder transaction in your wallet.
              </p>
            </div>
          )}

          {step === 'confirmed' && (
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="text-2xl font-semibold">Order confirmed</h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Your order has been created and funds are held in escrow.
              </p>
              <div className="mt-6 grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border border-border bg-muted/40 p-4">
                  <p className="text-sm text-muted-foreground">Order ID</p>
                  <p className="mt-1 font-mono text-sm">{orderId}</p>
                </div>
                <div className="rounded-lg border border-border bg-muted/40 p-4">
                  <p className="text-sm text-muted-foreground">Transaction hash</p>
                  <p className="mt-1 font-mono text-sm">{txHash}</p>
                </div>
                <div className="rounded-lg border border-border bg-muted/40 p-4">
                  <p className="text-sm text-muted-foreground">Escrow deposit</p>
                  <p className="mt-1 text-lg font-semibold">{formatUve(totals.totalUve)}</p>
                </div>
                <div className="rounded-lg border border-border bg-muted/40 p-4">
                  <p className="text-sm text-muted-foreground">USD equivalent</p>
                  <p className="mt-1 text-lg font-semibold">
                    {totals.hasUsd ? formatCurrency(totals.totalUsd) : 'Unavailable'}
                  </p>
                </div>
              </div>

              <div className="mt-6 flex flex-wrap gap-4">
                <Link
                  href={`/orders/${orderId}`}
                  className="flex-1 rounded-lg bg-primary px-4 py-3 text-center text-sm font-medium text-primary-foreground hover:bg-primary/90"
                >
                  View Order
                </Link>
                <Link
                  href="/orders"
                  className="flex-1 rounded-lg border border-border px-4 py-3 text-center text-sm hover:bg-accent"
                >
                  Go to Orders
                </Link>
              </div>
            </div>
          )}
        </div>

        <div className="space-y-6">
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Summary</h2>
            <div className="mt-4 space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Pricing model</span>
                <span>Fixed</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Offer sequence</span>
                <span>#{offering.id.sequence}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Selected region</span>
                <span>{region || 'Select'}</span>
              </div>
              <div className="border-t border-border pt-3">
                <div className="flex justify-between font-medium">
                  <span>Total deposit</span>
                  <span>{formatUve(totals.totalUve)}</span>
                </div>
                <div className="flex justify-between text-muted-foreground">
                  <span>USD equivalent</span>
                  <span>{totals.hasUsd ? formatCurrency(totals.totalUsd) : 'Unavailable'}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Wallet Status</h2>
            <div className="mt-3 text-sm">
              {status === 'connected' && account ? (
                <p>Connected as {truncateAddress(account.address)}</p>
              ) : (
                <p className="text-muted-foreground">Connect your wallet to sign the order.</p>
              )}
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Escrow Notice</h2>
            <p className="mt-3 text-sm text-muted-foreground">
              Deposits are held in escrow and released to the provider when deployment succeeds.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
