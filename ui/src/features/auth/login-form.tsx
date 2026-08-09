import { zodResolver } from '@hookform/resolvers/zod'
import { FingerprintIcon, MailIcon } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
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
// Only email+password is wired up. The other methods a tenant can enable
// (email/phone OTP, WebAuthn, Google, Apple — PRD §5.1) are non-functional
// placeholders: this can't yet know which methods a tenant actually has
// enabled, since lib/client (the piece that would fetch tenant config and
// call the auth API) doesn't exist yet (UI_CODING_STANDARDS.md §3, §13).
export function LoginForm() {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  })

  function onSubmit(_values: LoginFormValues) {
    // TODO: call lib/client's auth endpoint once it exists. For now this is a
    // structural placeholder — see the module comment above.
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
              <Button type="submit" disabled={isSubmitting}>
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
