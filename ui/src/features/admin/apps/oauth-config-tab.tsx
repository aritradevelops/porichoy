import { useState } from 'react'
import { EyeIcon, EyeOffIcon, PlusIcon, RotateCwIcon, XIcon } from 'lucide-react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import type { AppRecord } from './apps-data'

// Sharp-radius switch (rounded-xs track/thumb, not rounded-full) — a filled-
// pill toggle would fight the Mono/Sharp token set (UI_CODING_STANDARDS.md
// §5.1's radius axis explicitly rejects Round); this is the same visual
// language as badge.tsx's rounded-pill just carried one step further down
// the radius scale for a control this size. Not a shadcn primitive — Phase 3
// didn't call for adding a Switch to components/ui, so this stays local to
// the one tab that needs it rather than growing the primitive set.
function Toggle({
  checked,
  onCheckedChange,
  disabled,
  label,
}: {
  checked: boolean
  onCheckedChange?: (checked: boolean) => void
  disabled?: boolean
  label: string
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={() => onCheckedChange?.(!checked)}
      className={cn(
        'relative inline-flex h-5 w-9 shrink-0 items-center rounded-xs border transition-colors',
        checked ? 'border-primary bg-primary' : 'border-border bg-muted',
        disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer',
      )}
    >
      <span
        className={cn(
          'inline-block size-3.5 shrink-0 translate-x-0.5 rounded-xs bg-background shadow-sm transition-transform',
          checked && 'translate-x-4',
        )}
      />
    </button>
  )
}

function Chip({
  label,
  selected,
  onToggle,
}: {
  label: string
  selected: boolean
  onToggle: () => void
}) {
  return (
    <button type="button" onClick={onToggle} className="cursor-pointer">
      <Badge
        variant={selected ? 'secondary' : 'outline'}
        className={cn('font-mono text-[11px]', !selected && 'text-muted-foreground')}
      >
        {label}
      </Badge>
    </button>
  )
}

const GRANT_TYPES = ['authorization_code', 'client_credentials', 'refresh_token']
const SCOPES = ['openid', 'profile', 'email', 'offline_access']

// Login methods are tenant-level only — "all apps under a tenant inherit the
// tenant's configured methods exactly, no per-app override" (PRD §5.1).
// UI_PAGES.md §6 nonetheless calls for "login-method toggles" on this tab; the
// reading that's consistent with both is a read-only mirror of what the
// tenant has enabled (rendered disabled, with a link to where they're
// actually configured) rather than a real per-app override control. Flagged
// back to the reviewer/UI_PAGES.md §10 as a judgment call, not a silent
// reinterpretation.
const LOGIN_METHODS = [
  { id: 'email_password', label: 'Email + password' },
  { id: 'otp', label: 'Email / phone OTP' },
  { id: 'webauthn', label: 'WebAuthn' },
  { id: 'google', label: 'Google' },
  { id: 'apple', label: 'Apple' },
] as const

const tenantEnabledLoginMethods = new Set(['email_password', 'google', 'webauthn'])

function formatTtl(seconds: number) {
  if (seconds % 86400 === 0)
    return `≈ ${seconds / 86400} day${seconds / 86400 === 1 ? '' : 's'}`
  if (seconds % 3600 === 0)
    return `≈ ${seconds / 3600} hour${seconds / 3600 === 1 ? '' : 's'}`
  if (seconds % 60 === 0)
    return `≈ ${seconds / 60} minute${seconds / 60 === 1 ? '' : 's'}`
  return `${seconds} seconds`
}

