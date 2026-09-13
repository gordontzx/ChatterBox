import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/auth': 'http://localhost:8080',
      '/users': 'http://localhost:8080',
      '/ws': { target: 'ws://localhost:8080', ws: true },
      '^/chats(\\?.*)?$': 'http://localhost:8080',
      '^/chats/(direct|group)(\\?.*)?$': 'http://localhost:8080',
      '^/chats/[^/]+/(messages|members|leave)(\\?.*)?$': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'node',
    setupFiles: ['./src/test/setup.ts'],
  },
})
