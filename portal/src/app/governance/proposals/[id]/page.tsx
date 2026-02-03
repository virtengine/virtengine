import type { Metadata } from 'next';
import Link from 'next/link';

type Props = {
  params: Promise<{ id: string }>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  return {
    title: `Proposal #${id}`,
    description: `View governance proposal #${id}`,
  };
}

export default async function ProposalDetailPage({ params }: Props) {
  const { id } = await params;

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link
          href="/governance/proposals"
          className="text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to Proposals
        </Link>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex items-center gap-3">
              <span className="font-mono text-muted-foreground">#{id}</span>
              <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
                Voting
              </span>
            </div>

            <h1 className="mt-4 text-2xl font-bold">Increase Provider Commission Cap</h1>
            
            <div className="mt-4 flex flex-wrap gap-4 text-sm text-muted-foreground">
              <span>Proposed by: virtengine1abc...xyz</span>
              <span>•</span>
              <span>Created: Jan 15, 2024</span>
              <span>•</span>
              <span>Ends: Jan 22, 2024</span>
            </div>

            <div className="mt-6 prose prose-sm dark:prose-invert max-w-none">
              <h2>Summary</h2>
              <p>
                This proposal seeks to increase the maximum provider commission rate from 20% to 25%.
                This change aims to attract more infrastructure providers to the network by offering
                more competitive revenue sharing.
              </p>

              <h2>Motivation</h2>
              <p>
                Current market analysis shows that competing platforms offer commission rates
                between 25-30% for infrastructure providers. By increasing our cap, we can:
              </p>
              <ul>
                <li>Attract enterprise-grade infrastructure providers</li>
                <li>Increase network capacity and reliability</li>
                <li>Remain competitive in the decentralized compute market</li>
              </ul>

              <h2>Specification</h2>
              <p>
                The change involves updating the <code>MaxProviderCommission</code> parameter
                in the market module from <code>0.20</code> to <code>0.25</code>.
              </p>

              <h2>Impact</h2>
              <p>
                This change may result in slightly higher prices for consumers, but is expected
                to be offset by increased supply and competition among providers.
              </p>
            </div>
          </div>

          {/* Discussion */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Discussion</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Join the community discussion on this proposal
            </p>
            <a
              href="https://forum.virtengine.com"
              target="_blank"
              rel="noopener noreferrer"
              className="mt-4 inline-block text-sm text-primary hover:underline"
            >
              View on Forum →
            </a>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Voting */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Cast Your Vote</h2>
            
            <div className="mt-4 space-y-3">
              <button
                type="button"
                className="w-full rounded-lg bg-success px-4 py-3 text-sm font-medium text-white hover:bg-success/90"
              >
                Vote Yes
              </button>
              <button
                type="button"
                className="w-full rounded-lg bg-destructive px-4 py-3 text-sm font-medium text-white hover:bg-destructive/90"
              >
                Vote No
              </button>
              <button
                type="button"
                className="w-full rounded-lg border border-border px-4 py-3 text-sm hover:bg-accent"
              >
                Abstain
              </button>
            </div>

            <p className="mt-4 text-sm text-muted-foreground">
              Your voting power: <span className="font-medium">1,250 VE</span>
            </p>
          </div>

          {/* Results */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Current Results</h2>
            
            <div className="mt-4 space-y-4">
              <VoteBar label="Yes" votes="2,450,000" percentage={68} color="bg-success" />
              <VoteBar label="No" votes="1,150,000" percentage={32} color="bg-destructive" />
            </div>

            <div className="mt-6 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Total Votes</span>
                <span className="font-medium">3,600,000 VE</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Quorum</span>
                <span className="font-medium text-success">Reached (40%)</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Threshold</span>
                <span className="font-medium">50% + 1</span>
              </div>
            </div>
          </div>

          {/* Timeline */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Timeline</h2>
            
            <div className="mt-4 space-y-4">
              <TimelineItem
                title="Voting Ends"
                date="Jan 22, 2024"
                status="upcoming"
              />
              <TimelineItem
                title="Voting Started"
                date="Jan 15, 2024"
                status="completed"
              />
              <TimelineItem
                title="Deposit Period Ended"
                date="Jan 14, 2024"
                status="completed"
              />
              <TimelineItem
                title="Proposal Created"
                date="Jan 8, 2024"
                status="completed"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function VoteBar({
  label,
  votes,
  percentage,
  color,
}: {
  label: string;
  votes: string;
  percentage: number;
  color: string;
}) {
  return (
    <div>
      <div className="flex justify-between text-sm">
        <span>{label}</span>
        <span className="text-muted-foreground">{votes} ({percentage}%)</span>
      </div>
      <div className="mt-1 h-2 rounded-full bg-muted">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${percentage}%` }} />
      </div>
    </div>
  );
}

function TimelineItem({
  title,
  date,
  status,
}: {
  title: string;
  date: string;
  status: 'completed' | 'upcoming';
}) {
  return (
    <div className="flex gap-3">
      <div className="relative flex flex-col items-center">
        <div
          className={`h-3 w-3 rounded-full ${
            status === 'completed' ? 'bg-success' : 'bg-muted'
          }`}
        />
        <div className="flex-1 w-px bg-border" />
      </div>
      <div className="pb-4">
        <div className="text-sm font-medium">{title}</div>
        <div className="text-xs text-muted-foreground">{date}</div>
      </div>
    </div>
  );
}
