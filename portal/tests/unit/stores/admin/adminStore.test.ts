import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  useAdminStore,
  selectActiveProposals,
  selectActiveValidators,
  selectOpenTickets,
  selectUrgentTickets,
} from '@/stores/adminStore';
import type { AdminRole } from '@/types/admin';

function seedAdminData() {
  const now = new Date();
  useAdminStore.setState({
    currentUserRoles: ['operator', 'governance', 'support', 'validator'] as AdminRole[],
    users: [
      {
        address: 've1user1',
        roles: ['operator', 'governance'] as AdminRole[],
        displayName: 'Alice',
        assignedAt: now,
        lastActive: now,
      },
      {
        address: 've1user2',
        roles: ['support'] as AdminRole[],
        displayName: 'Bob',
        assignedAt: now,
        lastActive: now,
      },
    ],
    proposals: [
      {
        id: 'prop-1',
        title: 'Upgrade v2',
        description: 'Software upgrade proposal',
        proposer: 've1proposer1',
        status: 'voting',
        submitTime: now,
        votingEndTime: new Date(Date.now() + 86400000),
        yesVotes: 10,
        noVotes: 2,
        abstainVotes: 1,
        vetoVotes: 0,
        totalDeposit: '1000000',
      },
      {
        id: 'prop-2',
        title: 'Params Change',
        description: 'Parameter change proposal',
        proposer: 've1proposer2',
        status: 'passed',
        submitTime: now,
        votingEndTime: new Date(Date.now() - 86400000),
        yesVotes: 20,
        noVotes: 1,
        abstainVotes: 0,
        vetoVotes: 0,
        totalDeposit: '500000',
      },
    ],
    validators: [
      {
        operatorAddress: 've1val1',
        moniker: 'Validator A',
        status: 'active',
        tokens: '10000000',
        delegatorShares: '10000000',
        commission: 0.1,
        uptime: 99.9,
        missedBlocks: 1,
        slashingEvents: [],
      },
      {
        operatorAddress: 've1val2',
        moniker: 'Validator B',
        status: 'jailed',
        tokens: '5000000',
        delegatorShares: '5000000',
        commission: 0.05,
        uptime: 85.0,
        missedBlocks: 100,
        jailedUntil: new Date(Date.now() + 86400000),
        slashingEvents: [],
      },
    ],
    supportTickets: [
      {
        id: 'ticket-1',
        ticketNumber: 'TK-001',
        subject: 'Deployment issue',
        submitter: 've1user1',
        provider: 've1prov1',
        priority: 'urgent',
        status: 'open',
        category: 'deployment',
        createdAt: now,
        updatedAt: now,
      },
      {
        id: 'ticket-2',
        ticketNumber: 'TK-002',
        subject: 'Billing question',
        submitter: 've1user2',
        provider: 've1prov2',
        priority: 'normal',
        status: 'assigned',
        category: 'billing',
        createdAt: now,
        updatedAt: now,
        assignedAgent: 'Agent One',
      },
    ],
    systemHealth: {
      blockHeight: 12345678,
      blockTime: 6.2,
      activeValidators: 42,
      totalValidators: 50,
      bondedTokens: '100000000',
      inflationRate: 0.07,
      communityPool: '50000',
      txThroughput: 120,
      avgGasPrice: 0.025,
      networkUptime: 99.95,
    },
  });
}

