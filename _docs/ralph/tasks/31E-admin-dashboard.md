# Task 31E: Admin Dashboard

**vibe-kanban ID:** `5a8a094a-b2e9-4977-bc91-e0ac2a1da87b`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31E |
| **Title** | feat(admin): Admin dashboard |
| **Priority** | P1 |
| **Wave** | 2 |
| **Estimated LOC** | 5000 |
| **Duration** | 4 weeks |
| **Dependencies** | lib/admin/ exists but incomplete |
| **Blocking** | None |

---

## Problem Statement

Operations team requires an admin dashboard for:
- Network health monitoring
- User management and VEID oversight
- Provider management and lease tracking
- Escrow monitoring and dispute resolution
- System configuration management

Currently, all administrative actions require CLI access and chain queries.

### Current State Analysis

```
lib/admin/                      ✅ Basic skeleton exists
lib/admin/src/components/       ⚠️  Few components
lib/admin/src/pages/            ❌ Minimal pages
deploy/admin/                   ❌ No deployment config
```

---

## Acceptance Criteria

### AC-1: Network Health Dashboard
- [ ] Real-time block height and sync status
- [ ] Validator set overview (active, jailed, tombstoned)
- [ ] Network-wide resource utilization
- [ ] System alerts and warnings
- [ ] Recent governance proposals

### AC-2: User Management Panel
- [ ] User search by address or VEID status
- [ ] VEID record inspection (respecting encryption)
- [ ] Account suspension/flagging capabilities
- [ ] KYC/AML review queue
- [ ] User activity logs

### AC-3: Provider Management
- [ ] Provider registry with status indicators
- [ ] Active lease overview by provider
- [ ] Provider health metrics
- [ ] Manual provider intervention tools
- [ ] Provider verification status

### AC-4: Financial Operations
- [ ] Escrow balance overview
- [ ] Pending withdrawals queue
- [ ] Dispute resolution interface
- [ ] Settlement history
- [ ] Revenue analytics

### AC-5: System Configuration
- [ ] Module parameter viewing/editing (via governance)
- [ ] Feature flag management
- [ ] Maintenance mode controls
- [ ] Audit log viewer

---

## Technical Requirements

### Dashboard Layout

```tsx
// lib/admin/src/app/layout.tsx

import { Sidebar } from '@/components/layout/Sidebar';
import { Header } from '@/components/layout/Header';
import { AdminAuthProvider } from '@/providers/AdminAuthProvider';

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AdminAuthProvider>
      <div className="flex h-screen bg-gray-100">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <Header />
          <main className="flex-1 overflow-y-auto p-6">
            {children}
          </main>
        </div>
      </div>
    </AdminAuthProvider>
  );
}

// lib/admin/src/components/layout/Sidebar.tsx

'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  Users,
  Server,
  Wallet,
  Settings,
  Shield,
  FileText,
  Activity,
} from 'lucide-react';

const navigation = [
  { name: 'Dashboard', href: '/admin', icon: LayoutDashboard },
  { name: 'Users', href: '/admin/users', icon: Users },
  { name: 'Providers', href: '/admin/providers', icon: Server },
  { name: 'Escrow', href: '/admin/escrow', icon: Wallet },
  { name: 'VEID Review', href: '/admin/veid', icon: Shield },
  { name: 'Audit Logs', href: '/admin/audit', icon: FileText },
  { name: 'System', href: '/admin/system', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <div className="w-64 bg-gray-900 text-white">
      <div className="p-6">
        <h1 className="text-xl font-bold">VirtEngine Admin</h1>
      </div>
      <nav className="mt-6">
        {navigation.map((item) => (
          <Link
            key={item.name}
            href={item.href}
            className={`flex items-center px-6 py-3 ${
              pathname === item.href
                ? 'bg-gray-800 border-l-4 border-blue-500'
                : 'hover:bg-gray-800'
            }`}
          >
            <item.icon className="h-5 w-5 mr-3" />
            {item.name}
          </Link>
        ))}
      </nav>
    </div>
  );
}
```

### Network Health Dashboard

