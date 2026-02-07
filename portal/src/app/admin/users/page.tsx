'use client';

import { useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
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
import { formatDate, formatRelativeTime } from '@/lib/utils';
import type { AdminRole } from '@/types/admin';

const roleStyles: Record<AdminRole, string> = {
  operator: 'bg-primary/10 text-primary',
  governance: 'bg-blue-100 text-blue-700',
  validator: 'bg-emerald-100 text-emerald-700',
  support: 'bg-amber-100 text-amber-700',
};

export default function AdminUsersPage() {
  const users = useAdminStore((s) => s.users);
  const assignRole = useAdminStore((s) => s.assignRole);
  const revokeRole = useAdminStore((s) => s.revokeRole);
  const [selectedRole, setSelectedRole] = useState<AdminRole | ''>('');
  const [targetAddress, setTargetAddress] = useState<string | null>(null);

  const handleAssign = (address: string) => {
    if (selectedRole) {
      assignRole(address, selectedRole);
      setSelectedRole('');
      setTargetAddress(null);
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">User Management</h1>
        <p className="mt-1 text-muted-foreground">Manage admin roles and user assignments</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-4">
        {(['operator', 'governance', 'validator', 'support'] as AdminRole[]).map((role) => (
          <Card key={role}>
            <CardContent className="p-4">
              <div className="text-sm capitalize text-muted-foreground">{role}s</div>
              <div className="mt-1 text-2xl font-bold">
                {users.filter((u) => u.roles.includes(role)).length}
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
                              .filter((r) => !user.roles.includes(r))
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
    </div>
  );
}