// Client credentials, redirect URIs, grant types/scopes, login methods,
// organizations toggle, token TTLs (UI_PAGES.md §6). Visual-only — nothing
// here persists until lib/client exists (UI_CODING_STANDARDS.md §13).
export function OAuthConfigTab({ app }: { app: AppRecord }) {
  const [secretVisible, setSecretVisible] = useState(false)
  const [redirectUris, setRedirectUris] = useState(app.redirectUris)
  const [grantTypes, setGrantTypes] = useState(new Set(app.grantTypes))
  const [scopes, setScopes] = useState(new Set(app.scopes))
  const [organizationsEnabled, setOrganizationsEnabled] = useState(
    app.supportsOrganizations,
  )

  function toggleInSet(set: Set<string>, value: string) {
    const next = new Set(set)
    if (next.has(value)) next.delete(value)
    else next.add(value)
    return next
  }

  function updateRedirectUri(index: number, value: string) {
    setRedirectUris((uris) => uris.map((uri, i) => (i === index ? value : uri)))
  }

  function removeRedirectUri(index: number) {
    setRedirectUris((uris) => uris.filter((_, i) => i !== index))
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Client credentials</CardTitle>
          <CardDescription>
            Used by {app.name} to authenticate with Porichoy.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <Field className="max-w-md">
            <FieldLabel>Client ID</FieldLabel>
            <Input readOnly value={app.clientId} className="font-mono text-xs" />
          </Field>
          <Field className="max-w-md">
            <FieldLabel>Client secret</FieldLabel>
            <div className="flex items-center gap-2">
              <Input
                readOnly
                type={secretVisible ? 'text' : 'password'}
                value={app.clientSecret}
                className="font-mono text-xs"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={
                  secretVisible ? 'Hide client secret' : 'Reveal client secret'
                }
                onClick={() => setSecretVisible((v) => !v)}
              >
                {secretVisible ? <EyeOffIcon /> : <EyeIcon />}
              </Button>
              <Button type="button" variant="outline" size="sm">
                <RotateCwIcon /> Regenerate
              </Button>
            </div>
          </Field>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Redirect URIs</CardTitle>
          <CardDescription>
            Where Porichoy is allowed to send the user back after sign-in.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {redirectUris.map((uri, index) => (
            <div key={index} className="flex items-center gap-2">
              <Input
                value={uri}
                onChange={(e) => updateRedirectUri(index, e.target.value)}
                className="font-mono text-xs"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label="Remove redirect URI"
                onClick={() => removeRedirectUri(index)}
              >
                <XIcon />
              </Button>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="self-start"
            onClick={() => setRedirectUris((uris) => [...uris, ''])}
          >
            <PlusIcon /> Add redirect URI
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Grant types &amp; scopes</CardTitle>
          <CardDescription>
            Allowed OAuth2 grants and OIDC scopes (PRD §5.2).
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <p className="text-xs font-medium text-muted-foreground">Grant types</p>
            <div className="flex flex-wrap gap-2">
              {GRANT_TYPES.map((grant) => (
                <Chip
                  key={grant}
                  label={grant}
                  selected={grantTypes.has(grant)}
                  onToggle={() => setGrantTypes((s) => toggleInSet(s, grant))}
                />
              ))}
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <p className="text-xs font-medium text-muted-foreground">Scopes</p>
            <div className="flex flex-wrap gap-2">
              {SCOPES.map((scope) => (
                <Chip
                  key={scope}
                  label={scope}
                  selected={scopes.has(scope)}
                  onToggle={() => setScopes((s) => toggleInSet(s, scope))}
                />
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Login methods</CardTitle>
          <CardDescription>
            Configured at the tenant level and inherited by every app under it (PRD
            §5.1) — manage them from{' '}
            <Link
              to="/admin/tenants"
              className="text-primary underline-offset-4 hover:underline"
            >
              Admin → Tenants
            </Link>
            , not here.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {LOGIN_METHODS.map((method) => (
            <div key={method.id} className="flex items-center justify-between gap-4">
              <p className="text-sm text-foreground">{method.label}</p>
              <Toggle
                checked={tenantEnabledLoginMethods.has(method.id)}
                disabled
                label={method.label}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Organizations</CardTitle>
          <CardDescription>
            Whether users sign up individually or as part of an organization (PRD §4.3).
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between gap-4">
            <p className="text-sm text-foreground">Organizations enabled</p>
            <Toggle
              checked={organizationsEnabled}
              onCheckedChange={setOrganizationsEnabled}
              label="Organizations enabled"
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Token &amp; session TTLs</CardTitle>
          <CardDescription>
            Deferred to engineering per PRD §12 — exposed here as plain duration fields,
            nothing fancy.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-6">
          <Field className="w-48">
            <FieldLabel htmlFor="access-token-ttl">
              Access token TTL (seconds)
            </FieldLabel>
            <Input
              id="access-token-ttl"
              type="number"
              defaultValue={app.accessTokenTtlSeconds}
            />
            <p className="text-xs text-muted-foreground">
              {formatTtl(app.accessTokenTtlSeconds)}
            </p>
          </Field>
          <Field className="w-48">
            <FieldLabel htmlFor="refresh-token-ttl">
              Refresh token TTL (seconds)
            </FieldLabel>
            <Input
              id="refresh-token-ttl"
              type="number"
              defaultValue={app.refreshTokenTtlSeconds}
            />
            <p className="text-xs text-muted-foreground">
              {formatTtl(app.refreshTokenTtlSeconds)}
            </p>
          </Field>
        </CardContent>
      </Card>

      <div>
        <Button type="button" size="sm">
          Save changes
        </Button>
      </div>
    </div>
  )
}
