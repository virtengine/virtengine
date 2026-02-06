# Task 29H: Organization Management (x/group)

**ID:** 29H  
**Title:** feat(portal): Organization management (x/group)  
**Priority:** P1 (High)  
**Wave:** 3 (After 29D)  
**Estimated LOC:** ~2000  
**Dependencies:** 29D (ProviderAPIClient)  
**Blocking:** None  

---

## Problem Statement

Enterprise customers need to manage teams/organizations with:
1. Shared deployments across team members
2. Role-based access control (admin, member, viewer)
3. Organization-level billing aggregation
4. Easy member invitation and management

VirtEngine has `x/group` module from Cosmos SDK but no portal UI for it.

---

## Acceptance Criteria

### AC-1: Organization List Page
- [ ] List all organizations user belongs to
- [ ] Show role in each organization
- [ ] Display member count
- [ ] Show total deployments per org
- [ ] Create new organization button

### AC-2: Organization Detail Page
- [ ] Organization name and description
- [ ] Member list with roles
- [ ] Deployment list for organization
- [ ] Organization settings (admin only)
- [ ] Leave organization option

### AC-3: Member Management
- [ ] Invite new members by address
- [ ] Set member role on invite
- [ ] Change existing member roles
- [ ] Remove members (admin only)
- [ ] View pending invitations

### AC-4: Role-Based Permissions
- [ ] **Admin**: Full access, can manage members
- [ ] **Member**: Can create/manage deployments
- [ ] **Viewer**: Read-only access to deployments
- [ ] Enforce permissions in UI
- [ ] Show permission-appropriate actions

### AC-5: Organization Billing
- [ ] Aggregated usage across org deployments
- [ ] Billing breakdown by member
- [ ] Invoice history for organization
- [ ] Download org invoices as PDF/CSV

### AC-6: Organization Switching
- [ ] Quick org switcher in header
- [ ] Persist selected org in session
- [ ] Filter deployments by org
- [ ] Show current org context

### AC-7: Create Organization Flow
- [ ] Create organization form
- [ ] Set initial metadata (name, description)
- [ ] Submit x/group MsgCreateGroup transaction
- [ ] Confirm creation success
- [ ] Navigate to new organization

---

## Technical Requirements

### Organization Types

```typescript
// lib/portal/src/types/organization.ts

export interface Organization {
  id: string;
  name: string;
  description?: string;
  admin: string;
  totalWeight: string;
  createdAt: Date;
  metadata: OrganizationMetadata;
}

export interface OrganizationMetadata {
  name: string;
  description?: string;
  website?: string;
  logo?: string;
}

export interface OrganizationMember {
  address: string;
  weight: string;
  role: 'admin' | 'member' | 'viewer';
  addedAt: Date;
  metadata?: MemberMetadata;
}

export interface MemberMetadata {
  name?: string;
  email?: string;
}

export interface OrganizationInvite {
  id: string;
  organizationId: string;
  inviterAddress: string;
  inviteeAddress: string;
  role: 'admin' | 'member' | 'viewer';
  createdAt: Date;
  expiresAt: Date;
  status: 'pending' | 'accepted' | 'rejected' | 'expired';
}

export interface CreateOrganizationRequest {
  name: string;
  description?: string;
  initialMembers?: { address: string; role: string }[];
}

export interface InviteMemberRequest {
  address: string;
  role: 'admin' | 'member' | 'viewer';
}
```

### Organization Hooks

