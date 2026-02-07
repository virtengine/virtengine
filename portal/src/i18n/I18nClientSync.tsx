/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect } from 'react';
import i18n, { DEFAULT_LANGUAGE, isRtlLanguage } from './index';

export function I18nClientSync() {
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

  return null;
}
