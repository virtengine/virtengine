/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { renderMarkdownToHtml } from '@/lib/governance';

interface ProposalMarkdownProps {
  content: string;
}

export function ProposalMarkdown({ content }: ProposalMarkdownProps) {
  const html = renderMarkdownToHtml(content);
  return (
    <div
      className="prose prose-sm dark:prose-invert max-w-none"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
