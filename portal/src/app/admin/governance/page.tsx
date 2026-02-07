'use client';

import { Badge } from '@/components/ui/Badge';
import { Card, CardContent } from '@/components/ui/Card';
import { useAdminStore } from '@/stores/adminStore';
import { formatDate } from '@/lib/utils';
import type { ProposalStatus } from '@/types/admin';

const statusConfig: Record<ProposalStatus, { bg: string; text: string; label: string }> = {
  voting: { bg: 'bg-primary/10', text: 'text-primary', label: 'Voting' },
  passed: { bg: 'bg-emerald-100', text: 'text-emerald-700', label: 'Passed' },
  rejected: { bg: 'bg-rose-100', text: 'text-rose-700', label: 'Rejected' },
  deposit: { bg: 'bg-amber-100', text: 'text-amber-700', label: 'Deposit' },
};

export default function AdminGovernancePage() {
  const proposals = useAdminStore((s) => s.proposals);

  const votingCount = proposals.filter((p) => p.status === 'voting').length;
  const passedCount = proposals.filter((p) => p.status === 'passed').length;
  const rejectedCount = proposals.filter((p) => p.status === 'rejected').length;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">Governance</h1>
        <p className="mt-1 text-muted-foreground">Monitor and manage governance proposals</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Active Voting</div>
            <div className="mt-1 text-2xl font-bold">{votingCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Passed</div>
            <div className="mt-1 text-2xl font-bold text-emerald-600">{passedCount}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Rejected</div>
            <div className="mt-1 text-2xl font-bold text-rose-600">{rejectedCount}</div>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-4">
        {proposals.map((proposal) => {
          const total =
            proposal.yesVotes + proposal.noVotes + proposal.abstainVotes + proposal.vetoVotes;
          const yesPct = total > 0 ? Math.round((proposal.yesVotes / total) * 100) : 0;
          const noPct = total > 0 ? Math.round((proposal.noVotes / total) * 100) : 0;
          const abstainPct = total > 0 ? Math.round((proposal.abstainVotes / total) * 100) : 0;
          const vetoPct = total > 0 ? Math.round((proposal.vetoVotes / total) * 100) : 0;
          const config = statusConfig[proposal.status];

          return (
            <Card key={proposal.id}>
              <CardContent className="p-6">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <span className="font-mono text-sm text-muted-foreground">#{proposal.id}</span>
                    <Badge className={`${config.bg} ${config.text}`}>{config.label}</Badge>
                  </div>
                  <span className="text-sm text-muted-foreground">
                    {proposal.status === 'voting'
                      ? `Ends ${formatDate(proposal.votingEndTime)}`
                      : `Ended ${formatDate(proposal.votingEndTime)}`}
                  </span>
                </div>

                <h3 className="mt-3 text-lg font-semibold">{proposal.title}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{proposal.description}</p>
                <p className="mt-2 text-xs text-muted-foreground">
                  Proposed by {proposal.proposer} · Submitted {formatDate(proposal.submitTime)}
                </p>

                <div className="mt-4 space-y-2">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-emerald-600">Yes: {yesPct}%</span>
                    <span className="text-rose-600">No: {noPct}%</span>
                    <span className="text-muted-foreground">Abstain: {abstainPct}%</span>
                    <span className="text-orange-600">Veto: {vetoPct}%</span>
                  </div>
                  <div className="flex h-2 overflow-hidden rounded-full bg-muted">
                    <div className="bg-emerald-500" style={{ width: `${yesPct}%` }} />
                    <div className="bg-rose-500" style={{ width: `${noPct}%` }} />
                    <div className="bg-slate-400" style={{ width: `${abstainPct}%` }} />
                    <div className="bg-orange-500" style={{ width: `${vetoPct}%` }} />
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-4 gap-2 text-xs text-muted-foreground">
                  <span>Yes: {proposal.yesVotes.toLocaleString()}</span>
                  <span>No: {proposal.noVotes.toLocaleString()}</span>
                  <span>Abstain: {proposal.abstainVotes.toLocaleString()}</span>
                  <span>Veto: {proposal.vetoVotes.toLocaleString()}</span>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </div>
  );
}