```tsx
// lib/admin/src/app/admin/page.tsx

import { Suspense } from 'react';
import { NetworkStats } from '@/components/dashboard/NetworkStats';
import { ValidatorOverview } from '@/components/dashboard/ValidatorOverview';
import { RecentBlocks } from '@/components/dashboard/RecentBlocks';
import { AlertsFeed } from '@/components/dashboard/AlertsFeed';
import { ResourceUtilization } from '@/components/dashboard/ResourceUtilization';

export default function AdminDashboard() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Network Overview</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Suspense fallback={<StatCardSkeleton />}>
          <NetworkStats />
        </Suspense>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Suspense fallback={<CardSkeleton />}>
          <ValidatorOverview />
        </Suspense>
        <Suspense fallback={<CardSkeleton />}>
          <ResourceUtilization />
        </Suspense>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Suspense fallback={<CardSkeleton />}>
          <RecentBlocks />
        </Suspense>
        <Suspense fallback={<CardSkeleton />}>
          <AlertsFeed />
        </Suspense>
      </div>
    </div>
  );
}

// lib/admin/src/components/dashboard/NetworkStats.tsx

'use client';

import { useQuery } from '@tanstack/react-query';
import { fetchNetworkStats } from '@/lib/api/network';
import { StatCard } from '@/components/ui/StatCard';

export function NetworkStats() {
  const { data: stats } = useQuery({
    queryKey: ['networkStats'],
    queryFn: fetchNetworkStats,
    refetchInterval: 5000,
  });

  return (
    <>
      <StatCard
        title="Block Height"
        value={stats?.blockHeight.toLocaleString()}
        trend={stats?.blocksPerMinute}
        trendLabel="blocks/min"
      />
      <StatCard
        title="Active Validators"
        value={`${stats?.activeValidators}/${stats?.totalValidators}`}
        status={stats?.activeValidators === stats?.totalValidators ? 'healthy' : 'warning'}
      />
      <StatCard
        title="Active Leases"
        value={stats?.activeLeases.toLocaleString()}
        trend={stats?.leasesChange24h}
        trendLabel="24h"
      />
      <StatCard
        title="Total Escrow"
        value={`${stats?.totalEscrow} VE`}
        trend={stats?.escrowChange24h}
        trendLabel="24h"
      />
    </>
  );
}
```

### User Management

```tsx
// lib/admin/src/app/admin/users/page.tsx

'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { DataTable } from '@/components/ui/DataTable';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { searchUsers, UserRecord } from '@/lib/api/users';

const columns = [
  {
    header: 'Address',
    accessor: 'address',
    cell: (row: UserRecord) => (
      <code className="text-sm">{row.address}</code>
    ),
  },
  {
    header: 'VEID Status',
    accessor: 'veidStatus',
    cell: (row: UserRecord) => (
      <Badge variant={getVeidBadgeVariant(row.veidStatus)}>
        {row.veidStatus}
      </Badge>
    ),
  },
  {
    header: 'Trust Score',
    accessor: 'trustScore',
    cell: (row: UserRecord) => (
      <span className={getTrustScoreClass(row.trustScore)}>
        {row.trustScore}/100
      </span>
    ),
  },
  {
    header: 'Created',
    accessor: 'createdAt',
    cell: (row: UserRecord) => new Date(row.createdAt).toLocaleDateString(),
  },
  {
    header: 'Actions',
    accessor: 'actions',
    cell: (row: UserRecord) => (
      <div className="flex gap-2">
        <Button size="sm" variant="outline" onClick={() => viewUser(row.address)}>
          View
        </Button>
        <Button size="sm" variant="destructive" disabled={row.flagged}>
          Flag
        </Button>
      </div>
    ),
  },
];

export default function UsersPage() {
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ['users', search, page],
    queryFn: () => searchUsers({ query: search, page, limit: 20 }),
  });

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">User Management</h1>
        <Input
          placeholder="Search by address or VEID..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-80"
        />
      </div>

      <DataTable
        columns={columns}
        data={data?.users || []}
        isLoading={isLoading}
        pagination={{
          page,
          totalPages: data?.totalPages || 1,
          onPageChange: setPage,
        }}
      />
    </div>
  );
}
```

### VEID Review Queue

```tsx
// lib/admin/src/app/admin/veid/page.tsx

'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { VEIDReviewCard } from '@/components/veid/VEIDReviewCard';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { fetchVEIDQueue, approveVEID, rejectVEID } from '@/lib/api/veid';

export default function VEIDReviewPage() {
  const [tab, setTab] = useState<'pending' | 'flagged' | 'recent'>('pending');
  const queryClient = useQueryClient();

  const { data: queue, isLoading } = useQuery({
    queryKey: ['veidQueue', tab],
    queryFn: () => fetchVEIDQueue(tab),
    refetchInterval: 30000,
  });

  const approveMutation = useMutation({
    mutationFn: approveVEID,
    onSuccess: () => queryClient.invalidateQueries(['veidQueue']),
  });

  const rejectMutation = useMutation({
    mutationFn: rejectVEID,
    onSuccess: () => queryClient.invalidateQueries(['veidQueue']),
  });

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">VEID Review Queue</h1>

      <Tabs value={tab} onValueChange={(v) => setTab(v as typeof tab)}>
        <TabsList>
          <TabsTrigger value="pending">
            Pending Review ({queue?.counts?.pending || 0})
          </TabsTrigger>
          <TabsTrigger value="flagged">
            Flagged ({queue?.counts?.flagged || 0})
          </TabsTrigger>
          <TabsTrigger value="recent">Recent Decisions</TabsTrigger>
        </TabsList>

        <TabsContent value="pending" className="mt-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {queue?.items?.map((item) => (
              <VEIDReviewCard
                key={item.id}
                item={item}
                onApprove={() => approveMutation.mutate(item.id)}
                onReject={(reason) => rejectMutation.mutate({ id: item.id, reason })}
              />
            ))}
          </div>
        </TabsContent>

        {/* Other tabs */}
      </Tabs>
    </div>
  );
}
```

