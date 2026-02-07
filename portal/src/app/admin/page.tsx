'use client';

import Link from 'next/link';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import {
  useAdminStore,
  selectActiveProposals,
  selectActiveValidators,
  selectOpenTickets,
  selectUrgentTickets,
} from '@/stores/adminStore';
import { formatRelativeTime } from '@/lib/utils';

export default function AdminDashboardPage() {
  const proposals = useAdminStore(selectActiveProposals);
  const validators = useAdminStore(selectActiveValidators);
  const openTickets = useAdminStore(selectOpenTickets);
  const urgentTickets = useAdminStore(selectUrgentTickets);
  const systemHealth = useAdminStore((s) => s.systemHealth);
  const hasRole = useAdminStore((s) => s.hasRole);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
        <p className="mt-1 text-muted-foreground">VirtEngine operator overview</p>
      </div>

      {/* Role badges */}
      <div className="flex flex-wrap gap-2">
        {hasRole('operator') && <Badge variant="default">Operator</Badge>}
        {hasRole('governance') && <Badge variant="info">Governance</Badge>}
        {hasRole('validator') && <Badge variant="success">Validator</Badge>}
        {hasRole('support') && <Badge variant="warning">Support</Badge>}
      </div>

      {/* Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Link href="/admin/governance">
          <Card className="transition hover:border-foreground/40">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Active Proposals
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{proposals.length}</div>
            </CardContent>
          </Card>
        </Link>
        <Link href="/admin/validators">
          <Card className="transition hover:border-foreground/40">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Active Validators
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {validators.length}/{systemHealth.totalValidators}
              </div>
            </CardContent>
          </Card>
        </Link>
        <Link href="/admin/support">
          <Card className="transition hover:border-foreground/40">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Open Tickets
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{openTickets.length}</div>
              {urgentTickets.length > 0 && (
                <Badge variant="destructive" className="mt-1">
                  {urgentTickets.length} urgent
                </Badge>
              )}
            </CardContent>
          </Card>
        </Link>
        <Link href="/admin/health">
          <Card className="transition hover:border-foreground/40">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Network Uptime
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{systemHealth.networkUptime}%</div>
            </CardContent>
          </Card>
        </Link>
      </div>

      {/* Recent activity */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Proposals</CardTitle>
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
            <CardTitle>Open Support Tickets</CardTitle>
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
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