```typescript
// lib/portal/src/hooks/useOrganizations.ts

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useWalletContext } from '../components/wallet/WalletProvider';
import { useVirtEngineClient } from './useVirtEngineClient';

export function useOrganizations() {
  const { address } = useWalletContext();
  const client = useVirtEngineClient();

  return useQuery({
    queryKey: ['organizations', address],
    queryFn: async () => {
      // Query x/group for groups where user is a member
      const response = await client.query(
        '/cosmos/group/v1/groups_by_member',
        { address }
      );
      
      return response.groups.map(parseOrganization);
    },
    enabled: !!address,
  });
}

export function useOrganization(orgId: string) {
  const client = useVirtEngineClient();

  return useQuery({
    queryKey: ['organization', orgId],
    queryFn: async () => {
      const [info, members] = await Promise.all([
        client.query(`/cosmos/group/v1/group_info/${orgId}`),
        client.query(`/cosmos/group/v1/group_members/${orgId}`),
      ]);
      
      return {
        ...parseOrganization(info.info),
        members: members.members.map(parseMember),
      };
    },
    enabled: !!orgId,
  });
}

export function useOrganizationMembers(orgId: string) {
  const client = useVirtEngineClient();

  return useQuery({
    queryKey: ['organization-members', orgId],
    queryFn: async () => {
      const response = await client.query(
        `/cosmos/group/v1/group_members/${orgId}`
      );
      return response.members.map(parseMember);
    },
    enabled: !!orgId,
  });
}

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  const client = useVirtEngineClient();
  const { address } = useWalletContext();

  return useMutation({
    mutationFn: async (request: CreateOrganizationRequest) => {
      const metadata = JSON.stringify({
        name: request.name,
        description: request.description,
      });

      const members = [
        { address, weight: '1', metadata: '' }, // Creator is admin
        ...(request.initialMembers || []).map(m => ({
          address: m.address,
          weight: m.role === 'admin' ? '1' : '0',
          metadata: JSON.stringify({ role: m.role }),
        })),
      ];

      const msg = {
        typeUrl: '/cosmos.group.v1.MsgCreateGroup',
        value: {
          admin: address,
          members,
          metadata,
        },
      };

      return client.signAndBroadcast([msg]);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organizations'] });
    },
  });
}

export function useInviteMember(orgId: string) {
  const queryClient = useQueryClient();
  const client = useVirtEngineClient();
  const { address } = useWalletContext();

  return useMutation({
    mutationFn: async (request: InviteMemberRequest) => {
      const msg = {
        typeUrl: '/cosmos.group.v1.MsgUpdateGroupMembers',
        value: {
          admin: address,
          groupId: orgId,
          memberUpdates: [{
            address: request.address,
            weight: request.role === 'viewer' ? '0' : '1',
            metadata: JSON.stringify({ role: request.role }),
          }],
        },
      };

      return client.signAndBroadcast([msg]);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-members', orgId] });
    },
  });
}

export function useRemoveMember(orgId: string) {
  const queryClient = useQueryClient();
  const client = useVirtEngineClient();
  const { address } = useWalletContext();

  return useMutation({
    mutationFn: async (memberAddress: string) => {
      const msg = {
        typeUrl: '/cosmos.group.v1.MsgUpdateGroupMembers',
        value: {
          admin: address,
          groupId: orgId,
          memberUpdates: [{
            address: memberAddress,
            weight: '0', // Weight 0 removes member
            metadata: '',
          }],
        },
      };

      return client.signAndBroadcast([msg]);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['organization-members', orgId] });
    },
  });
}

function parseOrganization(raw: any): Organization {
  const metadata = JSON.parse(raw.metadata || '{}');
  return {
    id: raw.id,
    name: metadata.name || `Organization ${raw.id}`,
    description: metadata.description,
    admin: raw.admin,
    totalWeight: raw.total_weight,
    createdAt: new Date(raw.created_at),
    metadata,
  };
}

function parseMember(raw: any): OrganizationMember {
  const metadata = JSON.parse(raw.member?.metadata || '{}');
  return {
    address: raw.member?.address,
    weight: raw.member?.weight,
    role: metadata.role || (raw.member?.weight === '1' ? 'member' : 'viewer'),
    addedAt: new Date(raw.member?.added_at),
    metadata,
  };
}
```

### Organization Components

