/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import { Label } from '@/components/ui/Label';
import { cn } from '@/lib/utils';

const LANGUAGE_OPTIONS = [
  { value: 'en', label: 'English' },
  { value: 'es', label: 'Español' },
  { value: 'de', label: 'Deutsch' },
  { value: 'ja', label: '日本語' },
];

interface LanguageSwitcherProps {
  className?: string;
}

export function LanguageSwitcher({ className }: LanguageSwitcherProps) {
  const { i18n, t } = useTranslation();

  const currentLanguage = useMemo(() => {
    const normalized = (i18n.language || 'en').split('-')[0];
    return LANGUAGE_OPTIONS.some((option) => option.value === normalized) ? normalized : 'en';
  }, [i18n.language]);

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Label htmlFor="language-select" className="sr-only">
        {t('Language')}
      </Label>
      <Select value={currentLanguage} onValueChange={(value) => void i18n.changeLanguage(value)}>
        <SelectTrigger
          id="language-select"
          className="h-9 w-[140px]"
          aria-label={t('Select language')}
        >
          <SelectValue placeholder={t('Language')} />
        </SelectTrigger>
        <SelectContent>
          {LANGUAGE_OPTIONS.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
