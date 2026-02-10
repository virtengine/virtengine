/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Mail } from 'lucide-react';
import { Button } from '../Button';

const meta = {
  title: 'UI/Button',
  component: Button,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: [
        'default',
        'destructive',
        'outline',
        'secondary',
        'ghost',
        'link',
        'success',
        'warning',
      ],
    },
    size: {
      control: 'select',
      options: ['default', 'sm', 'lg', 'xl', 'icon', 'icon-sm', 'icon-lg'],
    },
    loading: {
      control: 'boolean',
    },
    disabled: {
      control: 'boolean',
    },
  },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: 'Button',
  },
};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
    children: 'Secondary',
  },
};

export const Destructive: Story = {
  args: {
    variant: 'destructive',
    children: 'Destructive',
  },
};

export const Outline: Story = {
  args: {
    variant: 'outline',
    children: 'Outline',
  },
};

export const Ghost: Story = {
  args: {
    variant: 'ghost',
    children: 'Ghost',
  },
};

export const Link: Story = {
  args: {
    variant: 'link',
    children: 'Link',
  },
};

export const Success: Story = {
  args: {
    variant: 'success',
    children: 'Success',
  },
};

export const Warning: Story = {
  args: {
    variant: 'warning',
    children: 'Warning',
  },
};

export const WithIcon: Story = {
  args: {
    children: (
      <>
        <Mail className="h-4 w-4" />
        Login with Email
      </>
    ),
  },
};

export const Loading: Story = {
  args: {
    loading: true,
    children: 'Please wait',
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
    children: 'Disabled',
  },
};

export const SmallSize: Story = {
  args: {
    size: 'sm',
    children: 'Small',
  },
};

export const LargeSize: Story = {
  args: {
    size: 'lg',
    children: 'Large',
  },
};

export const ExtraLargeSize: Story = {
  args: {
    size: 'xl',
    children: 'Extra Large',
  },
};

export const IconButton: Story = {
  args: {
    variant: 'outline',
    size: 'icon',
    children: <Mail className="h-4 w-4" />,
  },
};

export const AllVariants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button variant="default">Default</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="destructive">Destructive</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="link">Link</Button>
      <Button variant="success">Success</Button>
      <Button variant="warning">Warning</Button>
    </div>
  ),
};

export const AllSizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Button size="sm">Small</Button>
      <Button size="default">Default</Button>
      <Button size="lg">Large</Button>
      <Button size="xl">Extra Large</Button>
    </div>
  ),
};
