/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Resource access panel showing connection details, credentials, and console link.
 */

'use client';

import { useState, useCallback } from 'react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { SimpleTooltip } from '@/components/ui/Tooltip';
import type {
  ResourceAccessInfo,
  AccessCredential,
  ApiEndpoint,
} from '@/features/orders/tracking-types';

// =============================================================================
// Main Component
// =============================================================================

interface ResourceAccessProps {
  access: ResourceAccessInfo;
}

export function ResourceAccess({ access }: ResourceAccessProps) {
  if (!access.isProvisioned) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Resource Access</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center py-8 text-center">
            <div className="rounded-full bg-muted p-3">
              <span className="text-2xl" role="img" aria-label="Lock">
                🔒
              </span>
            </div>
            <p className="mt-3 text-sm text-muted-foreground">
              Access credentials will be available once the deployment is running.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Resource Access</CardTitle>
          <Badge variant="success" dot size="sm">
            Provisioned
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Console Link */}
        {access.consoleUrl && (
          <div>
            <a
              href={access.consoleUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
            >
              Open Web Console
              <span aria-hidden="true">↗</span>
            </a>
          </div>
        )}

        {/* Credentials */}
        {access.credentials.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium">Credentials</h4>
            {access.credentials.map((cred) => (
              <CredentialCard key={`${cred.type}-${cred.host}`} credential={cred} />
            ))}
          </div>
        )}

        {/* API Endpoints */}
        {access.endpoints.length > 0 && (
          <div className="space-y-3">
            <h4 className="text-sm font-medium">API Endpoints</h4>
            {access.endpoints.map((endpoint) => (
              <EndpointCard key={endpoint.name} endpoint={endpoint} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// =============================================================================
// Credential Card
// =============================================================================

function CredentialCard({ credential }: { credential: AccessCredential }) {
  const typeLabels: Record<string, string> = {
    ssh: 'SSH',
    api: 'API',
    console: 'Console',
    vpn: 'VPN',
  };

  const typeVariants: Record<string, 'default' | 'info' | 'success' | 'warning'> = {
    ssh: 'default',
    api: 'info',
    console: 'success',
    vpn: 'warning',
  };

  const fields: { label: string; value: string; sensitive?: boolean }[] = [];

  if (credential.host) {
    fields.push({
      label: credential.port ? 'Host:Port' : 'Host',
      value: credential.port ? `${credential.host}:${credential.port}` : credential.host,
    });
  }
  if (credential.username) {
    fields.push({ label: 'Username', value: credential.username });
  }
  if (credential.password) {
    fields.push({ label: 'Password', value: credential.password, sensitive: true });
  }
  if (credential.apiKey) {
    fields.push({ label: 'API Key', value: credential.apiKey, sensitive: true });
  }
  if (credential.url) {
    fields.push({ label: 'URL', value: credential.url });
  }
  if (credential.privateKey) {
    fields.push({ label: 'Private Key', value: credential.privateKey, sensitive: true });
  }

  return (
    <div className="rounded-lg border border-border bg-muted/30 p-3">
      <div className="mb-2 flex items-center gap-2">
        <Badge variant={typeVariants[credential.type] ?? 'default'} size="sm">
          {typeLabels[credential.type] ?? credential.type}
        </Badge>
        <span className="text-sm font-medium">{credential.label}</span>
      </div>
      <div className="space-y-1.5">
        {fields.map((field) => (
          <CredentialField
            key={field.label}
            label={field.label}
            value={field.value}
            sensitive={field.sensitive}
          />
        ))}
      </div>
    </div>
  );
}

// =============================================================================
// Credential Field with Copy
// =============================================================================

function CredentialField({
  label,
  value,
  sensitive = false,
}: {
  label: string;
  value: string;
  sensitive?: boolean;
}) {
  const [copied, setCopied] = useState(false);
  const [revealed, setRevealed] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for non-secure contexts
    }
  }, [value]);

  const displayValue = sensitive && !revealed ? '•'.repeat(Math.min(value.length, 32)) : value;
  const isMultiline = value.includes('\n');

  return (
    <div className="flex items-start gap-2 text-sm">
      <span className="w-20 shrink-0 text-muted-foreground">{label}</span>
      <div className="flex min-w-0 flex-1 items-start gap-1">
        {isMultiline ? (
          <pre className="flex-1 overflow-x-auto whitespace-pre-wrap break-all rounded bg-muted px-2 py-1 font-mono text-xs">
            {displayValue}
          </pre>
        ) : (
          <code className="flex-1 truncate rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
            {displayValue}
          </code>
        )}
        <div className="flex shrink-0 gap-1">
          {sensitive && (
            <SimpleTooltip content={revealed ? 'Hide' : 'Reveal'}>
              <Button
                variant="ghost"
                size="icon-sm"
                onClick={() => setRevealed(!revealed)}
                aria-label={revealed ? 'Hide value' : 'Show value'}
              >
                <span className="text-xs">{revealed ? '🙈' : '👁'}</span>
              </Button>
            </SimpleTooltip>
          )}
          <SimpleTooltip content={copied ? 'Copied!' : 'Copy'}>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={handleCopy}
              aria-label="Copy to clipboard"
            >
              <span className="text-xs">{copied ? '✓' : '📋'}</span>
            </Button>
          </SimpleTooltip>
        </div>
      </div>
    </div>
  );
}

// =============================================================================
// Endpoint Card
// =============================================================================

function EndpointCard({ endpoint }: { endpoint: ApiEndpoint }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(endpoint.url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback
    }
  }, [endpoint.url]);

  return (
    <div className="rounded-lg border border-border bg-muted/30 p-3">
      <div className="flex items-center justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium">{endpoint.name}</span>
            {endpoint.method && (
              <Badge variant="outline" size="sm">
                {endpoint.method}
              </Badge>
            )}
          </div>
          {endpoint.description && (
            <p className="mt-0.5 text-xs text-muted-foreground">{endpoint.description}</p>
          )}
          <code className="mt-1 block truncate text-xs text-muted-foreground">{endpoint.url}</code>
        </div>
        <SimpleTooltip content={copied ? 'Copied!' : 'Copy URL'}>
          <Button variant="ghost" size="icon-sm" onClick={handleCopy} aria-label="Copy URL">
            <span className="text-xs">{copied ? '✓' : '📋'}</span>
          </Button>
        </SimpleTooltip>
      </div>
    </div>
  );
}
