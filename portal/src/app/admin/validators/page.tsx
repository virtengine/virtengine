'use client';

import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import { Progress } from '@/components/ui/Progress';
import { useAdminStore } from '@/stores/adminStore';
import { formatDate, formatTokenAmount } from '@/lib/utils';
import type { ValidatorStatus } from '@/types/admin';
import { useTranslation } from 'react-i18next';

const statusStyles: Record<ValidatorStatus, string> = {
  active: 'bg-emerald-100 text-emerald-700',
  inactive: 'bg-slate-200 text-slate-600',
  jailed: 'bg-rose-100 text-rose-700',
  unbonding: 'bg-amber-100 text-amber-700',
};

export default function AdminValidatorsPage() {
  const { t } = useTranslation();
  const validators = useAdminStore((s) => s.validators);

  const activeCount = validators.filter((v) => v.status === 'active').length;
  const jailedCount = validators.filter((v) => v.status === 'jailed').length;
  const totalStake = validators.reduce((sum, v) => sum + Number(v.tokens), 0);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">{t('Validators')}</h1>
        <p className="mt-1 text-muted-foreground">
          {t('Monitor validator set status and slashing events')}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active</div>
            <div className="mt-1 text-2xl font-bold text-emerald-600">{activeCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total</div>
            <div className="mt-1 text-2xl font-bold">{validators.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Jailed</div>
            <div className="mt-1 text-2xl font-bold text-rose-600">{jailedCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total Stake</div>
            <div className="mt-1 text-2xl font-bold">{formatTokenAmount(totalStake)} VE</div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('Validator Set')}</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Moniker</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Stake</TableHead>
                <TableHead>Commission</TableHead>
                <TableHead>Uptime</TableHead>
                <TableHead>Missed Blocks</TableHead>
                <TableHead>Slashing</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {validators.map((validator) => (
                <TableRow key={validator.operatorAddress}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{validator.moniker}</div>
                      <div className="text-xs text-muted-foreground">
                        {validator.operatorAddress}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge className={statusStyles[validator.status]}>{validator.status}</Badge>
                    {validator.jailedUntil && (
                      <div className="mt-1 text-xs text-muted-foreground">
                        Until {formatDate(validator.jailedUntil)}
                      </div>
                    )}
                  </TableCell>
                  <TableCell>{formatTokenAmount(validator.tokens)} VE</TableCell>
                  <TableCell>{(validator.commission * 100).toFixed(1)}%</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Progress
                        value={validator.uptime}
                        className="w-16"
                        size="sm"
                        variant={
                          validator.uptime >= 99
                            ? 'success'
                            : validator.uptime >= 95
                              ? 'warning'
                              : 'destructive'
                        }
                      />
                      <span className="text-sm">{validator.uptime}%</span>
                    </div>
                  </TableCell>
                  <TableCell>{validator.missedBlocks.toLocaleString()}</TableCell>
                  <TableCell>
                    {validator.slashingEvents.length > 0 ? (
                      <div className="space-y-1">
                        {validator.slashingEvents.map((event) => (
                          <div key={event.id} className="text-xs">
                            <Badge variant="destructive" size="sm">
                              {event.reason === 'double_sign' ? 'Double Sign' : 'Downtime'}
                            </Badge>
                            <span className="ml-1 text-muted-foreground">
                              Block #{event.blockHeight}
                            </span>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <span className="text-sm text-muted-foreground">None</span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
