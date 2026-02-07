import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock the portal-adapter
const mockCheckRequirements = vi.fn();
vi.mock('@/lib/portal-adapter', () => ({
  useIdentity: () => ({
    state: {
      status: 'unknown',
      score: null,
      completedScopes: [],
      isLoading: false,
      error: null,
    },
    actions: {
      refresh: vi.fn(),
      checkRequirements: mockCheckRequirements,
    },
  }),
  ScopeRequirements: ({ action }: { action: string }) => (
    <div data-testid="scope-requirements">{action}</div>
  ),
  RemediationGuide: ({ onStartStep }: { remediation: unknown; onStartStep?: () => void }) => (
    <div data-testid="remediation-guide">
      <button type="button" onClick={onStartStep}>
        Start
      </button>
    </div>
  ),
}));

import { IdentityRequirements } from '@/components/identity/IdentityRequirements';

describe('IdentityRequirements', () => {
  beforeEach(() => {
    mockCheckRequirements.mockClear();
  });

  it('renders nothing when requirements are met', () => {
    mockCheckRequirements.mockReturnValue(null);

    const { container } = render(<IdentityRequirements action="place_order" />);
    expect(container.firstChild).toBeNull();
  });

  it('renders verification required card when gating error exists', () => {
    mockCheckRequirements.mockReturnValue({
      action: 'place_order',
      reason: 'Insufficient verification',
      remediation: { steps: [] },
    });

    render(<IdentityRequirements action="place_order" />);
    expect(screen.getByText('Verification Required')).toBeInTheDocument();
  });

  it('renders scope requirements for the given action', () => {
    mockCheckRequirements.mockReturnValue({
      action: 'place_order',
      reason: 'Insufficient verification',
      remediation: { steps: [] },
    });

    render(<IdentityRequirements action="place_order" />);
    expect(screen.getByTestId('scope-requirements')).toHaveTextContent('place_order');
  });

  it('renders remediation guide', () => {
    mockCheckRequirements.mockReturnValue({
      action: 'place_order',
      reason: 'Insufficient verification',
      remediation: { steps: ['verify_email'] },
    });

    render(<IdentityRequirements action="place_order" />);
    expect(screen.getByTestId('remediation-guide')).toBeInTheDocument();
  });

  it('calls onStartVerification when remediation start is clicked', async () => {
    const onStart = vi.fn();
    mockCheckRequirements.mockReturnValue({
      action: 'place_order',
      reason: 'Insufficient verification',
      remediation: { steps: [] },
    });

    render(<IdentityRequirements action="place_order" onStartVerification={onStart} />);

    const startButton = screen.getByText('Start');
    startButton.click();

    expect(onStart).toHaveBeenCalledTimes(1);
  });
});
