'use client';

import Link from 'next/link';
import { useParams } from 'next/navigation';

export default function OrderDetailClient() {
  const params = useParams();
  const id = params.id as string;

  return (
    <div className="container py-8">
      <div className="mb-6">
        <Link href="/orders" className="text-sm text-muted-foreground hover:text-foreground">
          ← Back to Orders
        </Link>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Main Content */}
        <div className="space-y-6 lg:col-span-2">
          <div className="rounded-lg border border-border bg-card p-6">
            <div className="flex items-start justify-between">
              <div>
                <h1 className="text-2xl font-bold">Order #{id}</h1>
                <p className="mt-1 text-muted-foreground">
                  Created on {new Date().toLocaleDateString()}
                </p>
              </div>
              <span className="flex items-center gap-2 rounded-full bg-success/10 px-3 py-1 text-sm text-success">
                <span className="status-dot status-dot-success" />
                Running
              </span>
            </div>

            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <InfoItem label="Resource Type" value="GPU Compute" />
              <InfoItem label="Provider" value="CloudCore" />
              <InfoItem label="Region" value="US-East" />
              <InfoItem label="Duration" value="3 days 4 hours" />
            </div>
          </div>

          {/* Resource Details */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Resource Details</h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-3">
              <ResourceMetric label="CPU Usage" value="42%" />
              <ResourceMetric label="Memory" value="16.4 GB" />
              <ResourceMetric label="GPU Utilization" value="87%" />
            </div>
          </div>

          {/* Timeline */}
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Timeline</h2>
            <div className="mt-4 space-y-4">
              <TimelineEvent title="Deployment Running" time="2 hours ago" status="current" />
              <TimelineEvent title="Deployment Created" time="3 days ago" status="completed" />
              <TimelineEvent title="Order Matched" time="3 days ago" status="completed" />
              <TimelineEvent title="Order Created" time="3 days ago" status="completed" />
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Cost Summary</h2>
            <div className="mt-4 space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Hourly Rate</span>
                <span>$2.50</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Hours Used</span>
                <span>76</span>
              </div>
              <div className="border-t border-border pt-3">
                <div className="flex justify-between font-medium">
                  <span>Total Cost</span>
                  <span>$190.00</span>
                </div>
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-semibold">Actions</h2>
            <div className="mt-4 space-y-3">
              <button
                type="button"
                className="w-full rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
              >
                View Logs
              </button>
              <button
                type="button"
                className="w-full rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
              >
                SSH Access
              </button>
              <button
                type="button"
                className="w-full rounded-lg border border-destructive px-4 py-2 text-sm text-destructive hover:bg-destructive/10"
              >
                Close Order
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-medium">{value}</dd>
    </div>
  );
}

function ResourceMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg bg-muted/50 p-4 text-center">
      <div className="text-2xl font-bold">{value}</div>
      <div className="mt-1 text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

function TimelineEvent({
  title,
  time,
  status,
}: {
  title: string;
  time: string;
  status: 'current' | 'completed' | 'pending';
}) {
  return (
    <div className="flex gap-4">
      <div className="relative flex flex-col items-center">
        <div
          className={`h-3 w-3 rounded-full ${
            status === 'current' ? 'bg-primary' : status === 'completed' ? 'bg-success' : 'bg-muted'
          }`}
        />
        <div className="w-px flex-1 bg-border" />
      </div>
      <div className="pb-4">
        <div className="font-medium">{title}</div>
        <div className="text-sm text-muted-foreground">{time}</div>
      </div>
    </div>
  );
}
