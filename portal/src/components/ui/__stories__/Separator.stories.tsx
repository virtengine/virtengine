/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Separator } from '../Separator';

const meta = {
  title: 'UI/Separator',
  component: Separator,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<typeof Separator>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Horizontal: Story = {
  render: () => (
    <div className="w-80 space-y-4">
      <div>
        <h4 className="text-sm font-medium">Overview</h4>
        <p className="text-sm text-muted-foreground">Wallet balances and recent activity.</p>
      </div>
      <Separator />
      <div>
        <h4 className="text-sm font-medium">Details</h4>
        <p className="text-sm text-muted-foreground">Invoices and usage history.</p>
      </div>
    </div>
  ),
};

export const Vertical: Story = {
  render: () => (
    <div className="flex h-20 items-center">
      <div className="px-4 text-sm">Customer</div>
      <Separator orientation="vertical" />
      <div className="px-4 text-sm">Provider</div>
    </div>
  ),
};
