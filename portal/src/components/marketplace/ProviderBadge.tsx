'use client';

import type { Provider } from '@/types/offerings';

interface ProviderBadgeProps {
  provider: Provider;
  size?: 'sm' | 'md' | 'lg';
  showReputation?: boolean;
}

export function ProviderBadge({
  provider,
  size = 'md',
  showReputation = true,
}: ProviderBadgeProps) {
  const sizeClasses = {
    sm: 'h-8 w-8 text-sm',
    md: 'h-10 w-10 text-base',
    lg: 'h-14 w-14 text-lg',
  };

  const textSizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  };

  const initial = provider.name.charAt(0).toUpperCase();

  return (
    <div className="flex items-center gap-3">
      <div
        className={`flex items-center justify-center rounded-full bg-primary/10 font-semibold text-primary ${sizeClasses[size]}`}
      >
        {initial}
      </div>
      <div>
        <div className="flex items-center gap-2">
          <span className={`font-medium ${textSizeClasses[size]}`}>
            {provider.name}
          </span>
          {provider.verified && (
            <span
              className="inline-flex items-center rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-300"
              title="Verified Provider"
            >
              ✓ Verified
            </span>
          )}
        </div>
        {showReputation && (
          <div className={`flex items-center gap-1 text-muted-foreground ${textSizeClasses[size]}`}>
            <ReputationStars rating={provider.reputation} />
            <span>({provider.reputation}/100)</span>
          </div>
        )}
      </div>
    </div>
  );
}

function ReputationStars({ rating }: { rating: number }) {
  const fullStars = Math.floor(rating / 20);
  const hasHalfStar = (rating % 20) >= 10;
  const starSlots = [0, 1, 2, 3, 4];

  return (
    <div className="flex items-center">
      {starSlots.map((starIndex) => {
        if (starIndex < fullStars) {
          return (
            <svg
              key={`star-${starIndex}`}
              className="h-4 w-4 text-yellow-500"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
            </svg>
          );
        }
        if (starIndex === fullStars && hasHalfStar) {
          return (
            <svg
              key={`star-${starIndex}`}
              className="h-4 w-4 text-yellow-500"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <defs>
                <linearGradient id={`half-${starIndex}`}>
                  <stop offset="50%" stopColor="currentColor" />
                  <stop offset="50%" stopColor="#d1d5db" />
                </linearGradient>
              </defs>
              <path
                fill={`url(#half-${starIndex})`}
                d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"
              />
            </svg>
          );
        }
        return (
          <svg
            key={`star-${starIndex}`}
            className="h-4 w-4 text-gray-300 dark:text-gray-600"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
          </svg>
        );
      })}
    </div>
  );
}

interface ProviderInfoCardProps {
  provider: Provider;
}

export function ProviderInfoCard({ provider }: ProviderInfoCardProps) {
  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
    });
  };

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <h2 className="mb-4 font-semibold">Provider Information</h2>

      <ProviderBadge provider={provider} size="lg" />

      {provider.description && (
        <p className="mt-4 text-sm text-muted-foreground">{provider.description}</p>
      )}

      <dl className="mt-4 space-y-3 text-sm">
        <div className="flex justify-between">
          <dt className="text-muted-foreground">Total Offerings</dt>
          <dd className="font-medium">{provider.totalOfferings}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-muted-foreground">Total Orders</dt>
          <dd className="font-medium">{provider.totalOrders}</dd>
        </div>
        <div className="flex justify-between">
          <dt className="text-muted-foreground">Member Since</dt>
          <dd className="font-medium">{formatDate(provider.createdAt)}</dd>
        </div>
        {provider.regions && provider.regions.length > 0 && (
          <div>
            <dt className="mb-1 text-muted-foreground">Regions</dt>
            <dd className="flex flex-wrap gap-1">
              {provider.regions.map((region) => (
                <span
                  key={region}
                  className="rounded-full bg-muted px-2 py-0.5 text-xs"
                >
                  {region}
                </span>
              ))}
            </dd>
          </div>
        )}
      </dl>

      {provider.website && (
        <a
          href={provider.website}
          target="_blank"
          rel="noopener noreferrer"
          className="mt-4 inline-flex items-center gap-1 text-sm text-primary hover:underline"
        >
          Visit Website
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
            />
          </svg>
        </a>
      )}

      {/* On-chain address */}
      <div className="mt-4 border-t border-border pt-4">
        <dt className="text-xs text-muted-foreground">Blockchain Address</dt>
        <dd className="mt-1 font-mono text-xs text-muted-foreground">
          {provider.address.slice(0, 20)}...{provider.address.slice(-8)}
        </dd>
      </div>
    </div>
  );
}