### Admin API Routes

```typescript
// lib/admin/src/app/api/admin/users/route.ts

import { NextRequest, NextResponse } from 'next/server';
import { requireAdminAuth } from '@/lib/auth';
import { searchUsersOnChain } from '@/lib/chain';

export async function GET(request: NextRequest) {
  const admin = await requireAdminAuth(request);
  if (!admin) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  const { searchParams } = new URL(request.url);
  const query = searchParams.get('query') || '';
  const page = parseInt(searchParams.get('page') || '1');
  const limit = parseInt(searchParams.get('limit') || '20');

  const users = await searchUsersOnChain({
    query,
    page,
    limit,
  });

  // Audit log the search
  await logAdminAction(admin.address, 'user_search', { query });

  return NextResponse.json(users);
}

// lib/admin/src/app/api/admin/veid/[id]/approve/route.ts

export async function POST(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  const admin = await requireAdminAuth(request);
  if (!admin) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  // Check admin has VEID_REVIEWER role
  const hasRole = await checkAdminRole(admin.address, 'VEID_REVIEWER');
  if (!hasRole) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 });
  }

  const body = await request.json();
  
  // Submit approval transaction
  const result = await submitVEIDApproval(params.id, admin.address, body.notes);

  // Audit log
  await logAdminAction(admin.address, 'veid_approved', { 
    veidId: params.id,
    txHash: result.txHash,
  });

  return NextResponse.json({ success: true, txHash: result.txHash });
}
```

---

## Directory Structure

```
lib/admin/src/
├── app/
│   ├── layout.tsx
│   ├── page.tsx                  # Redirect to /admin
│   ├── admin/
│   │   ├── page.tsx              # Dashboard
│   │   ├── users/
│   │   │   ├── page.tsx          # User list
│   │   │   └── [address]/
│   │   │       └── page.tsx      # User detail
│   │   ├── providers/
│   │   │   ├── page.tsx          # Provider list
│   │   │   └── [id]/
│   │   │       └── page.tsx      # Provider detail
│   │   ├── escrow/
│   │   │   └── page.tsx          # Escrow overview
│   │   ├── veid/
│   │   │   └── page.tsx          # VEID review queue
│   │   ├── audit/
│   │   │   └── page.tsx          # Audit logs
│   │   └── system/
│   │       └── page.tsx          # System config
│   └── api/admin/
│       ├── users/
│       ├── providers/
│       ├── escrow/
│       ├── veid/
│       └── audit/
├── components/
│   ├── layout/
│   │   ├── Sidebar.tsx
│   │   └── Header.tsx
│   ├── dashboard/
│   │   ├── NetworkStats.tsx
│   │   ├── ValidatorOverview.tsx
│   │   └── AlertsFeed.tsx
│   ├── users/
│   ├── providers/
│   ├── veid/
│   └── ui/
│       ├── DataTable.tsx
│       └── StatCard.tsx
├── lib/
│   ├── api/
│   ├── auth.ts
│   └── chain.ts
└── providers/
    └── AdminAuthProvider.tsx
```

---

## Security Requirements

1. **Authentication**: Admin accounts must be on-chain with roles module
2. **Authorization**: Role-based access (ADMIN, VEID_REVIEWER, SUPPORT)
3. **Audit Logging**: Every admin action logged with timestamp and actor
4. **Encryption**: VEID data decrypted only when needed, never cached
5. **Session Management**: Short session timeouts (1 hour inactive)
6. **IP Allowlisting**: Optional IP restrictions for production

---

## Testing Requirements

### Unit Tests
- Role authorization checks
- Data formatting functions
- Pagination logic

### Integration Tests
- API authentication flows
- Chain query integration
- Audit log creation

### E2E Tests
- Admin login flow
- User search and view
- VEID review process
