'use client';

import { useMFA, MFAPolicyConfig, TrustedBrowserManager, MFAAuditLog } from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs';

interface MFASettingsProps {
  className?: string;
}

/**
 * MFA Settings Component
 * Manage MFA configuration, trusted browsers, and view audit log
 */
export function MFASettings({ className }: MFASettingsProps) {
  const { state } = useMFA();

  if (state.isLoading) {
    return (
      <div className={cn('animate-pulse rounded-lg bg-muted p-6', className)}>
        <div className="h-6 w-48 rounded bg-muted-foreground/20" />
        <div className="mt-4 h-64 w-full rounded bg-muted-foreground/20" />
      </div>
    );
  }

  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle>Two-Factor Authentication Settings</CardTitle>
        <CardDescription>Manage your security settings and trusted devices</CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="factors">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="factors">Factors</TabsTrigger>
            <TabsTrigger value="browsers">Trusted Browsers</TabsTrigger>
            <TabsTrigger value="audit">Audit Log</TabsTrigger>
          </TabsList>
          <TabsContent value="factors" className="mt-4">
            <MFAPolicyConfig currentPolicy={state.policy} enrolledFactors={state.enrolledFactors} />
          </TabsContent>
          <TabsContent value="browsers" className="mt-4">
            <TrustedBrowserManager trustedBrowsers={state.trustedBrowsers} />
          </TabsContent>
          <TabsContent value="audit" className="mt-4">
            <MFAAuditLog />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
