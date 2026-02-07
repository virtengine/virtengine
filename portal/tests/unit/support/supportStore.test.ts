import { describe, it, expect, beforeEach } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import {
  useSupportStore,
  getSlaTargetHours,
  type SupportCategory,
  type SupportPriority,
  type SupportStatus,
} from '@/stores/supportStore';

describe('supportStore', () => {
  beforeEach(() => {
    const { result } = renderHook(() => useSupportStore());
    // Reset store to seed state by re-creating
    act(() => {
      useSupportStore.setState({
        tickets: useSupportStore.getState().tickets,
        providers: useSupportStore.getState().providers,
        isLoading: false,
        error: null,
      });
    });
  });

  describe('initial state', () => {
    it('should have seed tickets', () => {
      const { result } = renderHook(() => useSupportStore());
      expect(result.current.tickets.length).toBeGreaterThanOrEqual(2);
    });

    it('should have providers', () => {
      const { result } = renderHook(() => useSupportStore());
      expect(result.current.providers.length).toBe(3);
    });

    it('should not be loading', () => {
      const { result } = renderHook(() => useSupportStore());
      expect(result.current.isLoading).toBe(false);
    });

    it('should have no error', () => {
      const { result } = renderHook(() => useSupportStore());
      expect(result.current.error).toBeNull();
    });
  });

  describe('createTicket', () => {
    it('should add a new ticket to the list', () => {
      const { result } = renderHook(() => useSupportStore());
      const initialCount = result.current.tickets.length;

      act(() => {
        result.current.createTicket({
          subject: 'Test ticket',
          description: 'Test description',
          category: 'technical',
          priority: 'normal',
          providerId: 'provider-orion',
        });
      });

      expect(result.current.tickets.length).toBe(initialCount + 1);
    });

    it('should set the new ticket to open status', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'New ticket',
          description: 'Some issue',
          category: 'billing',
          priority: 'high',
          providerId: 'provider-orion',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.status).toBe('open');
    });

    it('should populate chain metadata', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Chain test',
          description: 'Chain metadata check',
          category: 'technical',
          priority: 'urgent',
          providerId: 'provider-orion',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.chain.providerAddress).toBeTruthy();
      expect(newTicket.chain.customerAddress).toBeTruthy();
      expect(newTicket.chain.contentRef).toBeTruthy();
      expect(newTicket.chain.responseDeadline).toBeInstanceOf(Date);
    });

    it('should set correct category and priority', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Priority test',
          description: 'Urgent issue',
          category: 'security',
          priority: 'urgent',
          providerId: 'provider-summit',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.category).toBe('security');
      expect(newTicket.priority).toBe('urgent');
    });

    it('should include the initial description as the first response', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Response test',
          description: 'First message content',
          category: 'account',
          priority: 'low',
          providerId: 'provider-orion',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.responses).toHaveLength(1);
      expect(newTicket.responses[0].message).toBe('First message content');
      expect(newTicket.responses[0].isAgent).toBe(false);
    });

    it('should generate timeline events on creation', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Timeline test',
          description: 'Should have timeline',
          category: 'technical',
          priority: 'normal',
          providerId: 'provider-orion',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.timeline.length).toBeGreaterThanOrEqual(2);
    });

    it('should associate related entity when provided', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Related entity test',
          description: 'Has related deployment',
          category: 'technical',
          priority: 'normal',
          providerId: 'provider-orion',
          relatedEntity: { type: 'deployment', id: 'ord-123' },
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.relatedEntity).toEqual({ type: 'deployment', id: 'ord-123' });
      expect(newTicket.chain.allocationId).toBe('ord-123');
    });

    it('should configure waldur sync when provider uses waldur', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Waldur sync test',
          description: 'Provider with Waldur',
          category: 'technical',
          priority: 'normal',
          providerId: 'provider-orion',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.sync.waldur.status).toBe('queued');
    });

    it('should set waldur sync to not_configured for native desk providers', () => {
      const { result } = renderHook(() => useSupportStore());

      act(() => {
        result.current.createTicket({
          subject: 'Native desk test',
          description: 'Provider without Waldur',
          category: 'technical',
          priority: 'normal',
          providerId: 'provider-summit',
        });
      });

      const newTicket = result.current.tickets[0];
      expect(newTicket.sync.waldur.status).toBe('not_configured');
    });
  });

  describe('addResponse', () => {
    it('should add a customer response to a ticket', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;
      const initialResponseCount = result.current.tickets[0].responses.length;

      act(() => {
        result.current.addResponse(ticketId, {
          message: 'Customer reply',
          isAgent: false,
        });
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.responses.length).toBe(initialResponseCount + 1);
      expect(ticket.responses[ticket.responses.length - 1].message).toBe('Customer reply');
      expect(ticket.responses[ticket.responses.length - 1].isAgent).toBe(false);
    });

    it('should add an agent response to a ticket', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;

      act(() => {
        result.current.addResponse(ticketId, {
          message: 'Agent reply',
          isAgent: true,
          author: 'Support Agent Test',
        });
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      const lastResponse = ticket.responses[ticket.responses.length - 1];
      expect(lastResponse.isAgent).toBe(true);
      expect(lastResponse.author).toBe('Support Agent Test');
    });

    it('should update ticket status to waiting_support on customer reply', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;

      act(() => {
        result.current.addResponse(ticketId, {
          message: 'Need help',
          isAgent: false,
        });
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.status).toBe('waiting_support');
    });

    it('should update ticket status to waiting_customer on agent reply', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;

      act(() => {
        result.current.addResponse(ticketId, {
          message: 'Working on it',
          isAgent: true,
        });
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.status).toBe('waiting_customer');
    });

    it('should add timeline event on response', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;
      const initialTimelineCount = result.current.tickets[0].timeline.length;

      act(() => {
        result.current.addResponse(ticketId, {
          message: 'Timeline check',
          isAgent: false,
        });
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.timeline.length).toBe(initialTimelineCount + 1);
    });

    it('should update lastResponseAt timestamp', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;
      const before = new Date();

      act(() => {
        result.current.addResponse(ticketId, {
          message: 'Timestamp check',
          isAgent: false,
        });
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.lastResponseAt).toBeDefined();
      expect(ticket.lastResponseAt!.getTime()).toBeGreaterThanOrEqual(before.getTime());
    });
  });

  describe('updateStatus', () => {
    it('should update ticket status', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;

      act(() => {
        result.current.updateStatus(ticketId, 'resolved');
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.status).toBe('resolved');
    });

    it('should add timeline event on status change', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;
      const initialTimelineCount = result.current.tickets[0].timeline.length;

      act(() => {
        result.current.updateStatus(ticketId, 'closed');
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.timeline.length).toBe(initialTimelineCount + 1);
      const lastEvent = ticket.timeline[ticket.timeline.length - 1];
      expect(lastEvent.label).toBe('Status updated on-chain');
    });

    it('should update the updatedAt timestamp', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;
      const before = new Date();

      act(() => {
        result.current.updateStatus(ticketId, 'in_progress');
      });

      const ticket = result.current.tickets.find((t) => t.id === ticketId)!;
      expect(ticket.updatedAt.getTime()).toBeGreaterThanOrEqual(before.getTime());
    });

    it('should not modify other tickets', () => {
      const { result } = renderHook(() => useSupportStore());
      const ticketId = result.current.tickets[0].id;
      const otherTicket = result.current.tickets[1];

      act(() => {
        result.current.updateStatus(ticketId, 'resolved');
      });

      const unchanged = result.current.tickets.find((t) => t.id === otherTicket.id)!;
      expect(unchanged.status).toBe(otherTicket.status);
    });
  });

  describe('getSlaTargetHours', () => {
    it('should return 72 for low priority', () => {
      expect(getSlaTargetHours('low')).toBe(72);
    });

    it('should return 48 for normal priority', () => {
      expect(getSlaTargetHours('normal')).toBe(48);
    });

    it('should return 24 for high priority', () => {
      expect(getSlaTargetHours('high')).toBe(24);
    });

    it('should return 4 for urgent priority', () => {
      expect(getSlaTargetHours('urgent')).toBe(4);
    });
  });

  describe('seed data integrity', () => {
    it('should have tickets with valid provider references', () => {
      const { result } = renderHook(() => useSupportStore());
      for (const ticket of result.current.tickets) {
        expect(ticket.provider).toBeDefined();
        expect(ticket.provider.id).toBeTruthy();
        expect(ticket.provider.name).toBeTruthy();
      }
    });

    it('should have tickets with valid chain metadata', () => {
      const { result } = renderHook(() => useSupportStore());
      for (const ticket of result.current.tickets) {
        expect(ticket.chain.ticketId).toBeTruthy();
        expect(ticket.chain.providerAddress).toBeTruthy();
        expect(ticket.chain.customerAddress).toBeTruthy();
        expect(ticket.chain.contentRef).toBeTruthy();
      }
    });

    it('should have tickets with sync records for all channels', () => {
      const { result } = renderHook(() => useSupportStore());
      for (const ticket of result.current.tickets) {
        expect(ticket.sync.chain).toBeDefined();
        expect(ticket.sync.provider).toBeDefined();
        expect(ticket.sync.waldur).toBeDefined();
      }
    });
  });
});
