/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider, SimpleTooltip } from '../Tooltip';
import { Button } from '../Button';

const meta = {
  title: 'UI/Tooltip',
  component: Tooltip,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="outline">Hover me</Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>This is a tooltip</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ),
};

export const Simple: Story = {
  render: () => (
    <SimpleTooltip content="This is a simple tooltip">
      <Button variant="outline">Hover me</Button>
    </SimpleTooltip>
  ),
};

export const Positions: Story = {
  render: () => (
    <div className="flex gap-4">
      <SimpleTooltip content="Top tooltip" side="top">
        <Button variant="outline">Top</Button>
      </SimpleTooltip>
      <SimpleTooltip content="Right tooltip" side="right">
        <Button variant="outline">Right</Button>
      </SimpleTooltip>
      <SimpleTooltip content="Bottom tooltip" side="bottom">
        <Button variant="outline">Bottom</Button>
      </SimpleTooltip>
      <SimpleTooltip content="Left tooltip" side="left">
        <Button variant="outline">Left</Button>
      </SimpleTooltip>
    </div>
  ),
};

export const WithDelay: Story = {
  render: () => (
    <SimpleTooltip content="This tooltip has a longer delay" delayDuration={500}>
      <Button variant="outline">Hover (500ms delay)</Button>
    </SimpleTooltip>
  ),
};

export const RichContent: Story = {
  render: () => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="outline">Rich Content</Button>
        </TooltipTrigger>
        <TooltipContent className="max-w-xs">
          <div className="space-y-2">
            <p className="font-semibold">Tooltip Title</p>
            <p className="text-sm text-muted-foreground">
              This tooltip contains more complex content with multiple lines of text.
            </p>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ),
};

export const OnIconButton: Story = {
  render: () => (
    <SimpleTooltip content="Delete item">
      <Button variant="outline" size="icon">
        🗑️
      </Button>
    </SimpleTooltip>
  ),
};
