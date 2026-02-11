import { describe, it, expect, beforeEach } from 'vitest';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { fireEvent, screen } from '@testing-library/react';
import AdminAuditPage from '@/app/admin/audit/page';
import { useAdminStore } from '@/stores/adminStore';
import i18n from '@/i18n';

const initialState = useAdminStore.getState();

expectTranslations([
  'Audit Logs',
  'Admin actions and security events',
  'Search by actor or action',
]);

describe.each(TEST_LOCALES)('AdminAuditPage (%s)', (locale) => {
  beforeEach(async () => {
    useAdminStore.setState(initialState, true);
    await setLocale(locale);
  });

  it('renders audit entries and filters by search', () => {
    const [first, , third] = useAdminStore.getState().auditLogs;
    renderWithI18n(<AdminAuditPage />);

    expect(screen.getByText(i18n.t('Audit Logs', { lng: locale }))).toBeInTheDocument();
    expect(screen.getAllByText(first.actor).length).toBeGreaterThan(0);

    fireEvent.change(
      screen.getByPlaceholderText(i18n.t('Search by actor or action', { lng: locale })),
      {
        target: { value: third.actor.split('...')[0] },
      }
    );

    expect(screen.queryAllByText(first.actor).length).toBe(0);
    expect(screen.getByText(third.actor)).toBeInTheDocument();
  });
});
