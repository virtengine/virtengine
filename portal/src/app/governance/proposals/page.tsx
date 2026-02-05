import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Governance Proposals',
  description: 'View and vote on governance proposals',
};

export default function GovernanceProposalsPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Governance</h1>
        <p className="mt-1 text-muted-foreground">View and vote on protocol governance proposals</p>
      </div>

      {/* Stats */}
      <div className="mb-8 grid gap-4 sm:grid-cols-3">
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Active Proposals</div>
          <div className="mt-1 text-2xl font-bold">3</div>
        </div>
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Your Voting Power</div>
          <div className="mt-1 text-2xl font-bold">1,250 VE</div>
        </div>
        <div className="rounded-lg border border-border bg-card p-4">
          <div className="text-sm text-muted-foreground">Participation Rate</div>
          <div className="mt-1 text-2xl font-bold">42%</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-4 border-b border-border">
        <button
          type="button"
          className="border-b-2 border-primary px-4 py-2 text-sm font-medium text-primary"
        >
          Active
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Passed
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Rejected
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          All
        </button>
      </div>

      {/* Proposals List */}
      <div className="space-y-4">
        <ProposalCard
          id="42"
          title="Increase Provider Commission Cap"
          description="Proposal to increase the maximum provider commission from 20% to 25% to attract more infrastructure providers."
          status="active"
          yesPercentage={68}
          endTime="2 days"
        />
        <ProposalCard
          id="41"
          title="Add Support for ARM Architecture"
          description="Enable ARM-based compute offerings in the marketplace to expand hardware diversity."
          status="active"
          yesPercentage={82}
          endTime="5 days"
        />
        <ProposalCard
          id="40"
          title="Community Fund Allocation Q1 2024"
          description="Allocate 500,000 VE tokens to the community development fund for grants and bounties."
          status="active"
          yesPercentage={91}
          endTime="7 days"
        />
        <ProposalCard
          id="39"
          title="Update Minimum Stake Requirements"
          description="Reduce minimum stake for providers from 10,000 VE to 5,000 VE to lower barriers to entry."
          status="passed"
          yesPercentage={76}
        />
        <ProposalCard
          id="38"
          title="Protocol Fee Adjustment"
          description="Proposal to reduce protocol fees from 5% to 3% on all marketplace transactions."
          status="rejected"
          yesPercentage={34}
        />
      </div>
    </div>
  );
}

interface ProposalCardProps {
  id: string;
  title: string;
  description: string;
  status: 'active' | 'passed' | 'rejected';
  yesPercentage: number;
  endTime?: string;
}

function ProposalCard({
  id,
  title,
  description,
  status,
  yesPercentage,
  endTime,
}: ProposalCardProps) {
  const statusConfig = {
    active: { bg: 'bg-primary/10', text: 'text-primary', label: 'Voting' },
    passed: { bg: 'bg-success/10', text: 'text-success', label: 'Passed' },
    rejected: { bg: 'bg-destructive/10', text: 'text-destructive', label: 'Rejected' },
  };
  const config = statusConfig[status];

  return (
    <Link
      href={`/governance/proposals/${id}`}
      className="card-hover block rounded-lg border border-border bg-card p-6 transition-all"
    >
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <span className="font-mono text-sm text-muted-foreground">#{id}</span>
          <span
            className={`rounded-full ${config.bg} ${config.text} px-2 py-1 text-xs font-medium`}
          >
            {config.label}
          </span>
        </div>
        {endTime && <span className="text-sm text-muted-foreground">Ends in {endTime}</span>}
      </div>

      <h3 className="mt-3 text-lg font-semibold">{title}</h3>
      <p className="mt-2 line-clamp-2 text-sm text-muted-foreground">{description}</p>

      <div className="mt-4">
        <div className="flex justify-between text-sm">
          <span>Yes: {yesPercentage}%</span>
          <span>No: {100 - yesPercentage}%</span>
        </div>
        <div className="mt-2 flex h-2 overflow-hidden rounded-full bg-muted">
          <div className="bg-success" style={{ width: `${yesPercentage}%` }} />
          <div className="bg-destructive" style={{ width: `${100 - yesPercentage}%` }} />
        </div>
      </div>

      {status === 'active' && (
        <div className="mt-4 flex gap-2">
          <Link
            href={`/governance/proposals/${id}/vote?vote=yes`}
            className="flex-1 rounded-lg bg-success px-4 py-2 text-center text-sm text-white hover:bg-success/90"
          >
            Vote Yes
          </Link>
          <Link
            href={`/governance/proposals/${id}/vote?vote=no`}
            className="flex-1 rounded-lg bg-destructive px-4 py-2 text-center text-sm text-white hover:bg-destructive/90"
          >
            Vote No
          </Link>
        </div>
      )}
    </Link>
  );
}
