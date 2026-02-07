'use client';

import { useMemo, useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { useAdminStore } from '@/stores/adminStore';
import { formatDate, formatRelativeTime, truncateAddress } from '@/lib/utils';
import type { AdminRole, UserAccount, VEIDStatus } from '@/types/admin';

const roleStyles: Record<AdminRole, string> = {
  operator: 'bg-primary/10 text-primary',
  governance: 'bg-blue-100 text-blue-700',
  validator: 'bg-emerald-100 text-emerald-700',
  support: 'bg-amber-100 text-amber-700',
};

const veidBadge: Record<VEIDStatus, string> = {
  verified: 'bg-emerald-100 text-emerald-700',
  pending: 'bg-amber-100 text-amber-700',
  unverified: 'bg-slate-100 text-slate-600',
  flagged: 'bg-rose-100 text-rose-700',
  rejected: 'bg-rose-100 text-rose-700',
};

export default function AdminUsersPage() {
  const users = useAdminStore((s) => s.users);
  const accounts = useAdminStore((s) => s.accounts);
  const userActivity = useAdminStore((s) => s.userActivity);
  const assignRole = useAdminStore((s) => s.assignRole);
  const revokeRole = useAdminStore((s) => s.revokeRole);
  const toggleAccountFlag = useAdminStore((s) => s.toggleAccountFlag);
  const toggleAccountSuspension = useAdminStore((s) => s.toggleAccountSuspension);
  const updateKycStatus = useAdminStore((s) => s.updateKycStatus);

  const [selectedRole, setSelectedRole] = useState<AdminRole | ''>('');
  const [targetAddress, setTargetAddress] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [selectedAccount, setSelectedAccount] = useState<UserAccount | null>(null);

  const filteredAccounts = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) {
      return accounts;
    }
    return accounts.filter(
      (account) =>
        account.address.toLowerCase().includes(term) ||
        account.displayName.toLowerCase().includes(term) ||
        account.veidStatus.toLowerCase().includes(term)
    );
  }, [accounts, search]);

  const kycQueue = accounts.filter(
    (account) => account.kycStatus === 'in_review' || account.amlStatus !== 'clear'
  );

  const handleAssign = (address: string) => {
    if (selectedRole) {
      assignRole(address, selectedRole);
      setSelectedRole('');
      setTargetAddress(null);
    }
  };

  const activityForSelected = selectedAccount
    ? userActivity.filter((log) => log.address === selectedAccount.address)
    : [];

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">User Management</h1>
        <p className="mt-1 text-muted-foreground">
          Admin roles, VEID oversight, and account operations
        </p>
      </div>

      <Tabs defaultValue="accounts">
        <TabsList>
          <TabsTrigger value="accounts">Accounts</TabsTrigger>
          <TabsTrigger value="admin">Admin Roles</TabsTrigger>
          <TabsTrigger value="kyc">KYC/AML Queue</TabsTrigger>
        </TabsList>

        <TabsContent value="accounts" className="space-y-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-xl font-semibold">User Accounts</h2>
              <p className="text-sm text-muted-foreground">
                Search by address, name, or VEID status
              </p>
            </div>
            <Input
              placeholder="Search accounts..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="sm:w-80"
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Accounts</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Account</TableHead>
                    <TableHead>VEID</TableHead>
                    <TableHead>Trust</TableHead>
                    <TableHead>KYC/AML</TableHead>
                    <TableHead>Last Active</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredAccounts.map((account) => (
                    <TableRow key={account.address}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{account.displayName}</div>
                          <div className="text-xs text-muted-foreground">
                            {truncateAddress(account.address)}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge className={veidBadge[account.veidStatus]}>
                          {account.veidStatus}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span
                          className={
                            account.trustScore >= 85
                              ? 'text-emerald-600'
                              : account.trustScore >= 65
                                ? 'text-amber-600'
                                : 'text-rose-600'
                          }
                        >
                          {account.trustScore}/100
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="text-xs text-muted-foreground">
                          KYC: {account.kycStatus}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          AML: {account.amlStatus}
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {formatRelativeTime(account.lastActive)}
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setSelectedAccount(account)}
                          >
                            View
                          </Button>
                          <Button
                            variant={account.flagged ? 'secondary' : 'destructive'}
                            size="sm"
                            onClick={() => toggleAccountFlag(account.address)}
                          >
                            {account.flagged ? 'Unflag' : 'Flag'}
                          </Button>
                          <Button
                            variant={account.suspended ? 'secondary' : 'outline'}
                            size="sm"
                            onClick={() => toggleAccountSuspension(account.address)}
                          >
                            {account.suspended ? 'Unsuspend' : 'Suspend'}
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          <div className="grid gap-6 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>VEID Record</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4 text-sm">
                {selectedAccount ? (
                  <>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Address</span>
                      <span>{truncateAddress(selectedAccount.address, 12, 6)}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Status</span>
                      <Badge className={veidBadge[selectedAccount.veidStatus]}>
                        {selectedAccount.veidStatus}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Risk Level</span>
                      <span className="capitalize">{selectedAccount.riskLevel}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Created</span>
                      <span>{formatDate(selectedAccount.createdAt)}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Trust Score</span>
                      <span>{selectedAccount.trustScore}/100</span>
                    </div>
                  </>
                ) : (
                  <p className="text-muted-foreground">
                    Select an account to inspect VEID details.
                  </p>
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>User Activity</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                {selectedAccount ? (
                  activityForSelected.length > 0 ? (
                    activityForSelected.map((log) => (
                      <div key={log.id} className="rounded-lg border border-border p-3">
                        <div className="flex items-center justify-between">
                          <span className="font-medium">{log.action}</span>
                          <span className="text-xs text-muted-foreground">
                            {formatRelativeTime(log.timestamp)}
                          </span>
                        </div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          Source: {log.sourceIp}
                        </div>
                      </div>
                    ))
                  ) : (
                    <p className="text-muted-foreground">No recent activity logged.</p>
                  )
                ) : (
                  <p className="text-muted-foreground">Select an account to view activity.</p>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="admin" className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-4">
            {(['operator', 'governance', 'validator', 'support'] as AdminRole[]).map((role) => (
              <Card key={role}>
                <CardContent className="p-4">
                  <div className="text-sm capitalize text-muted-foreground">{role}s</div>
                  <div className="mt-1 text-2xl font-bold">
                    {users.filter((user) => user.roles.includes(role)).length}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Admin Users</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Roles</TableHead>
                    <TableHead>Assigned</TableHead>
                    <TableHead>Last Active</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((user) => (
                    <TableRow key={user.address}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{user.displayName}</div>
                          <div className="text-xs text-muted-foreground">{user.address}</div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {user.roles.map((role) => (
                            <Badge key={role} className={roleStyles[role]}>
                              {role}
                              <button
                                type="button"
                                className="ml-1 text-xs opacity-60 hover:opacity-100"
                                onClick={() => revokeRole(user.address, role)}
                                aria-label={`Remove ${role} role`}
                              >
                                ×
                              </button>
                            </Badge>
                          ))}
                        </div>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {formatDate(user.assignedAt)}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {user.lastActive ? formatRelativeTime(user.lastActive) : 'Never'}
                      </TableCell>
                      <TableCell>
                        {targetAddress === user.address ? (
                          <div className="flex items-center gap-2">
                            <Select
                              value={selectedRole}
                              onValueChange={(value) => setSelectedRole(value as AdminRole)}
                            >
                              <SelectTrigger className="w-32">
                                <SelectValue placeholder="Role" />
                              </SelectTrigger>
                              <SelectContent>
                                {(['operator', 'governance', 'validator', 'support'] as AdminRole[])
                                  .filter((role) => !user.roles.includes(role))
                                  .map((role) => (
                                    <SelectItem key={role} value={role}>
                                      {role}
                                    </SelectItem>
                                  ))}
                              </SelectContent>
                            </Select>
                            <Button
                              variant="outline"
                              size="sm"
                              disabled={!selectedRole}
                              onClick={() => handleAssign(user.address)}
                            >
                              Add
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => {
                                setTargetAddress(null);
                                setSelectedRole('');
                              }}
                            >
                              Cancel
                            </Button>
                          </div>
                        ) : (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setTargetAddress(user.address)}
                          >
                            Add Role
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="kyc" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>KYC/AML Review Queue</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Account</TableHead>
                    <TableHead>KYC Status</TableHead>
                    <TableHead>AML Status</TableHead>
                    <TableHead>Risk</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {kycQueue.map((account) => (
                    <TableRow key={account.address}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{account.displayName}</div>
                          <div className="text-xs text-muted-foreground">
                            {truncateAddress(account.address)}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          className={
                            account.kycStatus === 'approved'
                              ? 'bg-emerald-100 text-emerald-700'
                              : 'bg-amber-100 text-amber-700'
                          }
                        >
                          {account.kycStatus}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge
                          className={
                            account.amlStatus === 'clear'
                              ? 'bg-emerald-100 text-emerald-700'
                              : 'bg-rose-100 text-rose-700'
                          }
                        >
                          {account.amlStatus}
                        </Badge>
                      </TableCell>
                      <TableCell className="capitalize">{account.riskLevel}</TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => updateKycStatus(account.address, 'approved')}
                          >
                            Approve
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            onClick={() => updateKycStatus(account.address, 'rejected')}
                          >
                            Reject
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
