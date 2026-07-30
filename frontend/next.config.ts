import type { NextConfig } from 'next';

const apiBase = process.env.BACKEND_URL || 'http://127.0.0.1:8080';

const nextConfig: NextConfig = {
  allowedDevOrigins: ['*.dev.coze.site'],
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: `${apiBase}/api/v1/:path*`,
      },
      {
        source: '/health',
        destination: `${apiBase}/health`,
      },
    ];
  },
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '*',
        pathname: '/**',
      },
    ],
  },
};

export default nextConfig;
