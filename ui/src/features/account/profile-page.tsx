import { useState } from 'react'
import { PencilIcon } from 'lucide-react'
import { getInitials } from '@/lib/utils'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

// Static placeholder profile — will come from GET /me once lib/client exists
// (UI_CODING_STANDARDS.md §13). Identity fields only (PRD §9.1); password and
// sessions live on the Security page instead.
const profile = {
  displayName: 'Jane Doe',
  email: 'jane@example.com',
  emailVerified: true,
  phone: '+1 (555) 019-2834',
  phoneVerified: false,
}

type Verification = 'verified' | 'pending'

interface EditableFieldProps {
  label: string
  value: string
  verification?: Verification
}

// Click-to-edit: no single "Edit profile" form, each field flips into its own
// input independently (UI_PAGES.md §2). Visual-only for Phase 1 — Save just
// exits edit mode without persisting (UI_IMPLEMENTATION_TODO Phase 1.2); once
// lib/client exists this becomes a real mutation, and for email/phone
// specifically the verification badge would flip to "Pending verification"
// until the new value is re-verified (PRD §9.1), not take effect immediately.
function EditableField({ label, value, verification }: EditableFieldProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [draft, setDraft] = useState(value)

  function cancel() {
    setDraft(value)
    setIsEditing(false)
  }

  return (
    <div className="flex items-center justify-between gap-4 border-t border-border py-4 first:border-t-0 first:pt-0">
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium text-muted-foreground">{label}</p>
        {isEditing ? (
          <Input
            autoFocus
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            className="mt-1.5 max-w-sm"
          />
        ) : (
          <div className="mt-1 flex items-center gap-2">
            <p className="text-sm text-foreground">{value}</p>
            {verification && (
              <Badge variant={verification === 'verified' ? 'success' : 'warning'}>
                {verification === 'verified' ? 'Verified' : 'Pending verification'}
              </Badge>
            )}
          </div>
        )}
      </div>

      {isEditing ? (
        <div className="flex shrink-0 items-center gap-2">
          <Button type="button" size="sm" variant="ghost" onClick={cancel}>
            Cancel
          </Button>
          <Button type="button" size="sm" onClick={() => setIsEditing(false)}>
            Save
          </Button>
        </div>
      ) : (
        <Button
          type="button"
          size="icon-sm"
          variant="ghost"
          aria-label={`Edit ${label}`}
          onClick={() => setIsEditing(true)}
        >
          <PencilIcon />
        </Button>
      )}
    </div>
  )
}

export function ProfilePage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Profile
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Identity details for your account. Password and sessions live under Security.
        </p>
      </div>

      <Card>
        <CardContent className="flex flex-col gap-1">
          {/* profile-row (avatar + name + email) — matches the identity
              summary already used in app-shell.tsx's user menu. */}
          <div className="flex items-center gap-4 pb-4">
            <div className="relative">
              <Avatar size="lg">
                <AvatarFallback>{getInitials(profile.displayName)}</AvatarFallback>
              </Avatar>
              {/* DP edit affordance — the avatar's own click-to-edit trigger
                  (PRD §9.1's profile_picture_url). Visual-only, no upload
                  wiring until lib/client exists. */}
              <button
                type="button"
                aria-label="Change profile picture"
                className="absolute -right-1 -bottom-1 flex size-6 items-center justify-center rounded-full border border-border bg-background text-muted-foreground transition-colors hover:text-foreground"
              >
                <PencilIcon className="size-3" />
              </button>
            </div>
            <div className="min-w-0">
              <p className="truncate text-base font-semibold text-foreground">
                {profile.displayName}
              </p>
              <p className="truncate text-sm text-muted-foreground">{profile.email}</p>
            </div>
          </div>

          <EditableField label="Display name" value={profile.displayName} />
          <EditableField
            label="Email address"
            value={profile.email}
            verification={profile.emailVerified ? 'verified' : 'pending'}
          />
          <EditableField
            label="Phone number"
            value={profile.phone}
            verification={profile.phoneVerified ? 'verified' : 'pending'}
          />
        </CardContent>
      </Card>
    </div>
  )
}
