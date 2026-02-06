# Task 31D: Governance Voting UI

**vibe-kanban ID:** `14c5b93e-5fd8-4b4f-a2f7-ee60f64fef8b`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31D |
| **Title** | feat(portal): Governance voting UI |
| **Priority** | P1 |
| **Wave** | 3 |
| **Estimated LOC** | 3500 |
| **Duration** | 3 weeks |
| **Dependencies** | x/gov module exists |
| **Blocking** | None |

---

## Problem Statement

The x/gov module exists for governance proposals, but there is no user interface for:
- Viewing active proposals
- Casting votes or weighted votes
- Delegating voting power
- Viewing proposal history
- Creating new proposals (for validators)

Users must use CLI commands to participate in governance, limiting community participation.

### Current State Analysis

```
x/gov/                          ✅ Governance module wired
portal/src/pages/governance/    ❌ Does not exist
lib/portal/components/voting/   ❌ Does not exist
API for proposal queries:       ✅ CometBFT LCD/gRPC available
```

---

## Acceptance Criteria

### AC-1: Proposal Listing Page
- [ ] Active proposals list with status badges
- [ ] Voting countdown timers
- [ ] Proposal type categorization (parameter, software upgrade, text, spend)
- [ ] Quorum progress indicator
- [ ] Search and filter functionality
- [ ] Pagination with infinite scroll

### AC-2: Proposal Detail View
- [ ] Full proposal content with markdown rendering
- [ ] Vote tally visualization (Yes/No/NoWithVeto/Abstain)
- [ ] Voting period timeline
- [ ] Deposit information
- [ ] Proposer information
- [ ] Previous proposals by same proposer

### AC-3: Vote Casting Interface
- [ ] Simple vote options (Yes/No/NoWithVeto/Abstain)
- [ ] Weighted voting interface (split vote percentages)
- [ ] Transaction signing integration
- [ ] Voting power preview
- [ ] Confirmation modal with fee estimate

### AC-4: Delegation Management
- [ ] View delegated voting power
- [ ] Delegate to validator (voting power follows stake delegation)
- [ ] View validator voting history
- [ ] Alerts for upcoming proposals

---

## Technical Requirements

### Proposal List Component

```tsx
// portal/src/app/governance/page.tsx

import { Suspense } from 'react';
import { ProposalList } from '@/components/governance/ProposalList';
import { ProposalFilters } from '@/components/governance/ProposalFilters';

export default async function GovernancePage() {
  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-6">Governance</h1>
      
      <ProposalFilters />
      
      <Suspense fallback={<ProposalListSkeleton />}>
        <ProposalList />
      </Suspense>
    </div>
  );
}

// portal/src/components/governance/ProposalList.tsx

'use client';

import { useQuery } from '@tanstack/react-query';
import { useVirtEngine } from '@virtengine/portal';
import { ProposalCard } from './ProposalCard';
import { Proposal } from '@/types/governance';

export function ProposalList() {
  const { client } = useVirtEngine();
  
  const { data: proposals, isLoading } = useQuery({
    queryKey: ['proposals'],
    queryFn: async () => {
      const res = await client.gov.proposals('PROPOSAL_STATUS_VOTING_PERIOD');
      return res.proposals as Proposal[];
    },
  });

  if (isLoading) return <ProposalListSkeleton />;

  return (
    <div className="space-y-4">
      {proposals?.map(proposal => (
        <ProposalCard key={proposal.id} proposal={proposal} />
      ))}
    </div>
  );
}
```

### Proposal Card Component

