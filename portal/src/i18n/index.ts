/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import en from './locales/en.json';
import es from './locales/es.json';
import de from './locales/de.json';
import ja from './locales/ja.json';

export const DEFAULT_LANGUAGE = 'en';
export const SUPPORTED_LANGUAGES = ['en', 'es', 'de', 'ja'] as const;
export const RTL_LANGUAGES = ['ar', 'he'] as const;

export function isRtlLanguage(language: string): boolean {
  return RTL_LANGUAGES.includes(language as (typeof RTL_LANGUAGES)[number]);
}

const resources = {
  en: { translation: en },
  es: { translation: es },
  de: { translation: de },
  ja: { translation: ja },
};

if (!i18n.isInitialized) {
  if (typeof window !== 'undefined') {
    i18n.use(LanguageDetector);
  }

  void i18n.use(initReactI18next).init({
    resources,
    lng: DEFAULT_LANGUAGE,
    fallbackLng: DEFAULT_LANGUAGE,
    supportedLngs: [...SUPPORTED_LANGUAGES, ...RTL_LANGUAGES],
    keySeparator: false,
    nsSeparator: false,
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
    },
  });
}

export default i18n;
