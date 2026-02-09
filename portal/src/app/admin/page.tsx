'use client';

import Link from 'next/link';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';
import {
  useAdminStore,
  selectActiveProposals,
  selectActiveValidators,
  selectOpenTickets,
  selectUrgentTickets,
} from '@/stores/adminStore';
import { formatRelativeTime } from '@/lib/utils';
import { useTranslation } from 'react-i18next';

const alertStyles = {
  info: 'bg-blue-100 text-blue-700',
  warning: 'bg-amber-100 text-amber-700',
  critical: 'bg-rose-100 text-rose-700',
};

export default function AdminDashboardPage() {
  const { t } = useTranslation();
  const proposals = useAdminStore(selectActiveProposals);
  const validators = useAdminStore(selectActiveValidators);
  const openTickets = useAdminStore(selectOpenTickets);
  const urgentTickets = useAdminStore(selectUrgentTickets);
  const systemHealth = useAdminStore((s) => s.systemHealth);
  const providers = useAdminStore((s) => s.providers);
  const escrowOverview = useAdminStore((s) => s.escrowOverview);
  const resourceUtilization = useAdminStore((s) => s.resourceUtilization);
  const networkAlerts = useAdminStore((s) => s.networkAlerts);
  const recentBlocks = useAdminStore((s) => s.recentBlocks);
  const disputes = useAdminStore((s) => s.disputes);
  const hasRole = useAdminStore((s) => s.hasRole);

  const activeLeases = providers.reduce((sum, provider) => sum + provider.activeLeases, 0);
  const providersAtRisk = providers.filter((provider) => provider.status !== 'active').length;
  const openDisputes = disputes.filter((dispute) => dispute.status !== 'resolved').length;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">{t('Admin Dashboard')}</h1>
        <p className="mt-1 text-muted-foreground">
          {t('Network health and operational readiness')}
        </p>
      </div>

      <div className="flex flex-wrap gap-2">
        {hasRole('operator') && <Badge variant="default">Operator</Badge>}
        {hasRole('governance') && <Badge variant="info">Governance</Badge>}
        {hasRole('validator') && <Badge variant="success">Validator</Badge>}
        {hasRole('support') && <Badge variant="warning">Support</Badge>}
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Block Height</div>
            <div className="mt-1 text-2xl font-bold">
              {systemHealth.blockHeight.toLocaleString()}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active Validators</div>
            <div className="mt-1 text-2xl font-bold">
              {systemHealth.activeValidators}/{systemHealth.totalValidators}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active Leases</div>
            <div className="mt-1 text-2xl font-bold">{activeLeases}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Total Escrow</div>
            <div className="mt-1 text-2xl font-bold">
              {escrowOverview.totalEscrow.toLocaleString()} VE
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Providers at Risk</div>
            <div className="mt-1 text-2xl font-bold">{providersAtRisk}</div>
            <p className="text-xs text-muted-foreground">Needs follow-up</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Open Disputes</div>
            <div className="mt-1 text-2xl font-bold">{openDisputes}</div>
            <p className="text-xs text-muted-foreground">Escrow review queue</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t('Resource Utilization')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {resourceUtilization.map((resource) => {
              const pct = Math.round((resource.usage / resource.capacity) * 100);
              const variant = pct > 95 ? 'destructive' : pct > 85 ? 'warning' : 'success';
              return (
                <div key={resource.category} className="space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{resource.category}</span>
                    <span className="text-muted-foreground">
                      {resource.usage}/{resource.capacity} ({pct}%)
                    </span>
                  </div>
                  <Progress value={pct} size="sm" variant={variant} />
                </div>
              );
            })}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('System Alerts')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {networkAlerts.map((alert) => (
              <div
                key={alert.id}
                className="rounded-lg border border-border p-3 text-sm transition"
              >
                <div className="flex items-center justify-between">
                  <span className="font-medium">{alert.title}</span>
                  <Badge className={alertStyles[alert.severity]}>{alert.severity}</Badge>
                </div>
                <p className="mt-1 text-xs text-muted-foreground">{alert.description}</p>
                <p className="mt-2 text-xs text-muted-foreground">
                  {formatRelativeTime(alert.createdAt)}
                </p>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t('Recent Blocks')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {recentBlocks.map((block) => (
              <div
                key={block.height}
                className="flex items-center justify-between rounded-lg border border-border p-3"
              >
                <div>
                  <div className="text-sm font-semibold">#{block.height}</div>
                  <div className="text-xs text-muted-foreground">
                    {block.proposer} · {block.txCount} txs
                  </div>
                </div>
                <span className="text-xs text-muted-foreground">
                  {formatRelativeTime(block.timestamp)}
                </span>
              </div>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('Open Support Tickets')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {openTickets.slice(0, 3).map((ticket) => (
              <Link
                key={ticket.id}
                href="/admin/support"
                className="block rounded-lg border border-border p-3 transition hover:border-foreground/40"
              >
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold">{ticket.subject}</span>
                  <Badge
                    className={
                      ticket.priority === 'urgent'
                        ? 'bg-rose-100 text-rose-700'
                        : ticket.priority === 'high'
                          ? 'bg-amber-100 text-amber-700'
                          : 'bg-slate-100 text-slate-600'
                    }
                  >
                    {ticket.priority}
                  </Badge>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">
                  {ticket.ticketNumber} · {ticket.provider} · Updated{' '}
                  {formatRelativeTime(ticket.updatedAt)}
                </div>
              </Link>
            ))}
            {urgentTickets.length > 0 && (
              <div className="text-xs text-rose-600">
                {urgentTickets.length} urgent tickets need attention.
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t('Governance Proposals')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {proposals.slice(0, 3).map((proposal) => {
              const total =
                proposal.yesVotes + proposal.noVotes + proposal.abstainVotes + proposal.vetoVotes;
              const yesPct = total > 0 ? Math.round((proposal.yesVotes / total) * 100) : 0;
              return (
                <Link
                  key={proposal.id}
                  href="/admin/governance"
                  className="block rounded-lg border border-border p-3 transition hover:border-foreground/40"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-semibold">
                      #{proposal.id} {proposal.title}
                    </span>
                    <Badge className="bg-primary/10 text-primary">Voting</Badge>
                  </div>
                  <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
                    <span>Yes: {yesPct}%</span>
                    <span>·</span>
                    <span>Ends {formatRelativeTime(proposal.votingEndTime)}</span>
                  </div>
                </Link>
              );
            })}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t('Validator Health')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {validators.map((validator) => (
              <div
                key={validator.operatorAddress}
                className="flex items-center justify-between rounded-lg border border-border p-3"
              >
                <div>
                  <div className="text-sm font-semibold">{validator.moniker}</div>
                  <div className="text-xs text-muted-foreground">
                    {validator.uptime}% uptime · {validator.missedBlocks} missed
                  </div>
                </div>
                <Progress
                  value={validator.uptime}
                  size="sm"
                  variant={
                    validator.uptime >= 99
                      ? 'success'
                      : validator.uptime >= 95
                        ? 'warning'
                        : 'destructive'
                  }
                  className="w-24"
                />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
