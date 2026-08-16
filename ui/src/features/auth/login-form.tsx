import { zodResolver } from '@hookform/resolvers/zod'
import { FingerprintIcon, MailIcon } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { useLocation, useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { ApiError } from '@/lib/client/http'
import { useAuth } from '@/lib/client/auth-context'
import { useLoginMutation } from '@/lib/client/auth'
import type { LocationWithFrom } from '@/lib/client/require-auth'
import { AppleIcon, GoogleIcon } from '@/components/brand-icons'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'

const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
})

type LoginFormValues = z.infer<typeof loginSchema>

// The sign-in Card itself — social/challenge buttons, email/password fields,
// forgot-password and sign-up links. Extracted out of login-page.tsx
// (UI_IMPLEMENTATION_TODO Phase 6.2) so it can be shared verbatim between the
// Centered layout (login-page.tsx) and the new Split layout
// (login-split-page.tsx): UI_PAGES.md §9 is explicit that the two variants
// differ only in the chrome around this form, not the form itself. Structure
// adapted from shadcn's login-03 block, reskinned onto our own design tokens
// (UI_CODING_STANDARDS.md §5.1).
//
// Only email+password is wired up, against POST /api/v1/auth/login (lib/client/auth.ts) — the
// only login method identity.Service implements server-side this pass. The other methods a
// tenant can enable (email/phone OTP, WebAuthn, Google, Apple — PRD §5.1) stay non-functional
// placeholders: this can't yet know which methods a tenant actually has enabled, since nothing
// fetches tenant config to check (no endpoint exposes it — see domain_handlers.go's Resolve).
export function LoginForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  })
  const { login } = useAuth()
  const loginMutation = useLoginMutation()
  const navigate = useNavigate()
  const location = useLocation() as LocationWithFrom

  async function onSubmit(values: LoginFormValues) {
    try {
      const auth = await loginMutation.mutateAsync(values)
      login(auth)
      navigate(location.state?.from?.pathname ?? '/', { replace: true })
    } catch {
      // loginMutation.error already carries the failure — surfaced below, nothing more to do.
    }
  }

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle className="text-2xl font-semibold tracking-tight">Sign in</CardTitle>
        <CardDescription>Choose how you&apos;d like to continue</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <FieldGroup>
            {/* 2x2 grid, not a single row of four — third-party identity
                providers on top (Google, Apple), challenge-based methods
                below (OTP, WebAuthn). All four are equally non-functional
                placeholders (gated on lib/client, same as before), the
                grouping is purely about what reads coherently at four items
                rather than two (UI_IMPLEMENTATION_TODO Phase 6.1). */}
            <Field className="grid grid-cols-2 gap-3">
              <Button type="button" variant="outline" disabled>
                <GoogleIcon />
                Google
              </Button>
              <Button type="button" variant="outline" disabled>
                <AppleIcon />
                Apple
              </Button>
              <Button type="button" variant="outline" disabled>
                <MailIcon />
                Send code
              </Button>
              <Button type="button" variant="outline" disabled>
                <FingerprintIcon />
                Use security key
              </Button>
            </Field>

            <FieldSeparator>or continue with</FieldSeparator>

            <Field>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                aria-invalid={!!errors.email}
                {...register('email')}
              />
              <FieldError errors={errors.email ? [errors.email] : undefined} />
            </Field>

            <Field>
              <div className="flex items-baseline justify-between">
                <FieldLabel htmlFor="password">Password</FieldLabel>
                <a
                  href="#"
                  className="text-muted-foreground hover:text-foreground text-xs transition-colors"
                >
                  Forgot password?
                </a>
              </div>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                aria-invalid={!!errors.password}
                {...register('password')}
              />
              <FieldError errors={errors.password ? [errors.password] : undefined} />
            </Field>

            <Field>
              {loginMutation.error instanceof ApiError && (
                <p className="text-destructive text-center text-sm" role="alert">
                  {loginMutation.error.message}
                </p>
              )}
              <Button type="submit" disabled={isSubmitting || loginMutation.isPending}>
                Sign in
              </Button>
              <FieldDescription className="text-center">
                Don&apos;t have an account? <a href="#">Sign up</a>
              </FieldDescription>
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
