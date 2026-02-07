/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { formatRelativeTime, formatTokenAmount } from '@/lib/utils';
import type {
  GovernanceProposal,
  ProposalMetadata,
  ProposalStatus,
  ProposalMessage,
  TallyResult,
  VoteOption,
  TallyParams,
} from '@/types/governance';

const STATUS_LABELS: Record<ProposalStatus, string> = {
  PROPOSAL_STATUS_UNSPECIFIED: 'Unknown',
  PROPOSAL_STATUS_DEPOSIT_PERIOD: 'Deposit',
  PROPOSAL_STATUS_VOTING_PERIOD: 'Voting',
  PROPOSAL_STATUS_PASSED: 'Passed',
  PROPOSAL_STATUS_REJECTED: 'Rejected',
  PROPOSAL_STATUS_FAILED: 'Failed',
};

const STATUS_STYLES: Record<ProposalStatus, { bg: string; text: string }> = {
  PROPOSAL_STATUS_UNSPECIFIED: { bg: 'bg-muted', text: 'text-muted-foreground' },
  PROPOSAL_STATUS_DEPOSIT_PERIOD: { bg: 'bg-warning/10', text: 'text-warning' },
  PROPOSAL_STATUS_VOTING_PERIOD: { bg: 'bg-primary/10', text: 'text-primary' },
  PROPOSAL_STATUS_PASSED: { bg: 'bg-success/10', text: 'text-success' },
  PROPOSAL_STATUS_REJECTED: { bg: 'bg-destructive/10', text: 'text-destructive' },
  PROPOSAL_STATUS_FAILED: { bg: 'bg-destructive/10', text: 'text-destructive' },
};

const VOTE_OPTION_LABELS: Record<VoteOption, string> = {
  VOTE_OPTION_UNSPECIFIED: 'Unspecified',
  VOTE_OPTION_YES: 'Yes',
  VOTE_OPTION_NO: 'No',
  VOTE_OPTION_ABSTAIN: 'Abstain',
  VOTE_OPTION_NO_WITH_VETO: 'No with Veto',
};

const PROPOSAL_TYPE_LABELS: Record<string, string> = {
  'cosmos.gov.v1beta1.TextProposal': 'Text',
  'cosmos.gov.v1beta1.ParameterChangeProposal': 'Parameter',
  'cosmos.gov.v1beta1.CommunityPoolSpendProposal': 'Spend',
  'cosmos.upgrade.v1beta1.SoftwareUpgradeProposal': 'Software Upgrade',
  'cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal': 'Software Upgrade',
  'cosmos.params.v1beta1.ParameterChangeProposal': 'Parameter',
  'cosmos.gov.v1.MsgUpdateParams': 'Parameter',
  'cosmos.distribution.v1beta1.MsgCommunityPoolSpend': 'Spend',
  'cosmos.upgrade.v1beta1.MsgSoftwareUpgrade': 'Software Upgrade',
  'cosmos.gov.v1.MsgExecLegacyContent': 'Legacy',
};

function safeBigInt(value?: string): bigint {
  try {
    if (!value) return 0n;
    return BigInt(value);
  } catch {
    return 0n;
  }
}

export function formatProposalStatus(status: ProposalStatus): string {
  return STATUS_LABELS[status] ?? 'Unknown';
}

export function getProposalStatusStyles(status: ProposalStatus): { bg: string; text: string } {
  return STATUS_STYLES[status] ?? STATUS_STYLES.PROPOSAL_STATUS_UNSPECIFIED;
}

export function formatVoteOption(option: VoteOption): string {
  return VOTE_OPTION_LABELS[option] ?? 'Unknown';
}

export function parseProposalMetadata(metadata?: string): ProposalMetadata | null {
  if (!metadata) return null;
  const trimmed = metadata.trim();
  if (!trimmed.startsWith('{') && !trimmed.startsWith('[')) return null;
  try {
    return JSON.parse(trimmed) as ProposalMetadata;
  } catch {
    return null;
  }
}

export function getProposalTitle(proposal: GovernanceProposal): string {
  const parsed = parseProposalMetadata(proposal.metadata);
  return (
    parsed?.title ||
    proposal.title ||
    parsed?.summary ||
    proposal.summary ||
    `Proposal #${proposal.id}`
  );
}

export function getProposalSummary(proposal: GovernanceProposal): string {
  const parsed = parseProposalMetadata(proposal.metadata);
  return parsed?.summary || proposal.summary || parsed?.description || 'No summary provided.';
}

export function getProposalBody(proposal: GovernanceProposal): string {
  const parsed = parseProposalMetadata(proposal.metadata);
  if (parsed?.content) return String(parsed.content);
  if (parsed?.description) return String(parsed.description);
  if (parsed?.summary) return String(parsed.summary);
  return proposal.summary || proposal.metadata || 'No proposal content available.';
}

