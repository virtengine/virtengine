/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Toaster } from '../Toaster';
import { Button } from '../Button';
import { toast } from '@/hooks/use-toast';

const meta = {
  title: 'UI/Toaster',
  component: Toaster,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<typeof Toaster>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="flex min-h-[200px] w-[360px] flex-col items-center justify-center gap-4">
      <Button
        onClick={() =>
          toast({
            title: 'Wallet connected',
            description: 'Keplr is now linked to your VirtEngine account.',
          })
        }
      >
        Show toast
      </Button>
      <Toaster />
    </div>
  ),
};
