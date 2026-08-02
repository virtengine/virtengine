import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LivenessChallenge } from './LivenessChallenge';

describe('LivenessChallenge', () => {
  it('remains unavailable and exposes no skip or completion authority', () => {
    render(<LivenessChallenge />);

    expect(screen.getByText('Liveness Check Unavailable')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /completed|skip|begin/i })).not.toBeInTheDocument();
  });
});
