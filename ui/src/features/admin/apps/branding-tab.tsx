import { useState } from 'react'
import { cn, getInitials } from '@/lib/utils'
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

interface LoginLayoutOptionProps {
  label: string
  description: string
  selected: boolean
  onSelect: () => void
  preview: React.ReactNode
}

function LoginLayoutOption({
  label,
  description,
  selected,
  onSelect,
  preview,
}: LoginLayoutOptionProps) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onSelect}
      className={cn(
        'flex flex-col gap-3 rounded-md border p-4 text-left transition-colors',
        selected ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted',
      )}
    >
      {preview}
      <div>
        <p className="text-sm font-medium text-foreground">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
    </button>
  )
}

// Miniature schematics for the two layouts — matches what login-page.tsx (the
// Centered variant, built) and its Split sibling (Phase 6) actually look
// like, just abstracted to boxes rather than a screenshot.
const centeredPreview = (
  <div className="flex h-16 items-center justify-center rounded-sm border border-border bg-muted">
    <div className="h-8 w-10 rounded-xs border border-border bg-background" />
  </div>
)

const splitPreview = (
  <div className="flex h-16 overflow-hidden rounded-sm border border-border">
    <div className="w-1/2 bg-muted" />
    <div className="flex w-1/2 items-center justify-center bg-background">
      <div className="h-8 w-6 rounded-xs border border-border bg-muted" />
    </div>
  </div>
)

// Logo upload, display name, login layout (UI_PAGES.md §6). Visual-only — no
// real upload/save wiring until lib/client exists (UI_CODING_STANDARDS.md
// §13). Login layout choice pairs with login-page.tsx's own Centered/Split
// variants landing in Phase 6.
export function BrandingTab({ app }: { app: AppRecord }) {
  const [displayName, setDisplayName] = useState(app.name)
  const [loginLayout, setLoginLayout] = useState<AppRecord['loginLayout']>(
    app.loginLayout,
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle>Branding</CardTitle>
        <CardDescription>
          How {app.name} appears on the consent and sign-in screens.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-6">
        <div className="flex items-center gap-4">
          <div className="flex size-16 shrink-0 items-center justify-center rounded-md border border-border bg-muted text-lg font-semibold text-foreground">
            {getInitials(displayName)}
          </div>
          <div className="flex flex-col items-start gap-1.5">
            <Button type="button" variant="outline" size="sm">
              Upload logo
            </Button>
            <p className="text-xs text-muted-foreground">
              PNG or SVG, shown on the consent screen (DATA_MODEL.md apps.logo_url).
            </p>
          </div>
        </div>

        <Field className="max-w-sm">
          <FieldLabel htmlFor="app-display-name">Display name</FieldLabel>
          <Input
            id="app-display-name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
          />
        </Field>

        <div className="flex flex-col gap-2">
          <p className="text-xs font-medium text-muted-foreground">Login layout</p>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <LoginLayoutOption
              label="Centered"
              description="Logo and form centered on the screen."
              preview={centeredPreview}
              selected={loginLayout === 'Centered'}
              onSelect={() => setLoginLayout('Centered')}
            />
            <LoginLayoutOption
              label="Split"
              description="Brand image fills one half, the form fills the other."
              preview={splitPreview}
              selected={loginLayout === 'Split'}
              onSelect={() => setLoginLayout('Split')}
            />
          </div>
        </div>

        <div>
          <Button type="button" size="sm">
            Save changes
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
