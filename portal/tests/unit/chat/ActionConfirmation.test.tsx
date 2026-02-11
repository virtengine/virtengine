/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ActionConfirmation } from '../../../src/components/chat/ActionConfirmation';
import type { ChatAction } from '../../../src/lib/portal-adapter';

const action: ChatAction = {
  id: 'action-1',
  toolName: 'delete-deployments',
  title: 'Stop deployments',
  summary: 'Will stop 2 deployments.',
  payload: {
    kind: 'provider-action',
    deploymentIds: ['dep-1', 'dep-2'],
    action: 'stop',
  },
  destructive: true,
  requiresConfirmation: true,
  impact: {
    count: 2,
    resources: [{ id: 'dep-1' }, { id: 'dep-2' }],
  },
};

describe('ActionConfirmation', () => {
  it('renders action summary and handles confirm/cancel', () => {
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(<ActionConfirmation action={action} onConfirm={onConfirm} onCancel={onCancel} />);

    expect(screen.getByText('Stop deployments')).toBeDefined();
    expect(screen.getByText('Will stop 2 deployments.')).toBeDefined();

    fireEvent.click(screen.getByRole('button', { name: 'Confirm' }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
