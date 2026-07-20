const apiTarget =
  process.env.RETAILPULSE_API_INTERNAL_URL ||
  process.env.NEXT_PUBLIC_API_BASE_URL ||
  'http://localhost:4005';

/** @type {import('next').NextConfig} */
const nextConfig = {
  async rewrites() {
    return [
      {
        source: '/v1/:path*',
        destination: `${apiTarget.replace(/\/$/, '')}/v1/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
