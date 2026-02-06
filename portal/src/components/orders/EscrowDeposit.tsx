/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { Separator } from '@/components/ui/Separator';
import type { EscrowInfo, PriceBreakdown } from '@/features/orders';
import { formatTokenAmount } from '@/features/orders';

interface EscrowDepositProps {
  escrowInfo: EscrowInfo;
  priceBreakdown: PriceBreakdown;
  isSubmitting: boolean;
  error: string | null;
  onSubmit: () => void;
}

/**
 * Step 3: Escrow Deposit
 * Shows wallet balance, deposit amount, and transaction approval.
 */
export function EscrowDeposit({
  escrowInfo,
  priceBreakdown,
  isSubmitting,
  error,
  onSubmit,
}: EscrowDepositProps) {
  return (
    <div className="space-y-6">
      {/* Wallet Balance */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Wallet Balance</CardTitle>
          <CardDescription>Your current wallet balance for this transaction</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between rounded-lg bg-muted/50 p-4">
            <div>
              <p className="text-sm text-muted-foreground">Available Balance</p>
              <p className="text-2xl font-bold">
                {formatTokenAmount(escrowInfo.walletBalanceUsd)} {priceBreakdown.currency}
              </p>
            </div>
            <div className="text-right">
              <p className="text-sm text-muted-foreground">Required Deposit</p>
              <p className="text-2xl font-bold text-primary">
                {formatTokenAmount(escrowInfo.depositUsd)} {priceBreakdown.currency}
              </p>
            </div>
          </div>

          {!escrowInfo.hasSufficientFunds && (
            <Alert className="mt-4" variant="destructive">
              <p className="text-sm">
                Insufficient funds. You need at least{' '}
                <strong>
                  {formatTokenAmount(escrowInfo.depositUsd)} {priceBreakdown.currency}
                </strong>{' '}
                to create this order. Please add funds to your wallet.
              </p>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Transaction Details */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Transaction Details</CardTitle>
          <CardDescription>Review before approving the escrow deposit</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Transaction Type</span>
              <span className="font-medium">MsgCreateOrder</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Escrow Deposit</span>
              <span className="font-medium">
                {formatTokenAmount(escrowInfo.depositUsd)} {priceBreakdown.currency}
              </span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Estimated Gas Fee</span>
              <span className="font-medium">~0.01 {priceBreakdown.currency}</span>
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <span className="font-semibold">Total Debit</span>
              <span className="text-lg font-bold text-primary">
                {formatTokenAmount(escrowInfo.depositUsd + 0.01)} {priceBreakdown.currency}
              </span>
            </div>
          </div>

          <div className="mt-6 rounded-md border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-200">
            <strong>Escrow:</strong> Your deposit is held in an on-chain escrow account. It will be
            used to pay for resources as they are consumed. Unused funds are returned when the order
            is closed.
          </div>
        </CardContent>
      </Card>

      {/* Error Display */}
      {error && (
        <Alert variant="destructive">
          <p className="text-sm font-medium">Transaction Error</p>
          <p className="mt-1 text-sm">{error}</p>
        </Alert>
      )}

      {/* Approve Button */}
      <Card>
        <CardContent className="py-4">
          <Button
            className="w-full"
            size="lg"
            onClick={onSubmit}
            loading={isSubmitting}
            disabled={!escrowInfo.hasSufficientFunds || isSubmitting}
          >
            {isSubmitting ? 'Approving Transaction...' : 'Approve & Create Order'}
          </Button>
          <p className="mt-2 text-center text-xs text-muted-foreground">
            Your wallet will prompt you to sign this transaction
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
