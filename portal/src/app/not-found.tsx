'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';

/**
 * Custom 404 page that handles client-side routing for GitHub Pages.
 *
 * When a user navigates directly to a dynamic route (e.g., /marketplace/provider/1),
 * GitHub Pages will serve a 404 since no static file exists. This component
 * attempts to render the route client-side if the path looks valid.
 */
export default function NotFound() {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    // Check if this looks like a valid dynamic route
    const dynamicRoutePatterns = [
      /^\/marketplace\/[^/]+\/[^/]+\/?$/, // /marketplace/[provider]/[sequence]
      /^\/orders\/[^/]+\/?$/, // /orders/[id]
      /^\/provider\/orders\/[^/]+\/?$/, // /provider/orders/[id]
      /^\/governance\/proposals\/[^/]+\/?$/, // /governance/proposals/[id]
    ];

    const isValidDynamicRoute = dynamicRoutePatterns.some((pattern) => pattern.test(pathname));

    if (isValidDynamicRoute) {
      // Force a client-side navigation to the same path
      // This allows the Next.js router to handle it properly
      router.replace(pathname);
    }
  }, [pathname, router]);

  return (
    <div className="container flex min-h-[60vh] flex-col items-center justify-center py-16 text-center">
      <h1 className="text-6xl font-bold text-primary">404</h1>
      <h2 className="mt-4 text-2xl font-semibold">Page Not Found</h2>
      <p className="mt-2 text-muted-foreground">
        The page you&apos;re looking for doesn&apos;t exist or has been moved.
      </p>
      <div className="mt-8 flex gap-4">
        <Link
          href="/"
          className="rounded-lg bg-primary px-6 py-3 font-medium text-primary-foreground hover:bg-primary/90"
        >
          Go Home
        </Link>
        <Link
          href="/marketplace"
          className="rounded-lg border border-border px-6 py-3 font-medium hover:bg-accent"
        >
          Browse Marketplace
        </Link>
      </div>
    </div>
  );
}
