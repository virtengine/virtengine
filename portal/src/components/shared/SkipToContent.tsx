/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useTranslation } from 'react-i18next';

export function SkipToContent() {
  const { t } = useTranslation();

  return (
    <a href="#main-content" className="skip-link">
      {t('Skip to main content')}
    </a>
  );
}
