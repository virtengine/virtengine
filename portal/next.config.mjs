/** @type {import('next').NextConfig} */
const isPages = process.env.GITHUB_PAGES === 'true';

const nextConfig = {
  reactStrictMode: true,

  // Enable static export for GitHub Pages deployment
  output: isPages ? 'export' : undefined,

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
