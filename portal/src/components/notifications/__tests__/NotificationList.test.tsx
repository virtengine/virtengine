import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { ReactNode } from 'react';
import { fireEvent, screen } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { NotificationCenter } from '@/components/notifications/NotificationCenter';
import i18n from '@/i18n';

vi.mock('@/components/ui/Dropdown', () => ({
  DropdownMenu: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  DropdownMenuTrigger: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  DropdownMenuContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

const markAllAsRead = vi.fn();
const markAsRead = vi.fn();

vi.mock('@/components/notifications/hooks/useNotifications', () => ({
  useNotifications: () => ({
    notifications: [
      {
        id: 'notif-1',
        userAddress: 'virtengine1demo',
        type: 'security_alert',
        title: 'Security update',
        body: 'New login detected',
        createdAt: new Date().toISOString(),
        channels: ['in_app'],
      },
    ],
    unreadCount: 1,
    isLoading: false,
    markAsRead,
    markAllAsRead,
  }),
}));

expectTranslations(['Notifications', 'Mark all read', 'Open notifications']);

describe.each(TEST_LOCALES)('NotificationCenter (%s)', (locale) => {
  beforeEach(async () => {
    markAllAsRead.mockClear();
    await setLocale(locale);
  });

  it('renders notifications and allows marking all as read', () => {
    renderWithI18n(<NotificationCenter />);

    fireEvent.click(
      screen.getByRole('button', { name: i18n.t('Open notifications', { lng: locale }) })
    );

    expect(screen.getByText(i18n.t('Notifications', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByText('Security update')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: i18n.t('Mark all read', { lng: locale }) }));
    expect(markAllAsRead).toHaveBeenCalled();
  });
});
