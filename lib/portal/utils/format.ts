/**
 * Format Utilities
 * VE-700 - VE-705: Display formatting helpers
 */

/**
 * Format identity score for display
 */
export function formatScore(score: number): string {
  return score.toFixed(0);
}

/**
 * Format token amount with proper decimals
 */
export function formatTokenAmount(
  amount: string,
  decimals: number = 6,
  symbol: string = 'VE'
): string {
  const value = parseInt(amount, 10);
  if (isNaN(value)) return `0 ${symbol}`;

  const divisor = Math.pow(10, decimals);
  const formatted = (value / divisor).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: decimals,
  });

  return `${formatted} ${symbol}`;
}

/**
 * Format duration in seconds to human-readable string
 */
export function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${seconds} second${seconds !== 1 ? 's' : ''}`;
  }

  if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60);
    return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
  }

  if (seconds < 86400) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (minutes === 0) {
      return `${hours} hour${hours !== 1 ? 's' : ''}`;
    }
    return `${hours}h ${minutes}m`;
  }

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  if (hours === 0) {
    return `${days} day${days !== 1 ? 's' : ''}`;
  }
  return `${days}d ${hours}h`;
}

/**
 * Format timestamp to locale string
 */
export function formatTimestamp(
  timestamp: number,
  options: Intl.DateTimeFormatOptions = {
    dateStyle: 'medium',
    timeStyle: 'short',
  }
): string {
  // Handle both seconds and milliseconds
  const ms = timestamp > 10000000000 ? timestamp : timestamp * 1000;
  return new Date(ms).toLocaleString(undefined, options);
}

/**
 * Format relative time (e.g., "5 minutes ago")
 */
export function formatRelativeTime(timestamp: number): string {
  const ms = timestamp > 10000000000 ? timestamp : timestamp * 1000;
  const diff = Date.now() - ms;

  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) {
    return 'just now';
  }

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) {
    return `${minutes} minute${minutes !== 1 ? 's' : ''} ago`;
  }

  const hours = Math.floor(minutes / 60);
  if (hours < 24) {
    return `${hours} hour${hours !== 1 ? 's' : ''} ago`;
  }

  const days = Math.floor(hours / 24);
  if (days < 30) {
    return `${days} day${days !== 1 ? 's' : ''} ago`;
  }

  const months = Math.floor(days / 30);
  return `${months} month${months !== 1 ? 's' : ''} ago`;
}

/**
 * Format address for display (truncated)
 */
export function formatAddress(address: string, chars: number = 8): string {
  if (address.length <= chars * 2 + 3) {
    return address;
  }
  return `${address.slice(0, chars)}...${address.slice(-chars)}`;
}

/**
 * Format bytes to human-readable size
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

/**
 * Format percentage
 */
export function formatPercent(value: number, decimals: number = 1): string {
  return `${value.toFixed(decimals)}%`;
}

/**
 * Format hash for display (truncated)
 */
export function formatHash(hash: string, chars: number = 8): string {
  if (!hash.startsWith('0x')) {
    hash = '0x' + hash;
  }
  if (hash.length <= chars * 2 + 5) {
    return hash;
  }
  return `${hash.slice(0, chars + 2)}...${hash.slice(-chars)}`;
}
