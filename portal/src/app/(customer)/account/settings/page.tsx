import type { Metadata } from 'next';
import { accountLink, validatorLink } from '@/lib/explorer';
import { NotificationPreferencesPanel } from '@/components/notifications/NotificationPreferences';

export const metadata: Metadata = {
  title: 'Account Settings',
  description: 'Manage your VirtEngine account profile, security, and access',
};

const connectedWallets = [
  { name: 'Keplr', address: 've1k5p...9x2m', status: 'Connected', default: true },
  { name: 'Leap', address: 've1n7q...2d8c', status: 'Connected', default: false },
];

const activeSessions = [
  {
    device: 'MacBook Pro',
    location: 'San Francisco, CA',
    lastActive: '2 minutes ago',
    current: true,
  },
  { device: 'Windows Desktop', location: 'Austin, TX', lastActive: '3 days ago', current: false },
  { device: 'iPhone 15 Pro', location: 'Chicago, IL', lastActive: '8 days ago', current: false },
];

const apiKeys = [
  { name: 'Analytics Bot', prefix: 've_live_93f2', lastUsed: 'Today, 09:41', status: 'Active' },
  { name: 'CI Pipeline', prefix: 've_live_11ac', lastUsed: 'Yesterday, 18:12', status: 'Active' },
  { name: 'Legacy Integrations', prefix: 've_live_73bd', lastUsed: 'Never', status: 'Paused' },
];

const delegates = [
  {
    name: 'Ops Lead',
    address: 'virtenginevaloper1q8n...4k9l',
    scope: 'Marketplace + Orders',
    limit: '$5,000 / mo',
  },
  {
    name: 'Billing Partner',
    address: 'virtenginevaloper1d3h...8s1w',
    scope: 'Invoices only',
    limit: '$1,000 / mo',
  },
];

