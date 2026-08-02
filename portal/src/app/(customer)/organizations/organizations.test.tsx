import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import OrganizationsPage from './page';
import OrganizationDetailPage from './[id]/page';

const placeholderText = /Acme Corp|Dev Team|Research Lab|members|deployments|Admin|Viewer/i;

describe('organization routes', () => {
  it('renders a typed unavailable state without placeholder data or creation controls', () => {
    const { container } = render(<OrganizationsPage />);

    expect(screen.getByRole('status')).toHaveAttribute('data-error-code', 'feature_unavailable');
    expect(
      screen.getByText('Organization support is unavailable for this deployment.')
    ).toBeVisible();
    expect(screen.queryByText(placeholderText)).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
    expect(container.querySelector('a[href^="/organizations/"]')).not.toBeInTheDocument();
  });

  it('keeps arbitrary detail routes typed unavailable without echoing the ID', async () => {
    const page = await OrganizationDetailPage({
      params: Promise.resolve({ id: 'legacy-fake-org' }),
    });
    render(page);

    expect(screen.getByRole('status')).toHaveAttribute('data-error-code', 'feature_unavailable');
    expect(
      screen.getByText('Organization support is unavailable for this deployment.')
    ).toBeVisible();
    expect(screen.queryByText(/legacy-fake-org/i)).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });
});
