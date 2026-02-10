/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import type { Meta, StoryObj } from '@storybook/react';
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '../Accordion';

const meta: Meta = {
  title: 'UI/Accordion',
  component: Accordion,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj;

export const Default: Story = {
  render: () => (
    <Accordion type="single" collapsible className="w-[400px]">
      <AccordionItem value="item-1">
        <AccordionTrigger>Is it accessible?</AccordionTrigger>
        <AccordionContent>Yes. It adheres to the WAI-ARIA design pattern.</AccordionContent>
      </AccordionItem>
      <AccordionItem value="item-2">
        <AccordionTrigger>Is it styled?</AccordionTrigger>
        <AccordionContent>
          Yes. It comes with default styles that matches the other components&apos; aesthetic.
        </AccordionContent>
      </AccordionItem>
      <AccordionItem value="item-3">
        <AccordionTrigger>Is it animated?</AccordionTrigger>
        <AccordionContent>
          Yes. It&apos;s animated by default, but you can disable it if you prefer.
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  ),
};

export const Multiple: Story = {
  render: () => (
    <Accordion type="multiple" className="w-[400px]">
      <AccordionItem value="item-1">
        <AccordionTrigger>First Section</AccordionTrigger>
        <AccordionContent>
          This accordion allows multiple items to be expanded at once.
        </AccordionContent>
      </AccordionItem>
      <AccordionItem value="item-2">
        <AccordionTrigger>Second Section</AccordionTrigger>
        <AccordionContent>You can expand this while keeping the first one open.</AccordionContent>
      </AccordionItem>
      <AccordionItem value="item-3">
        <AccordionTrigger>Third Section</AccordionTrigger>
        <AccordionContent>All three sections can be open simultaneously.</AccordionContent>
      </AccordionItem>
    </Accordion>
  ),
};

export const DefaultOpen: Story = {
  render: () => (
    <Accordion type="single" defaultValue="item-2" className="w-[400px]">
      <AccordionItem value="item-1">
        <AccordionTrigger>Section 1</AccordionTrigger>
        <AccordionContent>Content for section 1.</AccordionContent>
      </AccordionItem>
      <AccordionItem value="item-2">
        <AccordionTrigger>Section 2 (Default Open)</AccordionTrigger>
        <AccordionContent>This section is open by default.</AccordionContent>
      </AccordionItem>
      <AccordionItem value="item-3">
        <AccordionTrigger>Section 3</AccordionTrigger>
        <AccordionContent>Content for section 3.</AccordionContent>
      </AccordionItem>
    </Accordion>
  ),
};

export const FAQ: Story = {
  render: () => (
    <div className="w-[500px]">
      <h2 className="mb-4 text-xl font-semibold">Frequently Asked Questions</h2>
      <Accordion type="single" collapsible>
        <AccordionItem value="item-1">
          <AccordionTrigger>What payment methods do you accept?</AccordionTrigger>
          <AccordionContent>
            We accept all major credit cards (Visa, MasterCard, American Express), PayPal, and
            cryptocurrency payments.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger>How long does deployment take?</AccordionTrigger>
          <AccordionContent>
            Most deployments complete within 2-5 minutes. Complex configurations may take up to 15
            minutes.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger>Is there a free tier?</AccordionTrigger>
          <AccordionContent>
            Yes! We offer a free tier with limited resources for testing and development purposes.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-4">
          <AccordionTrigger>How do I get support?</AccordionTrigger>
          <AccordionContent>
            You can reach our support team through the in-app chat, email at support@virtengine.com,
            or through our Discord community.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};
