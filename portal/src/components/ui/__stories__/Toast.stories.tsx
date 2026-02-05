/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import {
  Toast,
  ToastAction,
  ToastClose,
  ToastDescription,
  ToastProvider,
  ToastTitle,
  ToastViewport,
} from '../Toast';

const meta = {
  title: 'UI/Toast',
  component: Toast,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<typeof Toast>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <ToastProvider>
      <div className="min-h-[160px] w-[360px]">
        <Toast defaultOpen>
          <div className="grid gap-1">
            <ToastTitle>Deployment queued</ToastTitle>
            <ToastDescription>
              Your workload will start once capacity is available.
            </ToastDescription>
          </div>
          <ToastAction altText="View deployment">View</ToastAction>
          <ToastClose />
        </Toast>
        <ToastViewport />
      </div>
    </ToastProvider>
  ),
};
