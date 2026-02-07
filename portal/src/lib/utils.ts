import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import i18n, { DEFAULT_LANGUAGE } from '@/i18n';

/**
 * Merge Tailwind CSS classes with clsx
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Format a token amount with decimals
 */
export function formatTokenAmount(
  amount: string | number,
  decimals = 6,
  displayDecimals = 2
): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount;
  const adjusted = num / Math.pow(10, decimals);
  const locale = i18n.language?.split('-')[0] || DEFAULT_LANGUAGE;
  return new Intl.NumberFormat(locale, {
    minimumFractionDigits: displayDecimals,
    maximumFractionDigits: displayDecimals,
  }).format(adjusted);
}

/**
 * Truncate an address for display
 */
export function truncateAddress(address: string, startChars = 10, endChars = 4): string {
  if (address.length <= startChars + endChars) {
    return address;
  }
  return `${address.slice(0, startChars)}...${address.slice(-endChars)}`;
}

/**
 * Format a date for display
 */
export function formatDate(date: Date | string, options?: Intl.DateTimeFormatOptions): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const locale = i18n.language?.split('-')[0] || DEFAULT_LANGUAGE;
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    ...options,
  }).format(d);
}

/**
 * Format a date/time for display
 */
export function formatDateTime(date: Date | string, options?: Intl.DateTimeFormatOptions): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const locale = i18n.language?.split('-')[0] || DEFAULT_LANGUAGE;
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    ...options,
  }).format(d);
}

/**
 * Format a time for display
 */
export function formatTime(date: Date | string, options?: Intl.DateTimeFormatOptions): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const locale = i18n.language?.split('-')[0] || DEFAULT_LANGUAGE;
  return new Intl.DateTimeFormat(locale, {
    hour: '2-digit',
    minute: '2-digit',
    ...options,
  }).format(d);
}

/**
 * Format a relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const locale = i18n.language?.split('-')[0] || DEFAULT_LANGUAGE;
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });
  const diffSeconds = Math.round((d.getTime() - now.getTime()) / 1000);
  const absSeconds = Math.abs(diffSeconds);

  if (absSeconds < 60) {
    return rtf.format(diffSeconds, 'second');
  }

  const diffMinutes = Math.round(diffSeconds / 60);
  const absMinutes = Math.abs(diffMinutes);
  if (absMinutes < 60) {
    return rtf.format(diffMinutes, 'minute');
  }

  const diffHours = Math.round(diffMinutes / 60);
  const absHours = Math.abs(diffHours);
  if (absHours < 24) {
    return rtf.format(diffHours, 'hour');
  }

  const diffDays = Math.round(diffHours / 24);
  return rtf.format(diffDays, 'day');
}

/**
 * Format currency amount
 */
export function formatCurrency(amount: number, currency = 'USD'): string {
  const locale = i18n.language?.split('-')[0] || DEFAULT_LANGUAGE;
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    currencyDisplay: 'symbol',
  }).format(amount);
}

/**
 * Generate a unique ID
 */
export function generateId(prefix = 'id'): string {
  return `${prefix}-${Math.random().toString(36).substring(2, 9)}`;
}

/**
 * Debounce a function
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), delay);
  };
}

/**
 * Sleep for a given duration
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