```tsx
// portal/src/components/governance/ProposalCard.tsx

'use client';

import Link from 'next/link';
import { formatDistanceToNow } from 'date-fns';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Proposal, ProposalStatus } from '@/types/governance';
import { calculateQuorum } from '@/lib/governance';

interface ProposalCardProps {
  proposal: Proposal;
}

export function ProposalCard({ proposal }: ProposalCardProps) {
  const quorumProgress = calculateQuorum(proposal);
  const endTime = new Date(proposal.voting_end_time);
  const timeRemaining = formatDistanceToNow(endTime, { addSuffix: true });

  return (
    <Link href={`/governance/${proposal.id}`}>
      <div className="border rounded-lg p-6 hover:shadow-md transition-shadow">
        <div className="flex justify-between items-start mb-4">
          <div>
            <Badge variant={getStatusVariant(proposal.status)}>
              {formatStatus(proposal.status)}
            </Badge>
            <span className="ml-2 text-sm text-gray-500">#{proposal.id}</span>
          </div>
          <span className="text-sm text-gray-500">
            Ends {timeRemaining}
          </span>
        </div>

        <h3 className="text-xl font-semibold mb-2">
          {proposal.title || proposal.content?.title || 'Untitled Proposal'}
        </h3>
        
        <p className="text-gray-600 line-clamp-2 mb-4">
          {proposal.summary || proposal.content?.description || ''}
        </p>

        <div className="flex items-center gap-4">
          <VoteTally proposal={proposal} />
          <div className="flex-1">
            <div className="text-sm text-gray-500 mb-1">
              Quorum: {(quorumProgress * 100).toFixed(1)}%
            </div>
            <Progress value={quorumProgress * 100} />
          </div>
        </div>
      </div>
    </Link>
  );
}

function VoteTally({ proposal }: { proposal: Proposal }) {
  const { yes, no, no_with_veto, abstain } = proposal.final_tally_result;
  const total = BigInt(yes) + BigInt(no) + BigInt(no_with_veto) + BigInt(abstain);
  
  if (total === 0n) return <span className="text-sm text-gray-500">No votes yet</span>;
  
  const yesPercent = Number((BigInt(yes) * 100n) / total);
  const noPercent = Number((BigInt(no) * 100n) / total);
  
  return (
    <div className="flex gap-2 text-sm">
      <span className="text-green-600">{yesPercent}% Yes</span>
      <span className="text-red-600">{noPercent}% No</span>
    </div>
  );
}
```

### Vote Casting Component

```tsx
// portal/src/components/governance/VoteModal.tsx

'use client';

import { useState } from 'react';
import { useVirtEngine } from '@virtengine/portal';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { VoteOption, WeightedVoteOption } from '@/types/governance';

interface VoteModalProps {
  proposalId: string;
  open: boolean;
  onClose: () => void;
}

export function VoteModal({ proposalId, open, onClose }: VoteModalProps) {
  const [mode, setMode] = useState<'simple' | 'weighted'>('simple');
  const [simpleVote, setSimpleVote] = useState<VoteOption | null>(null);
  const [weights, setWeights] = useState({
    yes: 0,
    no: 0,
    noWithVeto: 0,
    abstain: 0,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { signAndBroadcast, address } = useVirtEngine();

  const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
      if (mode === 'simple' && simpleVote) {
        await signAndBroadcast({
          typeUrl: '/cosmos.gov.v1.MsgVote',
          value: {
            proposalId,
            voter: address,
            option: simpleVote,
          },
        });
      } else {
        // Weighted vote
        const options: WeightedVoteOption[] = [
          { option: 'VOTE_OPTION_YES', weight: (weights.yes / 100).toString() },
          { option: 'VOTE_OPTION_NO', weight: (weights.no / 100).toString() },
          { option: 'VOTE_OPTION_NO_WITH_VETO', weight: (weights.noWithVeto / 100).toString() },
          { option: 'VOTE_OPTION_ABSTAIN', weight: (weights.abstain / 100).toString() },
        ].filter(opt => parseFloat(opt.weight) > 0);

        await signAndBroadcast({
          typeUrl: '/cosmos.gov.v1.MsgVoteWeighted',
          value: {
            proposalId,
            voter: address,
            options,
          },
        });
      }
      onClose();
    } catch (error) {
      console.error('Vote failed:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const totalWeight = weights.yes + weights.no + weights.noWithVeto + weights.abstain;

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cast Your Vote</DialogTitle>
        </DialogHeader>

        <div className="flex gap-2 mb-6">
          <Button
            variant={mode === 'simple' ? 'default' : 'outline'}
            onClick={() => setMode('simple')}
          >
            Simple Vote
          </Button>
          <Button
            variant={mode === 'weighted' ? 'default' : 'outline'}
            onClick={() => setMode('weighted')}
          >
            Weighted Vote
          </Button>
        </div>

        {mode === 'simple' ? (
          <div className="grid grid-cols-2 gap-4">
            {(['VOTE_OPTION_YES', 'VOTE_OPTION_NO', 'VOTE_OPTION_NO_WITH_VETO', 'VOTE_OPTION_ABSTAIN'] as VoteOption[]).map(option => (
              <Button
                key={option}
                variant={simpleVote === option ? 'default' : 'outline'}
                onClick={() => setSimpleVote(option)}
                className={getVoteButtonClass(option, simpleVote === option)}
              >
                {formatVoteOption(option)}
              </Button>
            ))}
          </div>
        ) : (
          <div className="space-y-4">
            <WeightSlider label="Yes" value={weights.yes} onChange={v => setWeights(w => ({ ...w, yes: v }))} color="green" />
            <WeightSlider label="No" value={weights.no} onChange={v => setWeights(w => ({ ...w, no: v }))} color="red" />
            <WeightSlider label="No with Veto" value={weights.noWithVeto} onChange={v => setWeights(w => ({ ...w, noWithVeto: v }))} color="orange" />
            <WeightSlider label="Abstain" value={weights.abstain} onChange={v => setWeights(w => ({ ...w, abstain: v }))} color="gray" />
            
            {totalWeight !== 100 && totalWeight > 0 && (
              <p className="text-red-500 text-sm">Weights must sum to 100% (current: {totalWeight}%)</p>
            )}
          </div>
        )}

        <Button
          className="w-full mt-6"
          onClick={handleSubmit}
          disabled={isSubmitting || (mode === 'simple' && !simpleVote) || (mode === 'weighted' && totalWeight !== 100)}
        >
          {isSubmitting ? 'Signing...' : 'Submit Vote'}
        </Button>
      </DialogContent>
    </Dialog>
  );
}
```