export default function AccountSettingsPage() {
  return (
    <div className="container py-8">
      <div className="mb-8 flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="text-3xl font-bold">Account Settings</h1>
          <p className="mt-2 text-muted-foreground">
            Control how your VirtEngine profile, wallets, and access tokens are managed.
          </p>
        </div>
        <div className="rounded-lg border border-border bg-muted/40 px-4 py-3 text-sm text-muted-foreground">
          Last synced: 3 minutes ago
        </div>
      </div>

      <div className="grid gap-8 lg:grid-cols-[240px_1fr]">
        <aside className="hidden lg:block">
          <div className="sticky top-6 space-y-2 rounded-xl border border-border bg-card p-4">
            <SectionLink href="#profile" label="Profile" />
            <SectionLink href="#wallets" label="Wallet" />
            <SectionLink href="#security" label="Security" />
            <SectionLink href="#notifications" label="Notifications" />
            <SectionLink href="#api-keys" label="API Keys" />
            <SectionLink href="#delegation" label="Delegation" />
          </div>
        </aside>

        <div className="space-y-8">
          <section
            id="profile"
            className="scroll-mt-24 rounded-xl border border-border bg-card p-6"
          >
            <div className="mb-6 flex items-start justify-between">
              <div>
                <h2 className="text-xl font-semibold">Profile</h2>
                <p className="text-sm text-muted-foreground">
                  Manage your public presence and encrypted profile data.
                </p>
              </div>
              <button
                type="button"
                className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
              >
                View public profile
              </button>
            </div>

            <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
              <div className="space-y-4">
                <div className="rounded-lg border border-border bg-muted/40 p-4">
                  <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                    On-chain profile
                  </h3>
                  <div className="mt-3 grid gap-3 text-sm">
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">VEID</span>
                      <span className="font-medium">veid_7x1a9q</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground">Roles</span>
                      <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                        Customer
                      </span>
                    </div>
                  </div>
                </div>

                <div className="grid gap-4">
                  <div>
                    <label className="text-sm font-medium" htmlFor="display-name">
                      Display name
                    </label>
                    <input
                      id="display-name"
                      type="text"
                      defaultValue="Avery Chen"
                      className="mt-2 w-full rounded-lg border border-border bg-background px-4 py-2 text-sm"
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium" htmlFor="avatar-upload">
                      Avatar (optional)
                    </label>
                    <div className="mt-2 flex items-center gap-4 rounded-lg border border-dashed border-border bg-muted/30 px-4 py-3">
                      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                        AC
                      </div>
                      <div className="flex-1">
                        <p className="text-sm font-medium">Upload a PNG or JPG</p>
                        <p className="text-xs text-muted-foreground">
                          Up to 2MB. Square crop recommended.
                        </p>
                      </div>
                      <button
                        type="button"
                        className="rounded-lg border border-border px-3 py-2 text-xs hover:bg-accent"
                      >
                        Choose file
                      </button>
                      <input id="avatar-upload" type="file" className="sr-only" />
                    </div>
                  </div>
                </div>
              </div>

              <div className="rounded-lg border border-border bg-muted/40 p-4 text-sm">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  Profile storage
                </h3>
                <p className="mt-3 text-muted-foreground">
                  Minimal identity metadata (VEID + role proofs) is stored on-chain. Extended
                  profile data is encrypted and stored off-chain for faster updates.
                </p>
                <div className="mt-4 space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">Encryption</span>
                    <span className="font-medium text-success">Enabled</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">Last key rotation</span>
                    <span className="font-medium">Jan 12, 2026</span>
                  </div>
                  <button
                    type="button"
                    className="mt-2 w-full rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
                  >
                    Save profile changes
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section
            id="wallets"
            className="scroll-mt-24 rounded-xl border border-border bg-card p-6"
          >
            <div className="mb-6 flex items-start justify-between">
              <div>
                <h2 className="text-xl font-semibold">Wallet</h2>
                <p className="text-sm text-muted-foreground">
                  Choose a default wallet for transactions and staking.
                </p>
              </div>
              <button
                type="button"
                className="rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
              >
                Connect wallet
              </button>
            </div>

            <div className="space-y-4">
              {connectedWallets.map((wallet) => (
                <div
                  key={wallet.name}
                  className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-4 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{wallet.name}</span>
                      {wallet.default && (
                        <span className="rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
                          Default
                        </span>
                      )}
                    </div>
                    <a
                      className="text-xs font-medium text-primary hover:underline"
                      href={accountLink(wallet.address)}
                      rel="noopener noreferrer"
                      target="_blank"
                    >
                      {wallet.address}
                    </a>
                  </div>
                  <div className="flex items-center gap-3 text-sm">
                    <span className="text-success">{wallet.status}</span>
                    {!wallet.default && (
                      <button
                        type="button"
                        className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                      >
                        Make default
                      </button>
                    )}
                    <button
                      type="button"
                      className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                    >
                      Disconnect
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section
            id="security"
            className="scroll-mt-24 rounded-xl border border-border bg-card p-6"
          >
            <div className="mb-6">
              <h2 className="text-xl font-semibold">Security</h2>
              <p className="text-sm text-muted-foreground">
                Review active sessions and manage MFA or password credentials.
              </p>
            </div>

            <div className="grid gap-6 lg:grid-cols-[1fr_1fr]">
              <div className="rounded-lg border border-border bg-muted/30 p-4">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  MFA
                </h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  Multi-factor authentication is currently enabled for this account.
                </p>
                <div className="mt-4 flex items-center justify-between">
                  <span className="text-sm font-medium text-success">Authenticator app</span>
                  <a
                    href="/account/settings/security"
                    className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                  >
                    Manage MFA
                  </a>
                </div>
              </div>

              <div className="rounded-lg border border-border bg-muted/30 p-4">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  Password
                </h3>
                <p className="mt-2 text-sm text-muted-foreground">
                  Custodial accounts can update their sign-in password at any time.
                </p>
                <button
                  type="button"
                  className="mt-4 rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                >
                  Change password
                </button>
              </div>
            </div>

            <div className="mt-6 space-y-4">
              <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                Active sessions
              </h3>
              {activeSessions.map((session) => (
                <div
                  key={`${session.device}-${session.location}`}
                  className="flex flex-col gap-3 rounded-lg border border-border bg-muted/30 p-4 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{session.device}</span>
                      {session.current && (
                        <span className="rounded-full bg-success/10 px-2 py-1 text-xs font-medium text-success">
                          Current
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {session.location} • {session.lastActive}
                    </p>
                  </div>
                  <button
                    type="button"
                    className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                    disabled={session.current}
                  >
                    Revoke session
                  </button>
                </div>
              ))}
            </div>
          </section>

          <section
            id="notifications"
            className="scroll-mt-24 rounded-xl border border-border bg-card p-6"
          >
            <div className="mb-6">
              <h2 className="text-xl font-semibold">Notifications</h2>
              <p className="text-sm text-muted-foreground">
                Configure alerts for billing, deployments, and security events.
              </p>
            </div>
            <NotificationPreferencesPanel />
          </section>

          <section
            id="api-keys"
            className="scroll-mt-24 rounded-xl border border-border bg-card p-6"
          >
            <div className="mb-6 flex items-start justify-between">
              <div>
                <h2 className="text-xl font-semibold">API Keys</h2>
                <p className="text-sm text-muted-foreground">
                  Generate and manage keys for SDK and automation access.
                </p>
              </div>
              <button
                type="button"
                className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent"
              >
                View docs
              </button>
            </div>

            <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
              <div className="space-y-4 rounded-lg border border-border bg-muted/30 p-4">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  Create new key
                </h3>
                <div>
                  <label className="text-sm font-medium" htmlFor="api-key-name">
                    Key name
                  </label>
                  <input
                    id="api-key-name"
                    type="text"
                    placeholder="e.g. Data pipeline"
                    className="mt-2 w-full rounded-lg border border-border bg-background px-4 py-2 text-sm"
                  />
                </div>
                <div>
                  <p className="text-sm font-medium">Permissions</p>
                  <div className="mt-2 space-y-2 text-sm text-muted-foreground">
                    <label className="flex items-center gap-2">
                      <input type="checkbox" defaultChecked className="rounded border-border" />
                      Read-only access
                    </label>
                    <label className="flex items-center gap-2">
                      <input type="checkbox" className="rounded border-border" />
                      Write & deployment actions
                    </label>
                    <label className="flex items-center gap-2">
                      <input type="checkbox" className="rounded border-border" />
                      Billing & usage exports
                    </label>
                  </div>
                </div>
                <button
                  type="button"
                  className="w-full rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
                >
                  Generate API key
                </button>
                <p className="text-xs text-muted-foreground">
                  New keys are shown once. Store them in a secure vault.
                </p>
              </div>

              <div className="space-y-3">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  Active keys
                </h3>
                {apiKeys.map((key) => (
                  <div key={key.prefix} className="rounded-lg border border-border bg-muted/30 p-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium">{key.name}</p>
                        <p className="text-xs text-muted-foreground">{key.prefix}</p>
                      </div>
                      <span
                        className={`rounded-full px-2 py-1 text-xs font-medium ${
                          key.status === 'Active'
                            ? 'bg-success/10 text-success'
                            : 'bg-muted text-muted-foreground'
                        }`}
                      >
                        {key.status}
                      </span>
                    </div>
                    <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
                      <span>Last used: {key.lastUsed}</span>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          className="rounded-lg border border-border px-2 py-1 hover:bg-accent"
                        >
                          Rotate
                        </button>
                        <button
                          type="button"
                          className="rounded-lg border border-border px-2 py-1 hover:bg-accent"
                        >
                          Revoke
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </section>

          <section
            id="delegation"
            className="scroll-mt-24 rounded-xl border border-border bg-card p-6"
          >
            <div className="mb-6">
              <h2 className="text-xl font-semibold">Delegation</h2>
              <p className="text-sm text-muted-foreground">
                Delegate limited access to teammates or partners with scoped permissions.
              </p>
            </div>

            <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
              <div className="space-y-4">
                {delegates.map((delegate) => (
                  <div
                    key={delegate.address}
                    className="rounded-lg border border-border bg-muted/30 p-4"
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium">{delegate.name}</p>
                        <a
                          className="text-xs font-medium text-primary hover:underline"
                          href={validatorLink(delegate.address)}
                          rel="noopener noreferrer"
                          target="_blank"
                        >
                          {delegate.address}
                        </a>
                      </div>
                      <button
                        type="button"
                        className="rounded-lg border border-border px-3 py-1.5 text-xs hover:bg-accent"
                      >
                        Edit
                      </button>
                    </div>
                    <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                      <span>Scope: {delegate.scope}</span>
                      <span>Spend limit: {delegate.limit}</span>
                    </div>
                  </div>
                ))}
              </div>

              <div className="rounded-lg border border-border bg-muted/30 p-4">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
                  Add delegate
                </h3>
                <div className="mt-4 space-y-3">
                  <div>
                    <label className="text-sm font-medium" htmlFor="delegate-address">
                      Wallet address
                    </label>
                    <input
                      id="delegate-address"
                      type="text"
                      placeholder="ve1..."
                      className="mt-2 w-full rounded-lg border border-border bg-background px-4 py-2 text-sm"
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium" htmlFor="delegate-scope">
                      Permission scope
                    </label>
                    <select
                      id="delegate-scope"
                      className="mt-2 w-full rounded-lg border border-border bg-background px-4 py-2 text-sm"
                    >
                      <option>Marketplace + Orders</option>
                      <option>Billing only</option>
                      <option>Read-only analytics</option>
                      <option>Provider management</option>
                    </select>
                  </div>
                  <button
                    type="button"
                    className="w-full rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
                  >
                    Send delegation invite
                  </button>
                </div>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}

function SectionLink({ href, label }: { href: string; label: string }) {
  return (
    <a
      href={href}
      className="flex items-center justify-between rounded-lg px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
    >
      <span>{label}</span>
      <span className="text-xs">/</span>
    </a>
  );
}
