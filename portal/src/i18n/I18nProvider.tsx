/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { ReactNode } from 'react';
import { I18nextProvider } from 'react-i18next';
import i18n from './index';
import { I18nClientSync } from './I18nClientSync';

interface I18nProviderProps {
  children: ReactNode;
}

export function I18nProvider({ children }: I18nProviderProps) {
  return (
    <I18nextProvider i18n={i18n}>
      <I18nClientSync />
      {children}
    </I18nextProvider>
  );
}
