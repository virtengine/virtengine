import { describe, it, expect, beforeEach } from 'vitest';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { screen } from '@testing-library/react';
import AdminValidatorsPage from '@/app/admin/validators/page';
import { useAdminStore } from '@/stores/adminStore';
import i18n from '@/i18n';

const initialState = useAdminStore.getState();

expectTranslations([
  'Validators',
  'Monitor validator set status and slashing events',
  'Validator Set',
]);

describe.each(TEST_LOCALES)('AdminValidatorsPage (%s)', (locale) => {
  beforeEach(async () => {
    useAdminStore.setState(initialState, true);
    await setLocale(locale);
  });

  it('renders validator list and slashing info', () => {
    const validators = useAdminStore.getState().validators;
    const jailed = validators.find((v) => v.status === 'jailed');
    renderWithI18n(<AdminValidatorsPage />);

    expect(screen.getByText(i18n.t('Validators', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByText(validators[0].moniker)).toBeInTheDocument();
    if (jailed) {
      expect(screen.getByText(jailed.moniker)).toBeInTheDocument();
    }
    expect(screen.getByText(i18n.t('Validator Set', { lng: locale }))).toBeInTheDocument();
  });
});
