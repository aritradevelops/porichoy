import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

// Reusable list-row: leading icon, title (+ optional inline badge/pill),
// subtext, right-aligned action slot. Mirrors the shell mockup's `.list-row`
// (icon square + list-main + list-actions). Used by Security's MFA methods
// and Active sessions, and Admin's Tenants/Apps tables (UI_IMPLEMENTATION_TODO
// Phase 1-3) — not a shadcn/ui primitive, so it lives alongside app-shell.tsx
// rather than in components/ui/.
interface ListRowProps {
  icon?: LucideIcon
  title: React.ReactNode
  subtext?: React.ReactNode
  action?: React.ReactNode
  className?: string
}

export function ListRow({
  icon: Icon,
  title,
  subtext,
  action,
  className,
}: ListRowProps) {
  return (
    <div
      className={cn(
        'flex flex-wrap items-center justify-between gap-3 border-t border-border py-3 first:border-t-0 first:pt-0',
        className,
      )}
    >
      <div className="flex min-w-0 items-center gap-3">
        {Icon && (
          <div className="flex size-9 shrink-0 items-center justify-center rounded-md border border-border bg-muted text-muted-foreground">
            <Icon className="size-4" />
          </div>
        )}
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-sm font-medium text-foreground">
            {title}
          </div>
          {subtext && (
            <div className="text-xs text-muted-foreground truncate">{subtext}</div>
          )}
        </div>
      </div>
      {action && <div className="flex shrink-0 items-center gap-2">{action}</div>}
    </div>
  )
}
