/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Textarea } from '../Textarea';
import { Label } from '../Label';

const meta = {
  title: 'UI/Textarea',
  component: Textarea,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<typeof Textarea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="grid w-96 gap-2">
      <Label htmlFor="notes">Deployment notes</Label>
      <Textarea id="notes" placeholder="Describe the workload and any special requirements." />
    </div>
  ),
};
