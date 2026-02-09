import type { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { act } from 'react';
import i18n from '@/i18n';
import { I18nProvider } from '@/i18n/I18nProvider';

export const TEST_LOCALES = ['en', 'de', 'es', 'ja'] as const;

export async function setLocale(locale: (typeof TEST_LOCALES)[number]) {
  await act(async () => {
    await i18n.changeLanguage(locale);
  });
}

export function renderWithI18n(ui: ReactElement) {
  return render(<I18nProvider>{ui}</I18nProvider>);
}

export function expectTranslations(keys: string[]) {
  keys.forEach((key) => {
    TEST_LOCALES.forEach((locale) => {
      const exists = i18n.exists(key, { lng: locale });
      if (!exists) {
        throw new Error(`Missing i18n key "${key}" for locale "${locale}"`);
      }
    });
  });
}
