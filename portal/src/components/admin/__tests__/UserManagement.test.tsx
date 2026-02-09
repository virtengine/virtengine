import { describe, it, expect, beforeEach } from 'vitest';
import { fireEvent, screen } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import AdminUsersPage from '@/app/admin/users/page';
import { useAdminStore } from '@/stores/adminStore';
import i18n from '@/i18n';

const initialState = useAdminStore.getState();

expectTranslations(['User Management', 'Admin roles, VEID oversight, and account operations']);

describe.each(TEST_LOCALES)('AdminUsersPage (%s)', (locale) => {
  beforeEach(async () => {
    useAdminStore.setState(initialState, true);
    await setLocale(locale);
  });

  it('renders user list with VEID status and actions', () => {
    const account = useAdminStore.getState().accounts[0];
    renderWithI18n(<AdminUsersPage />);

    expect(screen.getByText(i18n.t('User Management', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByText(account.displayName)).toBeInTheDocument();
    expect(screen.getAllByText(account.veidStatus).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('button', { name: /flag/i }).length).toBeGreaterThan(0);
  });

  it('filters accounts by search input', () => {
    const [first, second] = useAdminStore.getState().accounts;
    renderWithI18n(<AdminUsersPage />);

    fireEvent.change(screen.getByPlaceholderText('Search accounts...'), {
      target: { value: first.displayName.slice(0, 4) },
    });

    expect(screen.getByText(first.displayName)).toBeInTheDocument();
    expect(screen.queryByText(second.displayName)).not.toBeInTheDocument();
  });
});
