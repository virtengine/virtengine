import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const shellSource = readFileSync(
  resolve(process.cwd(), 'src/components/deployments/ShellTerminal.tsx'),
  'utf8'
);
const deploymentDetailSource = readFileSync(
  resolve(process.cwd(), 'src/app/provider/orders/[id]/DeploymentDetailClient.tsx'),
  'utf8'
);

describe('deployment shell security regression', () => {
  it('does not authorize shell access from a hardcoded or aggregate VEID score', () => {
    expect(deploymentDetailSource).not.toMatch(/veidScore|minShellScore|hasShellAccess/);
    expect(deploymentDetailSource).toContain('<ShellTerminal');
  });

  it('does not read persistent bearer tokens or construct a direct WebSocket', () => {
    expect(shellSource).not.toContain('localStorage');
    expect(shellSource).not.toMatch(/ve_session_token|ve_portal_token/);
    expect(shellSource).not.toMatch(/searchParams\.set\(['"]token['"]/);
    expect(shellSource).not.toMatch(/new WebSocket\s*\(/);
    expect(shellSource).toContain('useDeploymentShell');
  });
});
