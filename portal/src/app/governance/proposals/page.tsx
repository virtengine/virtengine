import type { Metadata } from 'next';
import { GovernanceProposalsClient } from './GovernanceProposalsClient';

export const metadata: Metadata = {
  title: 'Governance Proposals',
  description: 'View and vote on governance proposals',
};

export default function GovernanceProposalsPage() {
  return <GovernanceProposalsClient />;
}
