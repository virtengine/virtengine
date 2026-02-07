import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  useAdminStore,
  selectActiveProposals,
  selectActiveValidators,
  selectOpenTickets,
  selectUrgentTickets,
} from '@/stores/adminStore';

describe('adminStore', () => {
  beforeEach(() => {
    // Reset the store to initial state before each test
    useAdminStore.setState(useAdminStore.getInitialState());
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
    it('updateTicketStatus changes the ticket status', () => {
      const { result } = renderHook(() => useAdminStore());
      const ticket = result.current.supportTickets[0];

      act(() => {
        result.current.updateTicketStatus(ticket.id, 'resolved');
      });

      const updated = result.current.supportTickets.find((t) => t.id === ticket.id);
      expect(updated?.status).toBe('resolved');
    });

    it('assignTicket sets agent and status', () => {
      const { result } = renderHook(() => useAdminStore());
      const openTicket = result.current.supportTickets.find((t) => t.status === 'open');
      if (!openTicket) return;

      act(() => {
        result.current.assignTicket(openTicket.id, 'Agent Smith');
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
