import { createMDX } from 'fumadocs-mdx/next';

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  reactStrictMode: true,
  serverExternalPackages: ['typescript', 'twoslash'],
  redirects: async () => [
    {
      source: '/docs',
      destination: '/docs/overview',
      permanent: true,
    },
    {
      source: '/docs/api-reference',
      destination: '/docs/api-reference/introduction',
      permanent: true,
    },
  ],
};

export default withMDX(config);
