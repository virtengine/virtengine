import { describe, it, expect, beforeEach } from 'vitest';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { screen } from '@testing-library/react';
import AdminSystemPage from '@/app/admin/system/page';
import { useAdminStore } from '@/stores/adminStore';
import i18n from '@/i18n';

const initialState = useAdminStore.getState();

const mockModuleParams = [
  {
    module: 'market',
    key: 'bid_duration',
    value: '20s',
    description: 'Duration for bid placement',
  },
  { module: 'escrow', key: 'payment_period', value: '30s', description: 'Escrow payment period' },
];

expectTranslations([
  'System Configuration',
  'Module parameters, feature flags, and maintenance controls',
  'Module Parameters',
]);

describe.each(TEST_LOCALES)('AdminSystemPage (%s)', (locale) => {
  beforeEach(async () => {
    useAdminStore.setState({ ...initialState, moduleParams: mockModuleParams }, true);
    await setLocale(locale);
  });

  it('renders module params with proposal action', () => {
    const param = useAdminStore.getState().moduleParams[0];
    renderWithI18n(<AdminSystemPage />);

    expect(screen.getByText(i18n.t('System Configuration', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByText(param.module)).toBeInTheDocument();
    expect(screen.getByText(param.key)).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /propose update/i }).length).toBeGreaterThan(0);
  });
});
