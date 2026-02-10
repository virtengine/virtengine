import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import('next').NextConfig} */
const isPages = process.env.GITHUB_PAGES === 'true';

const isDocker = process.env.DOCKER_BUILD === 'true';

const nextConfig = {
  reactStrictMode: true,

  // Enable static export for GitHub Pages, standalone for Docker
  output: isPages ? 'export' : isDocker ? 'standalone' : undefined,

  // Use trailing slashes for better GitHub Pages compatibility
  trailingSlash: isPages,

  // Base path for GitHub Pages (repo name)
  basePath: isPages ? '/virtengine' : '',

  // Asset prefix for GitHub Pages (same as basePath, no trailing slash)
  assetPrefix: isPages ? '/virtengine' : '',

  // Disable image optimization for static export
  images: isPages
    ? { unoptimized: true }
    : {
        remotePatterns: [
          {
            protocol: 'https',
            hostname: 'rpc.virtengine.com',
          },
          {
            protocol: 'https',
            hostname: '*.virtengine.io',
          },
        ],
      },

  transpilePackages: ['virtengine-portal-lib', 'virtengine-capture-lib'],

  experimental: {
    // typedRoutes: true, // Re-enable when all routes are complete
  },

  webpack: (config, { isServer }) => {
    config.resolve.alias = {
      ...(config.resolve.alias ?? {}),
      '@virtengine/portal': path.resolve(__dirname, '../lib/portal'),
      '@virtengine/capture': path.resolve(__dirname, '../lib/capture'),
    };

    // Handle SVG imports
    config.module.rules.push({
      test: /\.svg$/,
      use: ['@svgr/webpack'],
    });

    return config;
  },

  headers: async () => {
    if (isPages) {
      return [];
    }
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          {
            key: 'Permissions-Policy',
            value: 'camera=(self), microphone=(), geolocation=()',
          },
        ],
      },
    ];
  },
};

export default nextConfig;