describe('adminStore', () => {
  beforeEach(() => {
    // Reset the store to initial state before each test, then seed data
    useAdminStore.setState(useAdminStore.getInitialState());
    seedAdminData();
  });

  describe('role management', () => {
    it('hasRole returns true when user has the role', () => {
      const { result } = renderHook(() => useAdminStore((s) => s.hasRole));
      expect(result.current('operator')).toBe(true);
      expect(result.current('governance')).toBe(true);
      expect(result.current('support')).toBe(true);
      expect(result.current('validator')).toBe(true);
    });

    it('hasAnyRole returns true when user has at least one matching role', () => {
      const { result } = renderHook(() => useAdminStore((s) => s.hasAnyRole));
      expect(result.current(['operator'])).toBe(true);
      expect(result.current(['governance', 'support'])).toBe(true);
    });

    it('assignRole adds a new role to a user', () => {
      const { result } = renderHook(() => useAdminStore());

      // Find a user who doesn't have all roles
      const user = result.current.users.find((u) => !u.roles.includes('support'));
      if (!user) return;

      act(() => {
        result.current.assignRole(user.address, 'support');
      });

      const updated = result.current.users.find((u) => u.address === user.address);
      expect(updated?.roles).toContain('support');
    });

    it('assignRole does not duplicate existing role', () => {
      const { result } = renderHook(() => useAdminStore());
      const user = result.current.users[0];
      const existingRole = user.roles[0];
      const originalLength = user.roles.length;

      act(() => {
        result.current.assignRole(user.address, existingRole);
      });

      const updated = result.current.users.find((u) => u.address === user.address);
      expect(updated?.roles.length).toBe(originalLength);
    });

    it('revokeRole removes a role from a user', () => {
      const { result } = renderHook(() => useAdminStore());
      const user = result.current.users[0];
      const roleToRemove = user.roles[0];

      act(() => {
        result.current.revokeRole(user.address, roleToRemove);
      });

      const updated = result.current.users.find((u) => u.address === user.address);
      expect(updated?.roles).not.toContain(roleToRemove);
    });
  });

  describe('ticket management', () => {
    it('updateTicketStatus changes the ticket status', async () => {
      const { result } = renderHook(() => useAdminStore());
      const ticket = result.current.supportTickets[0];

      await act(async () => {
        await result.current.updateTicketStatus(ticket.id, 'resolved');
      });

      const updated = result.current.supportTickets.find((t) => t.id === ticket.id);
      expect(updated?.status).toBe('resolved');
    });

    it('assignTicket sets agent and status', async () => {
      const { result } = renderHook(() => useAdminStore());
      const openTicket = result.current.supportTickets.find((t) => t.status === 'open');
      if (!openTicket) return;

      await act(async () => {
        await result.current.assignTicket(openTicket.id, 'Agent Smith');
      });

      const updated = result.current.supportTickets.find((t) => t.id === openTicket.id);
      expect(updated?.assignedAgent).toBe('Agent Smith');
      expect(updated?.status).toBe('assigned');
    });
  });

  describe('selectors', () => {
    it('selectActiveProposals returns only voting proposals', () => {
      const state = useAdminStore.getState();
      const active = selectActiveProposals(state);
      expect(active.length).toBeGreaterThan(0);
      active.forEach((p) => expect(p.status).toBe('voting'));
    });

    it('selectActiveValidators returns only active validators', () => {
      const state = useAdminStore.getState();
      const active = selectActiveValidators(state);
      expect(active.length).toBeGreaterThan(0);
      active.forEach((v) => expect(v.status).toBe('active'));
    });

    it('selectOpenTickets excludes resolved and closed', () => {
      const state = useAdminStore.getState();
      const open = selectOpenTickets(state);
      open.forEach((t) => {
        expect(t.status).not.toBe('resolved');
        expect(t.status).not.toBe('closed');
      });
    });

    it('selectUrgentTickets returns only urgent non-resolved tickets', () => {
      const state = useAdminStore.getState();
      const urgent = selectUrgentTickets(state);
      urgent.forEach((t) => {
        expect(t.priority).toBe('urgent');
        expect(t.status).not.toBe('resolved');
        expect(t.status).not.toBe('closed');
      });
    });
  });

  describe('seed data', () => {
    it('initialises with users', () => {
      const state = useAdminStore.getState();
      expect(state.users.length).toBeGreaterThan(0);
    });

    it('initialises with proposals', () => {
      const state = useAdminStore.getState();
      expect(state.proposals.length).toBeGreaterThan(0);
    });

    it('initialises with validators', () => {
      const state = useAdminStore.getState();
      expect(state.validators.length).toBeGreaterThan(0);
    });

    it('initialises with support tickets', () => {
      const state = useAdminStore.getState();
      expect(state.supportTickets.length).toBeGreaterThan(0);
    });

    it('initialises with system health metrics', () => {
      const state = useAdminStore.getState();
      expect(state.systemHealth.blockHeight).toBeGreaterThan(0);
      expect(state.systemHealth.networkUptime).toBeGreaterThan(0);
    });
  });
});
