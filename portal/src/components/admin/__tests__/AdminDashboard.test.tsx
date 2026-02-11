import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { screen } from '@testing-library/react';
import AdminDashboardPage from '@/app/admin/page';
import { useAdminStore } from '@/stores/adminStore';
import { AdminLayout } from '@/layouts/AdminLayout';
import i18n from '@/i18n';

const replaceMock = vi.fn();

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: replaceMock,
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => '/',
  useSearchParams: () => new URLSearchParams(),
}));

const initialState = useAdminStore.getState();

expectTranslations([
  'Admin Dashboard',
  'Network health and operational readiness',
  'Resource Utilization',
  'Validator Health',
]);

describe.each(TEST_LOCALES)('AdminDashboardPage (%s)', (locale) => {
  beforeEach(async () => {
    useAdminStore.setState(initialState, true);
    replaceMock.mockClear();
    await setLocale(locale);
  });

  it('renders network stats and node health', () => {
    const { systemHealth } = useAdminStore.getState();
    renderWithI18n(<AdminDashboardPage />);

    expect(screen.getByText(i18n.t('Admin Dashboard', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByText(systemHealth.blockHeight.toLocaleString())).toBeInTheDocument();
    expect(
      screen.getByText(`${systemHealth.activeValidators}/${systemHealth.totalValidators}`)
    ).toBeInTheDocument();
    expect(screen.getByText(i18n.t('Resource Utilization', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByText(i18n.t('Validator Health', { lng: locale }))).toBeInTheDocument();
  });
});

describe('AdminLayout access control', () => {
  beforeEach(() => {
    useAdminStore.setState(initialState, true);
    replaceMock.mockClear();
  });

  it('redirects non-admin users', () => {
    useAdminStore.setState({ currentUserRoles: [] });
    renderWithI18n(
      <AdminLayout>
        <div>Admin Content</div>
      </AdminLayout>
    );

    expect(replaceMock).toHaveBeenCalledWith('/dashboard');
  });
});