```typescript
// lib/portal/src/components/organization/OrganizationList.tsx

import { useOrganizations } from '../../hooks/useOrganizations';
import { OrganizationCard } from './OrganizationCard';
import { CreateOrganizationButton } from './CreateOrganizationButton';

export function OrganizationList() {
  const { data: organizations, isLoading, error } = useOrganizations();

  if (isLoading) {
    return <OrganizationListSkeleton />;
  }

  if (error) {
    return <ErrorAlert message="Failed to load organizations" />;
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">Organizations</h2>
        <CreateOrganizationButton />
      </div>

      {organizations?.length === 0 ? (
        <EmptyState
          title="No organizations"
          description="Create an organization to manage team deployments"
          action={<CreateOrganizationButton />}
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {organizations?.map((org) => (
            <OrganizationCard key={org.id} organization={org} />
          ))}
        </div>
      )}
    </div>
  );
}

// lib/portal/src/components/organization/OrganizationCard.tsx

import Link from 'next/link';
import { Organization } from '../../types/organization';

interface OrganizationCardProps {
  organization: Organization;
}

export function OrganizationCard({ organization }: OrganizationCardProps) {
  return (
    <Link href={`/organizations/${organization.id}`}>
      <div className="border rounded-lg p-4 hover:border-primary transition-colors">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
            <span className="text-lg font-semibold">
              {organization.name.charAt(0).toUpperCase()}
            </span>
          </div>
          <div>
            <h3 className="font-medium">{organization.name}</h3>
            <p className="text-sm text-muted-foreground">
              {organization.description || 'No description'}
            </p>
          </div>
        </div>
        
        <div className="mt-4 flex items-center gap-4 text-sm text-muted-foreground">
          <span>{organization.totalWeight} members</span>
          <span>•</span>
          <span>Created {organization.createdAt.toLocaleDateString()}</span>
        </div>
      </div>
    </Link>
  );
}

// lib/portal/src/components/organization/MemberList.tsx

import { useState } from 'react';
import { useOrganizationMembers, useRemoveMember } from '../../hooks/useOrganizations';
import { OrganizationMember } from '../../types/organization';
import { useWalletContext } from '../wallet/WalletProvider';

interface MemberListProps {
  orgId: string;
  isAdmin: boolean;
}

export function MemberList({ orgId, isAdmin }: MemberListProps) {
  const { data: members, isLoading } = useOrganizationMembers(orgId);
  const removeMember = useRemoveMember(orgId);
  const { address } = useWalletContext();

  const handleRemove = async (memberAddress: string) => {
    if (confirm('Are you sure you want to remove this member?')) {
      await removeMember.mutateAsync(memberAddress);
    }
  };

  if (isLoading) {
    return <MemberListSkeleton />;
  }

  return (
    <div className="space-y-2">
      {members?.map((member) => (
        <div
          key={member.address}
          className="flex items-center justify-between p-3 border rounded-lg"
        >
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center">
              <span className="text-sm">
                {member.address.slice(0, 2).toUpperCase()}
              </span>
            </div>
            <div>
              <p className="font-mono text-sm">
                {member.address.slice(0, 12)}...{member.address.slice(-6)}
              </p>
              <p className="text-xs text-muted-foreground capitalize">
                {member.role}
              </p>
            </div>
          </div>

          {isAdmin && member.address !== address && (
            <button
              onClick={() => handleRemove(member.address)}
              className="text-sm text-destructive hover:underline"
              disabled={removeMember.isPending}
            >
              Remove
            </button>
          )}
        </div>
      ))}
    </div>
  );
}

// lib/portal/src/components/organization/InviteMemberDialog.tsx

import { useState } from 'react';
import { useInviteMember } from '../../hooks/useOrganizations';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../ui/dialog';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select';

interface InviteMemberDialogProps {
  orgId: string;
  open: boolean;
  onClose: () => void;
}

export function InviteMemberDialog({ orgId, open, onClose }: InviteMemberDialogProps) {
  const [address, setAddress] = useState('');
  const [role, setRole] = useState<'admin' | 'member' | 'viewer'>('member');
  const inviteMember = useInviteMember(orgId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    await inviteMember.mutateAsync({ address, role });
    onClose();
    setAddress('');
    setRole('member');
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite Member</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium">Wallet Address</label>
            <Input
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="virtengine1..."
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium">Role</label>
            <Select value={role} onValueChange={(v) => setRole(v as any)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="admin">Admin - Full access</SelectItem>
                <SelectItem value="member">Member - Can manage deployments</SelectItem>
                <SelectItem value="viewer">Viewer - Read-only access</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={inviteMember.isPending}>
              {inviteMember.isPending ? 'Inviting...' : 'Invite'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// lib/portal/src/components/organization/OrganizationSwitcher.tsx

import { useState } from 'react';
import { useOrganizations } from '../../hooks/useOrganizations';
import { useOrganizationStore } from '../../stores/organization';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';
import { Button } from '../ui/button';
import { ChevronDown, Building2, Plus } from 'lucide-react';

export function OrganizationSwitcher() {
  const { data: organizations } = useOrganizations();
  const { currentOrgId, setCurrentOrg } = useOrganizationStore();
  
  const currentOrg = organizations?.find((o) => o.id === currentOrgId);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" className="w-[200px] justify-between">
          <div className="flex items-center gap-2">
            <Building2 className="h-4 w-4" />
            <span className="truncate">
              {currentOrg?.name || 'Personal'}
            </span>
          </div>
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      
      <DropdownMenuContent className="w-[200px]">
        <DropdownMenuItem onClick={() => setCurrentOrg(null)}>
          <Building2 className="h-4 w-4 mr-2" />
          Personal
        </DropdownMenuItem>
        
        <DropdownMenuSeparator />
        
        {organizations?.map((org) => (
          <DropdownMenuItem
            key={org.id}
            onClick={() => setCurrentOrg(org.id)}
          >
            <Building2 className="h-4 w-4 mr-2" />
            {org.name}
          </DropdownMenuItem>
        ))}
        
        <DropdownMenuSeparator />
        
        <DropdownMenuItem onClick={() => {}}>
          <Plus className="h-4 w-4 mr-2" />
          Create Organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
```

