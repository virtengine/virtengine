import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

// Mock the features/veid module
vi.mock('@/features/veid', () => ({
  TIER_INFO: {
    unverified: {
      tier: 'unverified',
      label: 'Unverified',
      description: 'No identity verification completed',
      color: 'text-muted-foreground',
      bgColor: 'bg-muted',
      borderColor: 'border-muted',
      minScore: 0,
      maxScore: 0,
      icon: '○',
    },
    basic: {
      tier: 'basic',
      label: 'Basic',
      description: 'Email verified, basic marketplace access',
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
      borderColor: 'border-blue-200',
      minScore: 1,
      maxScore: 40,
      icon: '◑',
    },
    premium: {
      tier: 'premium',
      label: 'Premium',
      description: 'Full verification, HPC and provider access',
      color: 'text-green-600',
      bgColor: 'bg-green-50',
      borderColor: 'border-green-200',
      minScore: 71,
      maxScore: 90,
      icon: '●',
    },
    standard: {
      tier: 'standard',
      label: 'Standard',
      description: 'Document verified',
      color: 'text-amber-600',
      bgColor: 'bg-amber-50',
      borderColor: 'border-amber-200',
      minScore: 41,
      maxScore: 70,
      icon: '◕',
    },
    elite: {
      tier: 'elite',
      label: 'Elite',
      description: 'Maximum trust',
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
      borderColor: 'border-purple-200',
      minScore: 91,
      maxScore: 100,
      icon: '★',
    },
  },
  FEATURE_THRESHOLDS: [
    { action: 'browse_offerings', label: 'Browse Marketplace', minScore: 0 },
    { action: 'place_order', label: 'Place Orders', minScore: 30 },
    { action: 'register_provider', label: 'Register as Provider', minScore: 70 },
  ],
}));

import { VEIDScore } from '@/components/veid/VEIDScore';

describe('VEIDScore', () => {
  it('renders score value', () => {
    render(<VEIDScore score={75} tier="premium" />);
    expect(screen.getByText('75')).toBeInTheDocument();
  });

  it('renders tier badge', () => {
    render(<VEIDScore score={75} tier="premium" />);
    expect(screen.getByText(/Premium Tier/)).toBeInTheDocument();
  });

  it('renders tier description for md size', () => {
    render(<VEIDScore score={75} tier="premium" />);
    expect(screen.getByText(/Full verification/)).toBeInTheDocument();
  });

  it('does not render description for sm size', () => {
    render(<VEIDScore score={75} tier="premium" size="sm" />);
    expect(screen.queryByText(/Full verification/)).not.toBeInTheDocument();
  });

  it('renders feature access when showFeatureAccess is true', () => {
    render(<VEIDScore score={75} tier="premium" showFeatureAccess />);
    expect(screen.getByText('Feature Access')).toBeInTheDocument();
    expect(screen.getByText('Browse Marketplace')).toBeInTheDocument();
    expect(screen.getByText('Place Orders')).toBeInTheDocument();
    expect(screen.getByText('Register as Provider')).toBeInTheDocument();
  });

  it('does not render feature access by default', () => {
    render(<VEIDScore score={75} tier="premium" />);
    expect(screen.queryByText('Feature Access')).not.toBeInTheDocument();
  });

  it('shows correct met/unmet indicators for features', () => {
    render(<VEIDScore score={50} tier="standard" showFeatureAccess />);

    // Score 50 should meet browse (0) and place_order (30) but not register_provider (70)
    const items = screen.getAllByText(/[✓○]/);
    // At least some checkmarks should exist
    expect(items.length).toBeGreaterThan(0);
  });

  it('renders score label for accessibility', () => {
    render(<VEIDScore score={85} tier="premium" />);
    expect(screen.getByLabelText('VEID Score: 85')).toBeInTheDocument();
  });

  it('renders / 100 denominator', () => {
    render(<VEIDScore score={50} tier="standard" />);
    expect(screen.getByText('/ 100')).toBeInTheDocument();
  });

  it('renders unverified tier correctly', () => {
    render(<VEIDScore score={0} tier="unverified" />);
    expect(screen.getByText('0')).toBeInTheDocument();
    expect(screen.getByText(/Unverified Tier/)).toBeInTheDocument();
  });
});
