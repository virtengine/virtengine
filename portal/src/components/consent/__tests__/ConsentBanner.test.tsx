import { describe, it, expect, beforeEach, vi } from 'vitest';
import { fireEvent, screen } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { ConsentBanner } from '@/components/consent/ConsentBanner';
import i18n from '@/i18n';

expectTranslations([
  'We use cookies to improve the VirtEngine experience.',
  'Accept all',
  'Reject all',
  'Customize',
  'Save preferences',
  'Necessary cookies',
  'Analytics cookies',
  'Marketing cookies',
]);

describe.each(TEST_LOCALES)('ConsentBanner (%s)', (locale) => {
  beforeEach(async () => {
    window.localStorage.clear();
    (window as Window & { gtag?: (...args: unknown[]) => void }).gtag = vi.fn();
    await setLocale(locale);
  });

  it('renders banner actions', () => {
    renderWithI18n(<ConsentBanner />);

    expect(
      screen.getByText(
        i18n.t('We use cookies to improve the VirtEngine experience.', { lng: locale })
      )
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18n.t('Accept all', { lng: locale }) })
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18n.t('Reject all', { lng: locale }) })
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18n.t('Customize', { lng: locale }) })
    ).toBeInTheDocument();
  });

  it('persists consent and gates tracking calls', () => {
    const gtag = (window as Window & { gtag?: (...args: unknown[]) => void }).gtag as ReturnType<
      typeof vi.fn
    >;
    renderWithI18n(<ConsentBanner />);

    expect(gtag).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: i18n.t('Accept all', { lng: locale }) }));

    const stored = window.localStorage.getItem('ve-consent-preferences');
    expect(stored).toContain('"analytics":true');
    expect(gtag).toHaveBeenCalled();
    expect(
      screen.queryByText(
        i18n.t('We use cookies to improve the VirtEngine experience.', { lng: locale })
      )
    ).not.toBeInTheDocument();
  });

  it('allows customization and persists preferences across sessions', () => {
    renderWithI18n(<ConsentBanner />);

    fireEvent.click(screen.getByRole('button', { name: i18n.t('Customize', { lng: locale }) }));
    const analytics = screen.getByLabelText(i18n.t('Analytics cookies', { lng: locale }));
    fireEvent.click(analytics);
    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Save preferences', { lng: locale }) })
    );

    const stored = window.localStorage.getItem('ve-consent-preferences');
    expect(stored).toContain('"analytics":true');

    renderWithI18n(<ConsentBanner />);
    expect(
      screen.queryByText(
        i18n.t('We use cookies to improve the VirtEngine experience.', { lng: locale })
      )
    ).not.toBeInTheDocument();
  });
});