export function getProposalForumUrl(proposal: GovernanceProposal): string | null {
  const parsed = parseProposalMetadata(proposal.metadata);
  if (parsed?.forum && typeof parsed.forum === 'string') return parsed.forum;
  return null;
}

export function getProposalType(messages?: ProposalMessage[]): string {
  if (!messages || messages.length === 0) return 'Unknown';
  const type = messages[0]?.['@type'];
  if (!type || typeof type !== 'string') return 'Unknown';
  return PROPOSAL_TYPE_LABELS[type] ?? type.split('.').pop() ?? 'Unknown';
}

export function getVotingTimeRemaining(endTime?: string): string {
  if (!endTime) return 'Unknown';
  return formatRelativeTime(endTime);
}

export function getTallyCounts(tally?: TallyResult) {
  return {
    yes: safeBigInt(tally?.yes_count),
    no: safeBigInt(tally?.no_count),
    abstain: safeBigInt(tally?.abstain_count),
    veto: safeBigInt(tally?.no_with_veto_count),
  };
}

export function getTallyTotal(tally?: TallyResult): bigint {
  const { yes, no, abstain, veto } = getTallyCounts(tally);
  return yes + no + abstain + veto;
}

export function calculateTallyPercentages(tally?: TallyResult) {
  const total = getTallyTotal(tally);
  if (total === 0n) {
    return { yes: 0, no: 0, abstain: 0, veto: 0 };
  }
  const counts = getTallyCounts(tally);
  return {
    yes: Number((counts.yes * 100n) / total),
    no: Number((counts.no * 100n) / total),
    abstain: Number((counts.abstain * 100n) / total),
    veto: Number((counts.veto * 100n) / total),
  };
}

export function calculateQuorumProgress(
  tally: TallyResult | undefined,
  bondedTokens?: string
): number {
  const totalVotes = getTallyTotal(tally);
  const bonded = safeBigInt(bondedTokens);
  if (bonded === 0n) return 0;
  const ratio = Number((totalVotes * 10000n) / bonded) / 10000;
  return Math.min(Math.max(ratio, 0), 1);
}

export function formatQuorumTarget(tallyParams?: TallyParams): string {
  if (!tallyParams?.quorum) return 'Quorum target unavailable';
  const quorum = parseFloat(tallyParams.quorum);
  if (Number.isNaN(quorum)) return 'Quorum target unavailable';
  return `Quorum target ${Math.round(quorum * 100)}%`;
}

export function formatTokenBalance(amount?: string, decimals = 6, displayDecimals = 2): string {
  if (!amount) return '0';
  return formatTokenAmount(amount, decimals, displayDecimals);
}

function escapeHtml(input: string): string {
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function renderInlineMarkdown(text: string): string {
  let rendered = escapeHtml(text);
  rendered = rendered.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  rendered = rendered.replace(/\*([^*]+)\*/g, '<em>$1</em>');
  rendered = rendered.replace(/`([^`]+)`/g, '<code>$1</code>');
  rendered = rendered.replace(
    /\[([^\]]+)\]\(([^)]+)\)/g,
    '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>'
  );
  return rendered;
}

export function renderMarkdownToHtml(markdown: string): string {
  const lines = markdown.split(/\r?\n/);
  const blocks: string[] = [];
  let inCodeBlock = false;
  let codeBuffer: string[] = [];
  let listBuffer: string[] = [];

  const flushList = () => {
    if (listBuffer.length === 0) return;
    blocks.push(`<ul>${listBuffer.join('')}</ul>`);
    listBuffer = [];
  };

  const flushCode = () => {
    if (!inCodeBlock || codeBuffer.length === 0) return;
    const code = escapeHtml(codeBuffer.join('\n'));
    blocks.push(`<pre><code>${code}</code></pre>`);
    codeBuffer = [];
  };

  for (const line of lines) {
    if (line.trim().startsWith('```')) {
      if (inCodeBlock) {
        flushCode();
        inCodeBlock = false;
      } else {
        flushList();
        inCodeBlock = true;
      }
      continue;
    }

    if (inCodeBlock) {
      codeBuffer.push(line);
      continue;
    }

    const headingMatch = line.match(/^(#{1,3})\s+(.*)$/);
    if (headingMatch) {
      flushList();
      const level = headingMatch[1].length;
      const content = renderInlineMarkdown(headingMatch[2]);
      blocks.push(`<h${level}>${content}</h${level}>`);
      continue;
    }

    const listMatch = line.match(/^(?:-|\*)\s+(.*)$/);
    if (listMatch) {
      listBuffer.push(`<li>${renderInlineMarkdown(listMatch[1])}</li>`);
      continue;
    }

    if (line.trim() === '') {
      flushList();
      continue;
    }

    flushList();
    blocks.push(`<p>${renderInlineMarkdown(line)}</p>`);
  }

  if (inCodeBlock) {
    flushCode();
  }
  flushList();

  return blocks.join('');
}
