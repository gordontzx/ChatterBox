import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// Defaults to the API running directly on the host (the normal `npm run
// dev` workflow). Override with API_PROXY_TARGET when this dev server runs
// inside a container itself (e.g. the `web` service in docker-compose.yml),
// where `localhost` would otherwise resolve to the container, not the API.
const apiTarget = process.env.API_PROXY_TARGET ?? 'http://localhost:8080'
const wsTarget = apiTarget.replace(/^http/, 'ws')

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/auth': apiTarget,
      '/users': apiTarget,
      '/ws': { target: wsTarget, ws: true },
      '^/chats(\\?.*)?$': apiTarget,
      '^/chats/(direct|group)(\\?.*)?$': apiTarget,
      '^/chats/[^/]+/(messages|members|leave|read|read_receipts)(\\?.*)?$': apiTarget,
    },
  },
  test: {
    environment: 'node',
    setupFiles: ['./src/test/setup.ts'],
  },
})