### Organization State Store

```typescript
// lib/portal/src/stores/organization.ts

import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface OrganizationState {
  currentOrgId: string | null;
  setCurrentOrg: (orgId: string | null) => void;
}

export const useOrganizationStore = create<OrganizationState>()(
  persist(
    (set) => ({
      currentOrgId: null,
      setCurrentOrg: (orgId) => set({ currentOrgId: orgId }),
    }),
    {
      name: 'organization-storage',
    }
  )
);
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/types/organization.ts` | Organization types | 80 |
| `lib/portal/src/hooks/useOrganizations.ts` | Organization hooks | 200 |
| `lib/portal/src/stores/organization.ts` | Zustand store | 30 |
| `lib/portal/src/components/organization/OrganizationList.tsx` | List page | 100 |
| `lib/portal/src/components/organization/OrganizationCard.tsx` | Card component | 60 |
| `lib/portal/src/components/organization/OrganizationDetail.tsx` | Detail page | 150 |
| `lib/portal/src/components/organization/MemberList.tsx` | Member list | 100 |
| `lib/portal/src/components/organization/InviteMemberDialog.tsx` | Invite dialog | 100 |
| `lib/portal/src/components/organization/CreateOrganizationDialog.tsx` | Create dialog | 120 |
| `lib/portal/src/components/organization/OrganizationSwitcher.tsx` | Header switcher | 80 |
| `lib/portal/src/components/organization/OrganizationBilling.tsx` | Billing view | 150 |
| `portal/src/app/organizations/page.tsx` | List page | 50 |
| `portal/src/app/organizations/[id]/page.tsx` | Detail page | 80 |

**Total: ~1300 lines**

---

## Validation Checklist

- [ ] Can create new organization
- [ ] Can view organization list
- [ ] Can view organization details
- [ ] Can invite members
- [ ] Can remove members (admin)
- [ ] Role permissions enforced
- [ ] Organization switcher works
- [ ] Billing aggregation shows correctly

---

## Vibe-Kanban Task ID

`05b55541-1e44-417c-b08c-7ac2ddfe034d`
