import type { Metadata } from 'next';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'Organizations',
  description: 'Manage your organizations and team deployments',
};

const orgPlaceholders = [
  {
    id: 'org-1',
    name: 'Acme Corp',
    description: 'Main production deployments',
    members: 12,
    role: 'Admin',
    deployments: 8,
  },
  {
    id: 'org-2',
    name: 'Dev Team',
    description: 'Development and staging workloads',
    members: 5,
    role: 'Member',
    deployments: 3,
  },
  {
    id: 'org-3',
    name: 'Research Lab',
    description: 'ML training and inference jobs',
    members: 3,
    role: 'Viewer',
    deployments: 15,
  },
] satisfies ReadonlyArray<{
  id: string;
  name: string;
  description: string;
  members: number;
  role: string;
  deployments: number;
}>;

export default function OrganizationsPage() {
  return (
    <div className="container py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Organizations</h1>
          <p className="mt-1 text-muted-foreground">
            Manage your organizations and team deployments
          </p>
        </div>
        <button
          type="button"
          className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          + New Organization
        </button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {orgPlaceholders.map((org) => (
          <OrgCard key={org.id} org={org} />
        ))}
      </div>

      {orgPlaceholders.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-4">
            <span className="text-4xl">🏢</span>
          </div>
          <h2 className="mt-4 text-lg font-medium">No organizations yet</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Create an organization to manage team deployments and billing
          </p>
          <button
            type="button"
            className="mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            Create Organization
          </button>
        </div>
      )}
    </div>
  );
}

function OrgCard({
  org,
}: {
  org: {
    id: string;
    name: string;
    description: string;
    members: number;
    role: string;
    deployments: number;
  };
}) {
  const roleColors: Record<string, string> = {
    Admin: 'bg-primary/10 text-primary',
    Member: 'bg-secondary text-secondary-foreground',
    Viewer: 'bg-muted text-muted-foreground',
  };

  return (
    <Link
      href={`/organizations/${org.id}`}
      className="rounded-lg border border-border bg-card p-4 transition-all hover:border-primary hover:shadow-md"
    >
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
          <span className="text-lg font-semibold text-primary">
            {org.name.charAt(0).toUpperCase()}
          </span>
        </div>
        <div className="min-w-0 flex-1">
          <h3 className="truncate font-medium">{org.name}</h3>
          <p className="truncate text-sm text-muted-foreground">{org.description}</p>
        </div>
      </div>

      <div className="mt-4 flex items-center gap-3 text-sm text-muted-foreground">
        <span>{org.members} members</span>
        <span>•</span>
        <span>{org.deployments} deployments</span>
        <span className={`ml-auto rounded-full px-2 py-0.5 text-xs ${roleColors[org.role] || ''}`}>
          {org.role}
        </span>
      </div>
    </Link>
  );
}