### API Routes

```typescript
// portal/src/app/api/governance/proposals/route.ts

import { NextRequest, NextResponse } from 'next/server';
import { createVirtEngineClient } from '@/lib/virtengine';

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const status = searchParams.get('status') || 'PROPOSAL_STATUS_VOTING_PERIOD';
  const page = parseInt(searchParams.get('page') || '1');
  const limit = parseInt(searchParams.get('limit') || '20');

  const client = await createVirtEngineClient();
  
  const proposals = await client.gov.v1.proposals({
    proposalStatus: status,
    pagination: {
      offset: (page - 1) * limit,
      limit,
    },
  });

  return NextResponse.json(proposals);
}

// portal/src/app/api/governance/proposals/[id]/route.ts

export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  const client = await createVirtEngineClient();
  
  const [proposal, tally, votes] = await Promise.all([
    client.gov.v1.proposal({ proposalId: params.id }),
    client.gov.v1.tallyResult({ proposalId: params.id }),
    client.gov.v1.votes({ proposalId: params.id }),
  ]);

  return NextResponse.json({
    proposal: proposal.proposal,
    tally: tally.tally,
    votes: votes.votes,
  });
}
```

---

## Directory Structure

```
portal/src/
├── app/
│   ├── governance/
│   │   ├── page.tsx              # Proposal list page
│   │   └── [id]/
│   │       └── page.tsx          # Proposal detail page
│   └── api/governance/
│       ├── proposals/
│       │   ├── route.ts          # GET proposals
│       │   └── [id]/
│       │       └── route.ts      # GET single proposal
│       └── vote/
│           └── route.ts          # POST vote (if needed)
├── components/governance/
│   ├── ProposalList.tsx
│   ├── ProposalCard.tsx
│   ├── ProposalDetail.tsx
│   ├── VoteTally.tsx
│   ├── VoteModal.tsx
│   ├── WeightedVoteSlider.tsx
│   ├── ProposalFilters.tsx
│   └── QuorumProgress.tsx
├── hooks/
│   └── useGovernance.ts
└── types/
    └── governance.ts
```

---

## Testing Requirements

### Unit Tests
- Vote weight calculation
- Quorum progress calculation
- Proposal status formatting

### Integration Tests
- Proposal fetching from chain
- Vote transaction signing
- Pagination handling

### E2E Tests
- Full voting flow
- Weighted vote submission
- Proposal filtering

---

## Security Considerations

1. **Transaction Signing**: Confirm vote details before signing
2. **Vote Confirmation**: Show what user is voting for
3. **Delegation Awareness**: Show if vote goes through delegation
4. **Replay Protection**: Use proper nonces
