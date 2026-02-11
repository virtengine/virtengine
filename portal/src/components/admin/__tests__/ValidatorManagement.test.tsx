import { describe, it, expect, beforeEach } from 'vitest';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { screen } from '@testing-library/react';
import AdminValidatorsPage from '@/app/admin/validators/page';
import { useAdminStore } from '@/stores/adminStore';
import i18n from '@/i18n';

const initialState = useAdminStore.getState();

const mockValidators = [
  {
    operatorAddress: 'vevaloper1abc',
    moniker: 'TestValidator',
    status: 'active' as const,
    tokens: '1000000',
    delegatorShares: '1000000',
    commission: 0.1,
    uptime: 99.5,
    missedBlocks: 3,
    slashingEvents: [],
  },
  {
    operatorAddress: 'vevaloper2def',
    moniker: 'JailedValidator',
    status: 'jailed' as const,
    tokens: '500000',
    delegatorShares: '500000',
    commission: 0.05,
    uptime: 80.0,
    missedBlocks: 100,
    jailedUntil: new Date('2025-12-01'),
    slashingEvents: [
      {
        id: 'slash-1',
        validatorAddress: 'vevaloper2def',
        reason: 'downtime' as const,
        slashedAmount: '1000',
        blockHeight: 100000,
        timestamp: new Date('2025-06-01'),
      },
    ],
  },
];

expectTranslations([
  'Validators',
  'Monitor validator set status and slashing events',
  'Validator Set',
]);

describe.each(TEST_LOCALES)('AdminValidatorsPage (%s)', (locale) => {
  beforeEach(async () => {
    useAdminStore.setState({ ...initialState, validators: mockValidators }, true);
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
