/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Governance types for proposals, votes, and staking delegations.
 */

export type ProposalStatus =
  | 'PROPOSAL_STATUS_UNSPECIFIED'
  | 'PROPOSAL_STATUS_DEPOSIT_PERIOD'
  | 'PROPOSAL_STATUS_VOTING_PERIOD'
  | 'PROPOSAL_STATUS_PASSED'
  | 'PROPOSAL_STATUS_REJECTED'
  | 'PROPOSAL_STATUS_FAILED';

export type VoteOption =
  | 'VOTE_OPTION_UNSPECIFIED'
  | 'VOTE_OPTION_YES'
  | 'VOTE_OPTION_ABSTAIN'
  | 'VOTE_OPTION_NO'
  | 'VOTE_OPTION_NO_WITH_VETO';

export interface WeightedVoteOption {
  option: VoteOption;
  weight: string;
}

export interface TallyResult {
  yes_count?: string;
  no_count?: string;
  abstain_count?: string;
  no_with_veto_count?: string;
}

export interface ProposalMessage {
  '@type': string;
  [key: string]: unknown;
}

export interface ProposalMetadata {
  title?: string;
  summary?: string;
  description?: string;
  forum?: string;
  content?: string;
  [key: string]: unknown;
}

export interface GovernanceProposal {
  id: string;
  title?: string;
  summary?: string;
  metadata?: string;
  status: ProposalStatus;
  final_tally_result?: TallyResult;
  submit_time?: string;
  deposit_end_time?: string;
  voting_start_time?: string;
  voting_end_time?: string;
  total_deposit?: Array<{ denom: string; amount: string }>;
  proposer?: string;
  messages?: ProposalMessage[];
}

export interface GovernanceVote {
  voter: string;
  options: WeightedVoteOption[];
  metadata?: string;
}

export interface TallyParams {
  quorum?: string;
  threshold?: string;
  veto_threshold?: string;
}

export interface VotingParams {
  voting_period?: string;
}

export interface GovernancePagination {
  next_key?: string | null;
  total?: string | null;
}

export interface GovernanceProposalWithTally extends GovernanceProposal {
  tally?: TallyResult;
}

export interface GovernanceProposalsResponse {
  proposals: GovernanceProposalWithTally[];
  pagination?: GovernancePagination;
  tallyParams?: TallyParams;
  bondedTokens?: string;
}

export interface GovernanceProposalDetailResponse {
  proposal: GovernanceProposal | null;
  tally?: TallyResult | null;
  votes?: GovernanceVote[];
  tallyParams?: TallyParams;
  votingParams?: VotingParams;
  bondedTokens?: string;
  relatedProposals?: GovernanceProposal[];
  voterVote?: GovernanceVote | null;
}

export interface GovernanceDelegation {
  delegation: {
    delegator_address: string;
    validator_address: string;
    shares: string;
  };
  balance: {
    denom: string;
    amount: string;
  };
}

export interface GovernanceValidator {
  operator_address: string;
  status?: string;
  tokens?: string;
  delegator_shares?: string;
  min_self_delegation?: string;
  description?: {
    moniker?: string;
    identity?: string;
    website?: string;
    details?: string;
  };
  commission?: {
    commission_rates?: {
      rate?: string;
    };
  };
}

export interface GovernanceDelegationsResponse {
  delegations: GovernanceDelegation[];
  validators: GovernanceValidator[];
}

export interface ValidatorVoteHistoryEntry {
  proposalId: string;
  vote: GovernanceVote | null;
  title?: string;
  status?: ProposalStatus;
}

export interface ValidatorVoteHistoryResponse {
  validator: string;
  voterAddress: string;
  proposals: ValidatorVoteHistoryEntry[];
}
