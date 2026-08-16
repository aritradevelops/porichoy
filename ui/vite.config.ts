/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  server: {
    // Login/signup resolve their tenant purely off the request's Host header
    // (TenantResolution, server/internal/adapters/rest/middleware.go) — local dev has to be
    // reached via a hostname that's actually registered as a tenant domain via
    // POST /api/v1/domains/register. admin.localhost resolves to loopback with no /etc/hosts
    // edit (RFC 6761 reserves *.localhost).
    allowedHosts: ['admin.localhost'],
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        // changeOrigin left at its default (false): preserves the browser's original Host
        // header through to the backend, which is what TenantResolution resolves the tenant
        // from. Setting this to true would rewrite Host to the proxy target's own address and
        // every request would 404 as tenant.unresolved_domain.
      },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
  },
})
