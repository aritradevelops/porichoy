import type { LucideIcon } from 'lucide-react'
import { KeyRoundIcon, MonitorIcon, SmartphoneIcon } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ListRow } from '@/components/list-row'

// Three fixed sections: Password, MFA methods, Active sessions
// (UI_PAGES.md §3). Static placeholder data throughout — nothing here is
// wired to lib/client yet (UI_CODING_STANDARDS.md §13).

// Only shown for accounts using the email+password login method (PRD §9.3) —
// assumed true for this placeholder account; would gate on the real login
// method once lib/client exists.
const password = {
  lastChanged: '4 months ago',
}

interface MfaMethod {
  id: string
  label: string
  icon: LucideIcon
  addedOn: string
  lastUsed: string
  isPrimary: boolean
}

// Mocked with a single enrolled method specifically so the force-MFA
// disabled-remove state below is actually visible (PRD §9.2 — "can't drop
// below one"); a real account can have several simultaneously, e.g. a TOTP
// app plus a WebAuthn key.
const mfaMethods: MfaMethod[] = [
  {
    id: 'totp',
    label: 'Authenticator app (TOTP)',
    icon: KeyRoundIcon,
    addedOn: '2026-03-12',
    lastUsed: '2 hours ago',
    isPrimary: true,
  },
]

// Tenant-level setting (PRD §9.2) — static placeholder until lib/client can
// fetch the real tenant config.
const tenantForcesMfa = true

interface Session {
  id: string
  label: string
  icon: LucideIcon
  location: string
  lastActive: string
  isCurrent: boolean
}

const sessions: Session[] = [
  {
    id: 'session-1',
    label: 'Chrome on macOS',
    icon: MonitorIcon,
    location: 'Kolkata, IN',
    lastActive: 'Active now',
    isCurrent: true,
  },
  {
    id: 'session-2',
    label: 'Safari on iPhone',
    icon: SmartphoneIcon,
    location: 'Kolkata, IN',
    lastActive: 'Last active 2 hours ago',
    isCurrent: false,
  },
  {
    id: 'session-3',
    label: 'Firefox on Windows',
    icon: MonitorIcon,
    location: 'Dhaka, BD',
    lastActive: 'Last active 3 days ago',
    isCurrent: false,
  },
]

export function SecurityPage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Security
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Password, MFA methods, and active sessions for your account.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Password</CardTitle>
          <CardDescription>Last changed {password.lastChanged}.</CardDescription>
          <CardAction>
            <Button type="button" variant="outline">
              Change password
            </Button>
          </CardAction>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>MFA methods</CardTitle>
          <CardDescription>
            {tenantForcesMfa
              ? 'Your tenant requires at least one active method.'
              : 'Add a method for step-up verification at sign-in.'}
          </CardDescription>
          <CardAction>
            <Button type="button">Add method</Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          {mfaMethods.map((method) => {
            const isOnlyMethod = mfaMethods.length === 1
            const removeDisabled = tenantForcesMfa && isOnlyMethod

            const removeButton = (
              <Button
                type="button"
                size="sm"
                variant="destructive"
                disabled={removeDisabled}
              >
                Remove
              </Button>
            )

            return (
              <ListRow
                key={method.id}
                icon={method.icon}
                title={
                  <>
                    {method.label}
                    {method.isPrimary && <Badge variant="secondary">Primary</Badge>}
                  </>
                }
                subtext={`Added ${method.addedOn} · last used ${method.lastUsed}`}
                action={
                  removeDisabled ? (
                    <span title="Your tenant requires at least one active MFA method — add another before removing this one.">
                      {removeButton}
                    </span>
                  ) : (
                    removeButton
                  )
                }
              />
            )
          })}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Active sessions</CardTitle>
          <CardDescription>
            Logging out a session signs that device out immediately.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {sessions.map((session) => (
            <ListRow
              key={session.id}
              icon={session.icon}
              title={
                <>
                  {session.label}
                  {session.isCurrent && <Badge variant="outline">This device</Badge>}
                </>
              }
              subtext={`${session.location} · ${session.lastActive}`}
              action={
                session.isCurrent ? undefined : (
                  <Button type="button" size="sm" variant="outline">
                    Log out
                  </Button>
                )
              }
            />
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
