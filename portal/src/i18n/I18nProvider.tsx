/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, type ReactNode } from 'react';
import { I18nextProvider } from 'react-i18next';
import i18n, { DEFAULT_LANGUAGE, isRtlLanguage } from './index';

interface I18nProviderProps {
  children: ReactNode;
}

export function I18nProvider({ children }: I18nProviderProps) {
  useEffect(() => {
    const updateDocumentLanguage = (language: string) => {
      const normalized = language.split('-')[0] || DEFAULT_LANGUAGE;
      document.documentElement.lang = normalized;
      document.documentElement.dir = isRtlLanguage(normalized) ? 'rtl' : 'ltr';
    };

    updateDocumentLanguage(i18n.language || DEFAULT_LANGUAGE);

    const handler = (language: string) => updateDocumentLanguage(language);
    i18n.on('languageChanged', handler);

    return () => {
      i18n.off('languageChanged', handler);
    };
  }, []);

  return <I18nextProvider i18n={i18n}>{children}</I18nextProvider>;
}
