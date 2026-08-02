import { beforeEach, describe, expect, it } from 'vitest';
import { clearLegacyOrganizationState, useOrganizationStore } from '@/stores/organizationStore';

describe('organizationStore', () => {
  beforeEach(() => {
    window.localStorage.clear();
    useOrganizationStore.setState({ currentOrgId: null });
  });

  it('does not retain a legacy persisted organization ID', () => {
    window.localStorage.setItem(
      've-organization-storage',
      JSON.stringify({ state: { currentOrgId: 'stale-org' }, version: 0 })
    );

    clearLegacyOrganizationState();

    expect(window.localStorage.getItem('ve-organization-storage')).toBeNull();
    expect(useOrganizationStore.getState().currentOrgId).toBeNull();
  });

  it('denies selection without authoritative organization capability', () => {
    const accepted = useOrganizationStore.getState().setCurrentOrg('untrusted-org');

    expect(accepted).toBe(false);
    expect(useOrganizationStore.getState().currentOrgId).toBeNull();
  });

  it('allows explicitly authoritative selection without persisting it', () => {
    const accepted = useOrganizationStore
      .getState()
      .setCurrentOrg('confirmed-org', { organizationsAvailable: true });

    expect(accepted).toBe(true);
    expect(useOrganizationStore.getState().currentOrgId).toBe('confirmed-org');
    expect(window.localStorage.getItem('ve-organization-storage')).toBeNull();
  });
});
