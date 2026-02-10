/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Skeleton, SkeletonText, SkeletonCard, SkeletonAvatar, SkeletonTable } from '../Skeleton';

const meta = {
  title: 'UI/Skeleton',
  component: Skeleton,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'circular', 'rectangular'],
    },
    animation: {
      control: 'select',
      options: ['pulse', 'shimmer', 'none'],
    },
  },
} satisfies Meta<typeof Skeleton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    className: 'h-4 w-[250px]',
  },
};

export const Circular: Story = {
  args: {
    variant: 'circular',
    className: 'h-12 w-12',
  },
};

export const Rectangular: Story = {
  args: {
    variant: 'rectangular',
    className: 'h-[125px] w-[250px]',
  },
};

export const ShimmerAnimation: Story = {
  args: {
    animation: 'shimmer',
    className: 'h-4 w-[250px]',
  },
};

export const NoAnimation: Story = {
  args: {
    animation: 'none',
    className: 'h-4 w-[250px]',
  },
};

export const TextSkeleton: Story = {
  render: () => <SkeletonText lines={4} className="w-[300px]" />,
};

export const CardSkeleton: Story = {
  render: () => <SkeletonCard className="w-[350px]" />,
};

export const AvatarSkeleton: Story = {
  render: () => (
    <div className="flex gap-4">
      <SkeletonAvatar size="sm" />
      <SkeletonAvatar size="default" />
      <SkeletonAvatar size="lg" />
    </div>
  ),
};

export const TableSkeleton: Story = {
  render: () => <SkeletonTable rows={5} columns={4} />,
};

export const ProfileCardSkeleton: Story = {
  render: () => (
    <div className="flex items-center space-x-4">
      <Skeleton variant="circular" className="h-12 w-12" />
      <div className="space-y-2">
        <Skeleton className="h-4 w-[200px]" />
        <Skeleton className="h-4 w-[150px]" />
      </div>
    </div>
  ),
};

export const FormSkeleton: Story = {
  render: () => (
    <div className="w-[300px] space-y-4">
      <div className="space-y-2">
        <Skeleton className="h-4 w-[80px]" />
        <Skeleton className="h-10 w-full" />
      </div>
      <div className="space-y-2">
        <Skeleton className="h-4 w-[100px]" />
        <Skeleton className="h-10 w-full" />
      </div>
      <div className="space-y-2">
        <Skeleton className="h-4 w-[120px]" />
        <Skeleton className="h-20 w-full" />
      </div>
      <Skeleton className="h-10 w-[100px]" />
    </div>
  ),
};
