// Shared static mock data for the Apps list (3.1), detail shell (3.2), and
// its two tabs (3.3/3.4) — kept in its own module rather than exported off
// apps-list-page.tsx so app-detail-page.tsx can look an app up by :appId
// without one page reaching into another page's internals. Will be replaced
// by the apps list/detail API once lib/client exists (UI_CODING_STANDARDS.md
// §13). Field names mirror DATA_MODEL.md's `apps` table where one exists.
export interface AppRecord {
  id: string
  name: string
  // org-enabled vs. individual sign-up (PRD §4.3, DATA_MODEL.md
  // `apps.supports_organizations`).
  supportsOrganizations: boolean
  status: 'active' | 'disabled'
  createdAt: string
  clientId: string
  clientSecret: string
  redirectUris: string[]
  loginLayout: 'Centered' | 'Split'
  accessTokenTtlSeconds: number
  refreshTokenTtlSeconds: number
  grantTypes: string[]
  scopes: string[]
}

export const apps: AppRecord[] = [
  {
    id: 'crm',
    name: 'Northbridge CRM',
    supportsOrganizations: true,
    status: 'active',
    createdAt: '2025-11-10',
    clientId: 'porichoy_8f2a1c4e9b',
    clientSecret: 'porichoy_secret_51H8f2a1c4e9bZk7yQmP3vLxT',
    redirectUris: [
      'https://crm.northbridge.example/callback',
      'https://crm.northbridge.example/callback/mobile',
    ],
    loginLayout: 'Centered',
    accessTokenTtlSeconds: 900,
    refreshTokenTtlSeconds: 2_592_000,
    grantTypes: ['authorization_code', 'refresh_token'],
    scopes: ['openid', 'profile', 'email'],
  },
  {
    id: 'docuvault',
    name: 'DocuVault',
    supportsOrganizations: false,
    status: 'active',
    createdAt: '2026-01-05',
    clientId: 'porichoy_3d9e77aa21',
    clientSecret: 'porichoy_secret_3d9e77aa21Wp6rY0nJq8sBx',
    redirectUris: ['https://docuvault.example/auth/callback'],
    loginLayout: 'Split',
    accessTokenTtlSeconds: 900,
    refreshTokenTtlSeconds: 1_209_600,
    grantTypes: ['authorization_code'],
    scopes: ['openid', 'profile', 'email'],
  },
  {
    id: 'helpdesk',
    name: 'Helpdesk Pro',
    supportsOrganizations: true,
    status: 'disabled',
    createdAt: '2026-02-19',
    clientId: 'porichoy_a114bd6f02',
    clientSecret: 'porichoy_secret_a114bd6f02Lm4tZ9cRk1uYh',
    redirectUris: ['https://helpdesk.example/callback'],
    loginLayout: 'Centered',
    accessTokenTtlSeconds: 600,
    refreshTokenTtlSeconds: 2_592_000,
    grantTypes: ['authorization_code', 'client_credentials', 'refresh_token'],
    scopes: ['openid', 'profile', 'email', 'offline_access'],
  },
  {
    id: 'internal-tools',
    name: 'Internal Tools',
    supportsOrganizations: false,
    status: 'active',
    createdAt: '2026-05-01',
    clientId: 'porichoy_77c02f3b88',
    clientSecret: 'porichoy_secret_77c02f3b88Qd2wXe5vNp0kT',
    redirectUris: ['https://tools.internal.example/callback'],
    loginLayout: 'Centered',
    accessTokenTtlSeconds: 900,
    refreshTokenTtlSeconds: 604_800,
    grantTypes: ['client_credentials'],
    scopes: ['openid'],
  },
]

export function getAppById(id: string | undefined): AppRecord | undefined {
  return apps.find((app) => app.id === id)
}
