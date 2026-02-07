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
import { usePriceConversion } from '@/hooks/usePriceConversion';
import { useTranslation } from 'react-i18next';
import { formatCurrency } from '@/lib/utils';

interface EscrowDepositProps {
  escrowInfo: EscrowInfo;
  priceBreakdown: PriceBreakdown;
  isSubmitting: boolean;
  error: string | null;
  onSubmit: () => void;
}

function formatUsd(value: number): string {
  const precision = value < 0.01 ? 4 : 2;
  return formatCurrency(Number(value.toFixed(precision)), 'USD');
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
  const { t } = useTranslation();
  const { uveToUsd, isLoading: rateLoading } = usePriceConversion();

  return (
    <div className="space-y-6">
      {/* Wallet Balance */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('Wallet Balance')}</CardTitle>
          <CardDescription>{t('Your current wallet balance for this transaction')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between rounded-lg bg-muted/50 p-4">
            <div>
              <p className="text-sm text-muted-foreground">{t('Available Balance')}</p>
              <p className="text-2xl font-bold">
                {t('{{amount}} {{currency}}', {
                  amount: formatTokenAmount(escrowInfo.walletBalanceUsd),
                  currency: priceBreakdown.currency,
                })}
              </p>
              {!rateLoading && uveToUsd(escrowInfo.walletBalanceUsd) !== null && (
                <p className="text-sm text-muted-foreground">
                  {t('~{{amount}} USD', {
                    amount: formatUsd(uveToUsd(escrowInfo.walletBalanceUsd)!),
                  })}
                </p>
              )}
            </div>
            <div className="text-right">
              <p className="text-sm text-muted-foreground">{t('Required Deposit')}</p>
              <p className="text-2xl font-bold text-primary">
                {t('{{amount}} {{currency}}', {
                  amount: formatTokenAmount(escrowInfo.depositUsd),
                  currency: priceBreakdown.currency,
                })}
              </p>
              {!rateLoading && uveToUsd(escrowInfo.depositUsd) !== null && (
                <p className="text-sm text-muted-foreground">
                  {t('~{{amount}} USD', {
                    amount: formatUsd(uveToUsd(escrowInfo.depositUsd)!),
                  })}
                </p>
              )}
            </div>
          </div>

          {!escrowInfo.hasSufficientFunds && (
            <Alert className="mt-4" variant="destructive">
              <p className="text-sm">
                {t(
                  'Insufficient funds. You need at least {{amount}} {{currency}} to create this order. Please add funds to your wallet.',
                  {
                    amount: formatTokenAmount(escrowInfo.depositUsd),
                    currency: priceBreakdown.currency,
                  }
                )}
              </p>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Transaction Details */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('Transaction Details')}</CardTitle>
          <CardDescription>{t('Review before approving the escrow deposit')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{t('Transaction Type')}</span>
              <span className="font-medium">MsgCreateOrder</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{t('Escrow Deposit')}</span>
              <span className="font-medium">
                {t('{{amount}} {{currency}}', {
                  amount: formatTokenAmount(escrowInfo.depositUsd),
                  currency: priceBreakdown.currency,
                })}
                {!rateLoading && uveToUsd(escrowInfo.depositUsd) !== null && (
                  <span className="ml-1 text-muted-foreground">
                    {t('(~{{amount}})', { amount: formatUsd(uveToUsd(escrowInfo.depositUsd)!) })}
                  </span>
                )}
              </span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">{t('Estimated Gas Fee')}</span>
              <span className="font-medium">
                {t('~0.01 {{currency}}', { currency: priceBreakdown.currency })}
              </span>
            </div>

            <Separator />

            <div className="flex items-center justify-between">
              <span className="font-semibold">{t('Total Debit')}</span>
              <div className="text-right">
                <span className="text-lg font-bold text-primary">
                  {t('{{amount}} {{currency}}', {
                    amount: formatTokenAmount(escrowInfo.depositUsd + 0.01),
                    currency: priceBreakdown.currency,
                  })}
                </span>
                {!rateLoading && uveToUsd(escrowInfo.depositUsd + 0.01) !== null && (
                  <p className="text-sm text-muted-foreground">
                    {t('~{{amount}} USD', {
                      amount: formatUsd(uveToUsd(escrowInfo.depositUsd + 0.01)!),
                    })}
                  </p>
                )}
              </div>
            </div>
          </div>

          <div className="mt-6 rounded-md border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-200">
            <strong>{t('Escrow:')}</strong>{' '}
            {t(
              'Your deposit is held in an on-chain escrow account. It will be used to pay for resources as they are consumed. Unused funds are returned when the order is closed.'
            )}
          </div>
        </CardContent>
      </Card>

      {/* Error Display */}
      {error && (
        <Alert variant="destructive">
          <p className="text-sm font-medium">{t('Transaction Error')}</p>
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
            {isSubmitting ? t('Approving Transaction...') : t('Approve & Create Order')}
          </Button>
          <p className="mt-2 text-center text-xs text-muted-foreground">
            {t('Your wallet will prompt you to sign this transaction')}
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
