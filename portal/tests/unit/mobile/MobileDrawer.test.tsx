import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import { MobileDrawer } from '@/components/layout/MobileDrawer';

describe('MobileDrawer', () => {
  it('renders nothing when closed', () => {
    const { container } = render(
      <MobileDrawer open={false} onClose={vi.fn()}>
        <div>Content</div>
      </MobileDrawer>
    );
    expect(container.innerHTML).toBe('');
  });

  it('renders children when open', () => {
    render(
      <MobileDrawer open={true} onClose={vi.fn()}>
        <div>Drawer Content</div>
      </MobileDrawer>
    );
    expect(screen.getByText('Drawer Content')).toBeInTheDocument();
  });

  it('renders a dialog with modal semantics', () => {
    render(
      <MobileDrawer open={true} onClose={vi.fn()} title="Test Menu">
        <div>Content</div>
      </MobileDrawer>
    );
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
  });

  it('displays the title when provided', () => {
    render(
      <MobileDrawer open={true} onClose={vi.fn()} title="Navigation">
        <div>Content</div>
      </MobileDrawer>
    );
    expect(screen.getByText('Navigation')).toBeInTheDocument();
  });

  it('calls onClose when close button is clicked', () => {
    const onClose = vi.fn();
    render(
      <MobileDrawer open={true} onClose={onClose}>
        <div>Content</div>
      </MobileDrawer>
    );

    const closeButton = screen.getByLabelText('Close menu');
    fireEvent.click(closeButton);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when backdrop is clicked', () => {
    const onClose = vi.fn();
    render(
      <MobileDrawer open={true} onClose={onClose}>
        <div>Content</div>
      </MobileDrawer>
    );

    // Find the backdrop (first child div with aria-hidden)
    const backdrop = document.querySelector('[aria-hidden="true"]');
    expect(backdrop).not.toBeNull();
    fireEvent.click(backdrop!);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when Escape is pressed', () => {
    const onClose = vi.fn();
    render(
      <MobileDrawer open={true} onClose={onClose}>
        <div>Content</div>
      </MobileDrawer>
    );

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('locks body scroll when open', () => {
    const { unmount } = render(
      <MobileDrawer open={true} onClose={vi.fn()}>
        <div>Content</div>
      </MobileDrawer>
    );

    expect(document.body.style.overflow).toBe('hidden');

    unmount();

    expect(document.body.style.overflow).toBe('');
  });
});
