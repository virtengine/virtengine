/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Security settings page – MFA enrollment, challenge flow, and management.
 */

import type { Metadata } from 'next';
import { SecuritySettingsContent } from './SecuritySettingsContent';

export const metadata: Metadata = {
  title: 'Security Settings',
  description: 'Manage your VirtEngine multi-factor authentication and account security',
};

export default function SecuritySettingsPage() {
  return (
    <div className="container max-w-4xl py-8">
      <div className="mb-8">
        <nav className="mb-4 text-sm text-muted-foreground">
          <a href="/account/settings" className="hover:text-foreground">
            Account Settings
          </a>
          <span className="mx-2">/</span>
          <span className="text-foreground">Security</span>
        </nav>
        <h1 className="text-3xl font-bold">Security Settings</h1>
        <p className="mt-2 text-muted-foreground">
          Manage multi-factor authentication, trusted devices, and review security activity.
        </p>
      </div>

      <SecuritySettingsContent />
    </div>
  );
}
