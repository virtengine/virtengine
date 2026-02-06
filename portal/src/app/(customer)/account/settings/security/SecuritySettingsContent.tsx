/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Client component wrapper for the security settings page.
 * Initialises MFA store and renders the MFASettings management UI.
 */

'use client';

import { useEffect } from 'react';
import { useMFAStore } from '@/features/mfa';
import { MFASettings } from '@/components/mfa';

export function SecuritySettingsContent() {
  const loadMFAData = useMFAStore((s) => s.loadMFAData);

  useEffect(() => {
    void loadMFAData();
  }, [loadMFAData]);

  return <MFASettings />;
}
