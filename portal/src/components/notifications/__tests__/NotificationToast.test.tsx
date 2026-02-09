import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useState } from 'react';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithI18n, setLocale, TEST_LOCALES, expectTranslations } from '@/test-utils/i18n';
import { ToastAction } from '@/components/ui/Toast';
import i18n from '@/i18n';

expectTranslations(['Notifications']);

describe.each(TEST_LOCALES)('NotificationToast (%s)', (locale) => {
  beforeEach(async () => {
    await setLocale(locale);
  });

  it('renders toast with action and auto-dismisses', async () => {
    vi.resetModules();
    const { useToast } = await import('@/hooks/use-toast');

    function ToastHarness({ locale: activeLocale }: { locale: string }) {
      const { toast, dismiss, toasts } = useToast();
      const [lastId, setLastId] = useState<string | null>(null);
      return (
        <div>
          <button
            type="button"
            onClick={() => {
              const created = toast({
                title: i18n.t('Notifications', { lng: activeLocale }),
                description: 'New alert',
                action: (
                  <ToastAction altText="Undo" onClick={() => {}}>
                    Undo
                  </ToastAction>
                ),
              });
              setLastId(created.id);
            }}
          >
            Trigger
          </button>
          <button type="button" onClick={() => lastId && dismiss(lastId)}>
            Auto dismiss
          </button>
          <div data-testid="toast-count">{toasts.length}</div>
          {toasts.map((item) => (
            <div key={item.id}>
              <span>{item.title}</span>
              <span data-testid="toast-open">{String(item.open)}</span>
              {item.action}
            </div>
          ))}
        </div>
      );
    }

    renderWithI18n(<ToastHarness locale={locale} />);

    fireEvent.click(screen.getByRole('button', { name: 'Trigger' }));
    expect(screen.getByTestId('toast-count').textContent).toBe('1');
    expect(screen.getByText(i18n.t('Notifications', { lng: locale }))).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Undo' })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Auto dismiss' }));
    await waitFor(() => {
      expect(screen.getByTestId('toast-open').textContent).toBe('false');
    });
  });
});
