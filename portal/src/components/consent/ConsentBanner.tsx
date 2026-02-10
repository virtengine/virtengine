/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useEffect, useMemo, useState } from 'react';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import { Label } from '@/components/ui/Label';
import { useTranslation } from 'react-i18next';

type ConsentPreferences = {
  necessary: true;
  analytics: boolean;
  marketing: boolean;
};

const STORAGE_KEY = 've-consent-preferences';

const DEFAULT_PREFERENCES: ConsentPreferences = {
  necessary: true,
  analytics: false,
  marketing: false,
};

function applyConsent(preferences: ConsentPreferences) {
  const gtag = (window as Window & { gtag?: (...args: unknown[]) => void }).gtag;
  if (typeof gtag === 'function') {
    gtag('consent', 'update', {
      analytics_storage: preferences.analytics ? 'granted' : 'denied',
      ad_storage: preferences.marketing ? 'granted' : 'denied',
      functionality_storage: 'granted',
      security_storage: 'granted',
    });
  }
}

function loadStoredPreferences(): ConsentPreferences | null {
  if (typeof window === 'undefined') return null;
  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<ConsentPreferences>;
    return {
      necessary: true,
      analytics: Boolean(parsed.analytics),
      marketing: Boolean(parsed.marketing),
    };
  } catch {
    return null;
  }
}

export function ConsentBanner() {
  const { t } = useTranslation();
  const [visible, setVisible] = useState(false);
  const [customize, setCustomize] = useState(false);
  const [preferences, setPreferences] = useState<ConsentPreferences>(DEFAULT_PREFERENCES);

  useEffect(() => {
    const stored = loadStoredPreferences();
    if (stored) {
      setPreferences(stored);
      setVisible(false);
    } else {
      setVisible(true);
    }
  }, []);

  const summary = useMemo(() => {
    if (preferences.analytics && preferences.marketing) return t('All optional cookies enabled');
    if (!preferences.analytics && !preferences.marketing) return t('Optional cookies disabled');
    return t('Custom cookie preferences');
  }, [preferences, t]);

  const persist = (next: ConsentPreferences) => {
    setPreferences(next);
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    applyConsent(next);
    setVisible(false);
    setCustomize(false);
  };

  if (!visible) return null;

  return (
    <div className="fixed bottom-4 left-4 right-4 z-50 sm:left-6 sm:right-6">
      <Card className="border-border bg-background/95 shadow-lg backdrop-blur">
        <CardContent className="space-y-4 p-4 sm:p-6">
          <div className="space-y-1">
            <p className="text-sm font-semibold">
              {t('We use cookies to improve the VirtEngine experience.')}
            </p>
            <p className="text-xs text-muted-foreground">{summary}</p>
          </div>

          {customize && (
            <div className="space-y-3 rounded-lg border border-border bg-muted/20 p-3">
              <div className="flex items-center justify-between">
                <Label className="text-xs text-muted-foreground">{t('Necessary cookies')}</Label>
                <span className="text-xs font-semibold text-emerald-600">{t('Always on')}</span>
              </div>
              <label className="flex items-center justify-between text-xs text-muted-foreground">
                <span>{t('Analytics cookies')}</span>
                <input
                  type="checkbox"
                  checked={preferences.analytics}
                  onChange={(event) =>
                    setPreferences({ ...preferences, analytics: event.target.checked })
                  }
                  className="h-4 w-4 rounded border-border"
                />
              </label>
              <label className="flex items-center justify-between text-xs text-muted-foreground">
                <span>{t('Marketing cookies')}</span>
                <input
                  type="checkbox"
                  checked={preferences.marketing}
                  onChange={(event) =>
                    setPreferences({ ...preferences, marketing: event.target.checked })
                  }
                  className="h-4 w-4 rounded border-border"
                />
              </label>
            </div>
          )}

          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              onClick={() => persist({ necessary: true, analytics: true, marketing: true })}
            >
              {t('Accept all')}
            </Button>
            <Button
              size="sm"
              variant="secondary"
              onClick={() => persist({ necessary: true, analytics: false, marketing: false })}
            >
              {t('Reject all')}
            </Button>
            <Button size="sm" variant="outline" onClick={() => setCustomize((prev) => !prev)}>
              {customize ? t('Hide preferences') : t('Customize')}
            </Button>
            {customize && (
              <Button size="sm" variant="ghost" onClick={() => persist(preferences)}>
                {t('Save preferences')}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
