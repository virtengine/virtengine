import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  
  // Enable static export for GitHub Pages deployment
  output: process.env.GITHUB_PAGES === 'true' ? 'export' : undefined,
  
  // Use trailing slashes for better GitHub Pages compatibility
  trailingSlash: process.env.GITHUB_PAGES === 'true' ? true : false,
  
  // Base path for GitHub Pages (repo name)
  basePath: process.env.GITHUB_PAGES === 'true' ? '/virtengine' : '',
  
  // Asset prefix for GitHub Pages (same as basePath, no trailing slash)
  assetPrefix: process.env.GITHUB_PAGES === 'true' ? '/virtengine' : '',
  
  // Disable image optimization for static export
  images: process.env.GITHUB_PAGES === 'true' 
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
  
  transpilePackages: [
    'virtengine-portal-lib',
    'virtengine-capture-lib',
  ],

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
