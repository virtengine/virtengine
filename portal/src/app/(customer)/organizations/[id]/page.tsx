import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Organization Details',
  description: 'View and manage your organization',
};

const mockMembers = [
  { address: 'virtengine1abc...xyz', role: 'Admin', weight: '1', addedDaysAgo: 30 },
  { address: 'virtengine1def...uvw', role: 'Member', weight: '1', addedDaysAgo: 14 },
  { address: 'virtengine1ghi...rst', role: 'Member', weight: '1', addedDaysAgo: 7 },
  { address: 'virtengine1jkl...opq', role: 'Viewer', weight: '0', addedDaysAgo: 2 },
] as const;

const mockBilling = {
  currentPeriod: 245.5,
  previousPeriod: 198.3,
  totalSpend: 1432.0,
  changePercent: 23.8,
} as const;

export default function OrganizationDetailPage({ params: _params }: { params: { id: string } }) {
  return (
    <div className="container py-8">
      {/* Header */}
      <div className="mb-6">
        <Link
          href="/organizations"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to organizations
        </Link>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
              <span className="text-xl font-semibold text-primary">A</span>
            </div>
            <div>
              <h1 className="text-2xl font-bold">Acme Corp</h1>
              <p className="text-sm text-muted-foreground">Main production deployments</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className="rounded-full bg-primary/10 px-3 py-1 text-sm text-primary">Admin</span>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="mb-6 grid gap-4 md:grid-cols-4">
        <StatCard label="Current Period" value={`$${mockBilling.currentPeriod}`} />
        <StatCard label="Previous Period" value={`$${mockBilling.previousPeriod}`} />
        <StatCard
          label="Change"
          value={`+${mockBilling.changePercent}%`}
          valueColor="text-destructive"
        />
        <StatCard label="Total Spend" value={`$${mockBilling.totalSpend}`} />
      </div>

      {/* Tabs */}
      <div className="mb-6 flex gap-4 border-b border-border">
        <button
          type="button"
          className="border-b-2 border-primary px-4 py-2 text-sm font-medium text-primary"
        >
          Members ({mockMembers.length})
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Billing
        </button>
        <button
          type="button"
          className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground"
        >
          Settings
        </button>
      </div>

      {/* Members section */}
      <div className="space-y-4">
        <div className="flex justify-end">
          <button
            type="button"
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            + Invite Member
          </button>
        </div>

        <div className="space-y-2">
          {mockMembers.map((member) => (
            <MemberRow key={member.address} member={member} />
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  valueColor,
}: {
  label: string;
  value: string;
  valueColor?: string;
}) {
  return (
    <div className="rounded-lg border p-4">
      <p className="text-sm text-muted-foreground">{label}</p>
      <p className={`mt-1 text-2xl font-semibold ${valueColor || ''}`}>{value}</p>
    </div>
  );
}

function MemberRow({
  member,
}: {
  member: {
    address: string;
    role: string;
    weight: string;
    addedDaysAgo: number;
  };
}) {
  const roleColors: Record<string, string> = {
    Admin: 'text-primary',
    Member: 'text-foreground',
    Viewer: 'text-muted-foreground',
  };

  return (
    <div className="flex items-center justify-between rounded-lg border p-3">
      <div className="flex items-center gap-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
          <span className="text-sm">{member.address.slice(0, 2).toUpperCase()}</span>
        </div>
        <div>
          <p className="font-mono text-sm">{member.address}</p>
          <p className={`text-xs capitalize ${roleColors[member.role] || ''}`}>{member.role}</p>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <span className="text-xs text-muted-foreground">{member.addedDaysAgo}d ago</span>
        {member.role !== 'Admin' && (
          <button type="button" className="text-sm text-destructive hover:underline">
            Remove
          </button>
        )}
      </div>
    </div>
  );
}
