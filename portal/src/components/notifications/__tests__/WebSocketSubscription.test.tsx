import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { screen, waitFor } from '@testing-library/react';
import { act } from 'react';

class MockWebSocket {
  static instances: MockWebSocket[] = [];
  url: string;
  listeners: Record<string, Array<(event: { data?: string }) => void>> = {};

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  addEventListener(type: string, callback: (event: { data?: string }) => void) {
    this.listeners[type] = this.listeners[type] ?? [];
    this.listeners[type].push(callback);
  }

  close() {
    this.emit('close', {});
  }

  emit(type: string, event: { data?: string }) {
    (this.listeners[type] ?? []).forEach((cb) => cb(event));
  }

  emitMessage(payload: unknown) {
    this.emit('message', { data: JSON.stringify(payload) });
  }
}

expectTranslations(['Notifications']);

describe.each(TEST_LOCALES)('useNotifications WebSocket (%s)', (locale) => {
  beforeEach(async () => {
    MockWebSocket.instances = [];
    await setLocale(locale);
    process.env.NEXT_PUBLIC_NOTIFICATIONS_WS = 'ws://localhost';
    global.fetch = vi.fn(() =>
      Promise.resolve({
        json: () =>
          Promise.resolve({
            notifications: [
              {
                id: 'old',
                userAddress: 'virtengine1demo',
                type: 'order_update',
                title: 'Old update',
                body: 'Old',
                createdAt: new Date(Date.now() - 1000).toISOString(),
                channels: ['in_app'],
              },
            ],
            unreadCount: 0,
          }),
      })
    ) as unknown as typeof fetch;

    global.WebSocket = MockWebSocket as unknown as typeof WebSocket;
    vi.resetModules();
  });

  it('prepends websocket notifications and increments unread count', async () => {
    const { useNotifications } = await import('@/components/notifications/hooks/useNotifications');

    function HookViewer() {
      const { notifications, unreadCount } = useNotifications();
      return (
        <div>
          <div data-testid="count">{unreadCount}</div>
          <div data-testid="first">{notifications[0]?.title ?? 'none'}</div>
        </div>
      );
    }

    renderWithI18n(<HookViewer />);

    await waitFor(() => {
      expect(screen.getByTestId('first').textContent).toBe('Old update');
    });

    const ws = MockWebSocket.instances[0];
    act(() => {
      ws.emitMessage({
        id: 'new',
        userAddress: 'virtengine1demo',
        type: 'security_alert',
        title: 'Realtime alert',
        body: 'New',
        createdAt: new Date().toISOString(),
        channels: ['in_app'],
      });
    });

    await waitFor(() => {
      expect(screen.getByTestId('count').textContent).toBe('1');
      expect(screen.getByTestId('first').textContent).toBe('Realtime alert');
    });
  });
});
