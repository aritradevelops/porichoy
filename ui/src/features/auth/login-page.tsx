import { zodResolver } from '@hookform/resolvers/zod'
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

// Centered layout variant (TECHNICAL_DESIGN §1) — tenant logo + form centered on
// screen. Structure adapted from shadcn's login-03 block (social buttons first,
// Field/FieldGroup for consistent rhythm, FieldSeparator divider), reskinned onto
// our own design tokens (UI_CODING_STANDARDS.md §5.1) rather than shadcn's defaults.
// The Split variant (brand image on one side) is a separate page, not a runtime
// branch of this one, since which variant renders is an admin-time choice
// (tenant.login_layout), not something decided per-request.
//
// Only email+password is wired up. The other methods a tenant can enable
// (email/phone OTP, WebAuthn, Google, Apple — PRD §5.1) are represented below as
// non-functional placeholders: this page can't yet know which methods a tenant
// actually has enabled, since lib/client (the piece that would fetch tenant config
// and call the auth API) doesn't exist yet (UI_CODING_STANDARDS.md §3, §13).
export function LoginPage() {
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
    <div
      className="bg-background flex min-h-svh items-center justify-center p-6"
      style={{
        backgroundImage: 'radial-gradient(var(--border) 1.5px, transparent 1.5px)',
        backgroundSize: '28px 28px',
        backgroundPosition: '-14px -14px',
      }}
    >
      <div className="flex w-full max-w-sm flex-col gap-6">
        {/* Tenant logo placeholder — will render tenant.logo_url once tenant
            config is fetched by lib/client. */}
        <div className="bg-primary text-primary-foreground shadow-[var(--shadow-card)] mx-auto flex h-12 w-12 items-center justify-center rounded-lg text-lg font-semibold">
          P
        </div>

        <Card>
          <CardHeader className="text-center">
            <CardTitle className="text-2xl font-semibold tracking-tight">
              Sign in
            </CardTitle>
            <CardDescription>Choose how you&apos;d like to continue</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} noValidate>
              <FieldGroup>
                <Field className="grid grid-cols-2 gap-3">
                  <Button type="button" variant="outline" disabled>
                    <GoogleIcon />
                    Google
                  </Button>
                  <Button type="button" variant="outline" disabled>
                    <AppleIcon />
                    Apple
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
                  <FieldError
                    errors={errors.password ? [errors.password] : undefined}
                  />
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
      </div>
    </div>
  )
}
