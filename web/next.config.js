/** @type {import('next').NextConfig} */
const path = require('path');
const nextConfig = {
  // Static export for Go embedding
  output: 'export',

  // Required for static export
  trailingSlash: true,

  // Disable image optimization for static export
  images: {
    unoptimized: true,
  },

  // TypeScript and ESLint configuration
  typescript: {
    // Temporarily ignore type errors for build
    ignoreBuildErrors: true,
  },

  eslint: {
    // Temporarily ignore ESLint errors for build
    ignoreDuringBuilds: true,
  },

  // Simplified config for stable builds
  poweredByHeader: false,
  reactStrictMode: true, // Enable strict mode to catch React issues early
  swcMinify: true, // Use SWC minification with safer settings

  // Compiler options to prevent TDZ errors
  compiler: {
    removeConsole: false, // Keep for debugging
    styledComponents: false, // Disable styled-components optimization
    emotion: false, // Disable emotion optimization
  },

  // Enhanced webpack config for proper module resolution and TDZ prevention
  webpack: (config, { dev, isServer }) => {
    // Fix module resolution for @/ paths in both server and client bundles
    config.resolve.alias = {
      ...config.resolve.alias,
      '@': path.resolve(__dirname, '.'),
      '@/components': path.resolve(__dirname, './components'),
      '@/lib': path.resolve(__dirname, './lib'),
      '@/types': path.resolve(__dirname, './types'),
      '@/app': path.resolve(__dirname, './app'),
      '@/public': path.resolve(__dirname, './public'),
    }

    // Prevent TDZ errors in production
    if (!dev && !isServer) {
      // Preserve module order to prevent TDZ
      config.optimization = {
        ...config.optimization,
        concatenateModules: false, // Prevent module concatenation that causes TDZ
        splitChunks: {
          ...config.optimization.splitChunks,
          chunks: 'all',
          maxSize: 244000, // Smaller chunks to reduce variable name conflicts
          minSize: 20000,
        },
      }

      // Configure Terser to be safer with variable names
      if (config.optimization.minimizer) {
        config.optimization.minimizer.forEach((minimizer) => {
          if (minimizer.constructor.name === 'TerserPlugin') {
            minimizer.options.terserOptions = {
              ...minimizer.options.terserOptions,
              compress: {
                ...minimizer.options.terserOptions?.compress,
                passes: 1, // Fewer passes to reduce complexity
                reduce_funcs: false, // Prevent function reduction that causes TDZ
                reduce_vars: false, // Prevent variable reduction that causes TDZ
                sequences: false, // Prevent sequence merging that causes TDZ
              },
              mangle: {
                ...minimizer.options.terserOptions?.mangle,
                // Preserve class names to prevent conflicts
                keep_classnames: true,
                // Preserve function names for debugging
                keep_fnames: true,
              },
            }
          }
        })
      }
    }

    return config;
  },
};

// Bundle analyzer for development
if (process.env.ANALYZE === 'true') {
  const withBundleAnalyzer = require('@next/bundle-analyzer')({
    enabled: true,
  });
  module.exports = withBundleAnalyzer(nextConfig);
} else {
  module.exports = nextConfig;
}
