import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Provider Dashboard',
  description: 'Manage your provider infrastructure and offerings',
};

export default function ProviderDashboardPage() {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Provider Dashboard</h1>
        <p className="mt-1 text-muted-foreground">
          Manage your infrastructure and monitor performance
        </p>
      </div>

      {/* Stats Overview */}
      <div className="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Active Leases" value="12" change="+2 this week" positive />
        <StatCard title="Revenue (30d)" value="$4,250" change="+15%" positive />
        <StatCard title="Uptime" value="99.8%" change="0.1% below target" />
        <StatCard title="Pending Bids" value="5" change="3 expiring soon" />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        {/* Resource Utilization */}
        <div className="rounded-lg border border-border bg-card p-6">
          <h2 className="text-lg font-semibold">Resource Utilization</h2>
          <div className="mt-4 space-y-4">
            <UtilizationBar label="CPU" used={65} total={100} unit="cores" />
            <UtilizationBar label="Memory" used={128} total={256} unit="GB" />
            <UtilizationBar label="GPU" used={4} total={8} unit="units" />
            <UtilizationBar label="Storage" used={2.4} total={10} unit="TB" />
          </div>
        </div>

        {/* Recent Activity */}
        <div className="rounded-lg border border-border bg-card p-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Recent Activity</h2>
            <Link href="/provider/orders" className="text-sm text-primary hover:underline">
              View all
            </Link>
          </div>
          <div className="mt-4 space-y-3">
            <ActivityItem
              title="New lease accepted"
              description="Order #1234 - GPU Compute"
              time="2 hours ago"
              type="success"
            />
            <ActivityItem
              title="Bid placed"
              description="HPC Cluster - $8.50/hr"
              time="5 hours ago"
              type="info"
            />
            <ActivityItem
              title="Lease completed"
              description="Order #1201 - $425.00 earned"
              time="1 day ago"
              type="success"
            />
            <ActivityItem
              title="Bid expired"
              description="CPU Compute - No match"
              time="2 days ago"
              type="warning"
            />
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="mt-8">
        <h2 className="mb-4 text-lg font-semibold">Quick Actions</h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <QuickActionCard
            title="Create Offering"
            description="List new compute resources"
            href="/provider/offerings/new"
            icon="➕"
          />
          <QuickActionCard
            title="Update Pricing"
            description="Adjust your pricing strategy"
            href="/provider/pricing"
            icon="💰"
          />
          <QuickActionCard
            title="View Orders"
            description="Manage active deployments"
            href="/provider/orders"
            icon="📋"
          />
          <QuickActionCard
            title="Configure Capacity"
            description="Set resource limits"
            href="/provider/offerings"
            icon="⚙️"
          />
        </div>
      </div>
    </div>
  );
}

function StatCard({
  title,
  value,
  change,
  positive,
}: {
  title: string;
  value: string;
  change: string;
  positive?: boolean;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="text-sm text-muted-foreground">{title}</div>
      <div className="mt-2 text-3xl font-bold">{value}</div>
      <div className={`mt-1 text-sm ${positive ? 'text-success' : 'text-muted-foreground'}`}>
        {change}
      </div>
    </div>
  );
}

function UtilizationBar({
  label,
  used,
  total,
  unit,
}: {
  label: string;
  used: number;
  total: number;
  unit: string;
}) {
  const percentage = Math.round((used / total) * 100);
  const barColor = percentage > 80 ? 'bg-destructive' : percentage > 60 ? 'bg-warning' : 'bg-success';

  return (
    <div>
      <div className="flex justify-between text-sm">
        <span>{label}</span>
        <span className="text-muted-foreground">
          {used} / {total} {unit} ({percentage}%)
        </span>
      </div>
      <div className="mt-2 h-2 rounded-full bg-muted">
        <div className={`h-full rounded-full ${barColor}`} style={{ width: `${percentage}%` }} />
      </div>
    </div>
  );
}

function ActivityItem({
  title,
  description,
  time,
  type,
}: {
  title: string;
  description: string;
  time: string;
  type: 'success' | 'warning' | 'info';
}) {
  const dotColors = {
    success: 'status-dot-success',
    warning: 'status-dot-warning',
    info: 'status-dot-pending',
  };

  return (
    <div className="flex items-start gap-3">
      <span className={`status-dot mt-2 ${dotColors[type]}`} />
      <div className="flex-1">
        <div className="text-sm font-medium">{title}</div>
        <div className="text-sm text-muted-foreground">{description}</div>
      </div>
      <div className="text-xs text-muted-foreground">{time}</div>
    </div>
  );
}

function QuickActionCard({
  title,
  description,
  href,
  icon,
}: {
  title: string;
  description: string;
  href: string;
  icon: string;
}) {
  return (
    <Link
      href={href}
      className="group rounded-lg border border-border bg-card p-4 transition-all card-hover"
    >
      <div className="text-2xl">{icon}</div>
      <h3 className="mt-2 font-medium group-hover:text-primary">{title}</h3>
      <p className="mt-1 text-sm text-muted-foreground">{description}</p>
    </Link>
  );
}
