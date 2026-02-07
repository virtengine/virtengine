'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Progress } from '@/components/ui/Progress';
import { Badge } from '@/components/ui/Badge';
import { useAdminStore } from '@/stores/adminStore';
import { formatTokenAmount } from '@/lib/utils';

export default function AdminHealthPage() {
  const health = useAdminStore((s) => s.systemHealth);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold">System Health</h1>
        <p className="mt-1 text-muted-foreground">Chain metrics and network status</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Block Height</div>
            <div className="mt-1 text-2xl font-bold">{health.blockHeight.toLocaleString()}</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Block Time</div>
            <div className="mt-1 text-2xl font-bold">{health.blockTime}s</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">TX Throughput</div>
            <div className="mt-1 text-2xl font-bold">{health.txThroughput} tx/s</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <div className="text-sm text-muted-foreground">Avg Gas Price</div>
            <div className="mt-1 text-2xl font-bold">{health.avgGasPrice} uve</div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Network</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Network Uptime</span>
              <div className="flex items-center gap-2">
                <Progress
                  value={health.networkUptime}
                  className="w-24"
                  size="sm"
                  variant="success"
                />
                <span className="text-sm font-medium">{health.networkUptime}%</span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Active Validators</span>
              <Badge variant="success">
                {health.activeValidators}/{health.totalValidators}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Inflation Rate</span>
              <span className="text-sm font-medium">{health.inflationRate}%</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Staking</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Bonded Tokens</span>
              <span className="text-sm font-medium">
                {formatTokenAmount(health.bondedTokens)} VE
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Community Pool</span>
              <span className="text-sm font-medium">
                {formatTokenAmount(health.communityPool)} VE
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Validator Set Health</span>
              <Badge
                variant={
                  health.activeValidators / health.totalValidators > 0.66 ? 'success' : 'warning'
                }
              >
                {Math.round((health.activeValidators / health.totalValidators) * 100)}% active
              </Badge>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
