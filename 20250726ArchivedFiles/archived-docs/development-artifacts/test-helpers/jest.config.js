module.exports = {
  testEnvironment: 'jsdom',
  roots: ['<rootDir>/dev/web/static/js'],
  testMatch: [
    '**/__tests__/**/*.js',
    '**/?(*.)+(spec|test).js'
  ],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/dev/web/static/js/$1'
  },
  collectCoverageFrom: [
    'dev/web/static/js/**/*.js',
    '!dev/web/static/js/**/*.test.js',
    '!dev/web/static/js/**/*.spec.js',
    '!dev/web/static/js/__tests__/**',
    '!dev/web/static/js/debug.js'
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  },
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js']
};