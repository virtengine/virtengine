/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Progress, CircularProgress } from '../Progress';

const meta = {
  title: 'UI/Progress',
  component: Progress,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    value: {
      control: { type: 'range', min: 0, max: 100 },
    },
    size: {
      control: 'select',
      options: ['sm', 'default', 'lg'],
    },
    variant: {
      control: 'select',
      options: ['default', 'success', 'warning', 'destructive'],
    },
    showValue: {
      control: 'boolean',
    },
  },
} satisfies Meta<typeof Progress>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    value: 60,
    className: 'w-[300px]',
  },
};

export const Small: Story = {
  args: {
    value: 45,
    size: 'sm',
    className: 'w-[300px]',
  },
};

export const Large: Story = {
  args: {
    value: 75,
    size: 'lg',
    className: 'w-[300px]',
  },
};

export const WithValue: Story = {
  args: {
    value: 66,
    showValue: true,
    className: 'w-[250px]',
  },
};

export const Success: Story = {
  args: {
    value: 100,
    variant: 'success',
    className: 'w-[300px]',
  },
};

export const Warning: Story = {
  args: {
    value: 50,
    variant: 'warning',
    className: 'w-[300px]',
  },
};

export const Destructive: Story = {
  args: {
    value: 20,
    variant: 'destructive',
    className: 'w-[300px]',
  },
};

export const AllVariants: Story = {
  render: () => (
    <div className="space-y-4 w-[300px]">
      <div>
        <span className="text-sm text-muted-foreground">Default</span>
        <Progress value={60} />
      </div>
      <div>
        <span className="text-sm text-muted-foreground">Success</span>
        <Progress value={100} variant="success" />
      </div>
      <div>
        <span className="text-sm text-muted-foreground">Warning</span>
        <Progress value={50} variant="warning" />
      </div>
      <div>
        <span className="text-sm text-muted-foreground">Destructive</span>
        <Progress value={20} variant="destructive" />
      </div>
    </div>
  ),
};

export const Circular: Story = {
  render: () => (
    <div className="flex gap-4 items-center">
      <CircularProgress value={25} />
      <CircularProgress value={50} size={60} />
      <CircularProgress value={75} size={80} showValue />
      <CircularProgress value={100} variant="success" />
    </div>
  ),
};

export const CircularVariants: Story = {
  render: () => (
    <div className="flex gap-4 items-center">
      <CircularProgress value={60} showValue />
      <CircularProgress value={60} variant="success" showValue />
      <CircularProgress value={60} variant="warning" showValue />
      <CircularProgress value={60} variant="destructive" showValue />
    </div>
  ),
};
