/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Label } from '../Label';
import { Input } from '../Input';

const meta = {
  title: 'UI/Label',
  component: Label,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<typeof Label>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="grid w-80 gap-4">
      <div className="grid gap-2">
        <Label htmlFor="email">Email</Label>
        <Input id="email" placeholder="you@virtengine.io" />
      </div>
      <div className="grid gap-2">
        <Label htmlFor="name" required>
          Full name
        </Label>
        <Input id="name" placeholder="Ada Lovelace" />
      </div>
    </div>
  ),
};
