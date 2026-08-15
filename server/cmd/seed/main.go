// Command seed is the one-time CLI bootstrap step
// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1): creates the root tenant, its default system
// app, the three default roles (Super Admin, Tenant Admin, User), and the initial root
// superadmin user. Run once against a fresh deployment, before anyone can log in — the
// unified UI itself requires an authenticated tenant session to render past the login
// screen, so this first account can't be created through the normal signup flow.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aritradevelops/porichoy/server/config"
	"github.com/aritradevelops/porichoy/server/internal/adapters/cache"
	"github.com/aritradevelops/porichoy/server/internal/adapters/crypto"
	"github.com/aritradevelops/porichoy/server/internal/adapters/postgres"
	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
)

// rootTenantExists checks for any non-deleted tenant with no parent — a direct query against
// the tenants table rather than going through tenant.Repository, since this is a one-off CLI
// bootstrap check, not a domain operation any other caller needs.
func rootTenantExists(ctx context.Context, db *bun.DB) (bool, error) {
	return db.NewSelect().
		Table("tenants").
		Where("parent_id IS NULL").
		Where("deleted_at IS NULL").
		Exists(ctx)
}

// superAdminPermissions/tenantAdminPermissions enumerate every module this codebase
// currently exposes — AUTHORIZATION_MODEL.md §5 forbids a module-level "*:*@scope" wildcard,
// each module's action-wildcard has to be granted explicitly. This list is provisional: it
// grows as new modules ship, and nothing enforces it yet (internal/authorization's runtime
// permission check doesn't exist). The User role's baseline stays empty (PRD §7.2).
var (
	superAdminPermissions = []string{
		"tenants:*@root",
		"domains:*@root",
		"provider_credentials:*@root",
	}
	tenantAdminPermissions = []string{
		"tenants:*@tenant",
		"domains:*@tenant",
		"provider_credentials:*@tenant",
	}
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db := postgres.Open(cfg.DBDSN)
	defer db.Close()

	if err := postgres.Migrate(db.DB, cfg.MigrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// This command is meant to run exactly once per deployment. Without this guard, a
	// partial-failure rerun (or an operator running it twice by mistake) would silently
	// create a second, independent root tenant + system app + roles + superadmin instead of
	// erroring — there's no uniqueness constraint on tenants.parent_id IS NULL to catch it at
	// the database level, so it's checked explicitly here before anything is created.
	if exists, err := rootTenantExists(context.Background(), db); err != nil {
		log.Fatalf("check for existing root tenant: %v", err)
	} else if exists {
		log.Fatal("a root tenant already exists — this command is meant to run once, on a fresh deployment")
	}

	tenantSvc := tenant.NewService(
		postgres.NewTenantRepository(db),
		postgres.NewDomainRepository(db),
		postgres.NewProviderCredentialRepository(db),
	)
	appSvc := app.NewService(postgres.NewAppRepository(db))
	// redis.NewClient connects lazily; this bootstrap command never calls
	// authorization.Service's cache-writing methods, so a Redis outage doesn't affect it.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()
	authzSvc := authorization.NewService(postgres.NewRoleRepository(db), postgres.NewRoleAssignmentRepository(db), cache.NewRedisCache(redisClient))
	identitySvc := identity.NewService(
		postgres.NewUserRepository(db),
		postgres.NewPasswordRepository(db),
		postgres.NewAppRepository(db),
		postgres.NewSessionRepository(db),
		postgres.NewRoleAssignmentRepository(db),
		crypto.NewIssuer(),
		postgres.NewTxRunner(db),
	)

	r := bufio.NewReader(os.Stdin)
	fmt.Println("Porichoy bootstrap — creates the root tenant and its initial superadmin.")
	tenantName, err := promptString(r, "Root tenant name")
	if err != nil {
		log.Fatalf("read tenant name: %v", err)
	}
	email, err := promptEmail(r, "Superadmin email")
	if err != nil {
		log.Fatalf("read email: %v", err)
	}
	password, err := promptPassword("Superadmin password")
	if err != nil {
		log.Fatalf("read password: %v", err)
	}

	// The whole bootstrap sequence below runs as one Bun transaction (tx.go's ambient-tx
	// mechanism, already used by identity.Service's own Signup/CreateRootUser) — seven
	// writes across four Services, none of which is independently meaningful without the
	// rest (a root tenant with no system app, or a superadmin with no role, isn't a usable
	// starting state). Any failure partway rolls everything back, so a rerun starts from a
	// genuinely clean slate instead of colliding with leftover partial state.
	var rootTenant *tenant.Tenant
	var sysApp *app.App
	var userRole *authorization.Role
	var rootUser *identity.User

	err = postgres.NewTxRunner(db).RunInTx(context.Background(), func(ctx context.Context) error {
		var err error

		rootTenant, err = tenantSvc.CreateRootTenant(ctx, tenantName)
		if err != nil {
			return fmt.Errorf("create root tenant: %w", err)
		}

		sysApp, err = appSvc.CreateSystemApp(ctx, rootTenant.ID, "System")
		if err != nil {
			return fmt.Errorf("create system app: %w", err)
		}

		superAdminRole, err := authzSvc.CreateSystemRole(ctx, rootTenant.ID, sysApp.ID, "Super Admin", superAdminPermissions)
		if err != nil {
			return fmt.Errorf("create Super Admin role: %w", err)
		}
		if _, err := authzSvc.CreateSystemRole(ctx, rootTenant.ID, sysApp.ID, "Tenant Admin", tenantAdminPermissions); err != nil {
			return fmt.Errorf("create Tenant Admin role: %w", err)
		}
		userRole, err = authzSvc.CreateSystemRole(ctx, rootTenant.ID, sysApp.ID, "User", nil)
		if err != nil {
			return fmt.Errorf("create User role: %w", err)
		}

		if err := appSvc.SetDefaultSignupRole(ctx, sysApp.ID, userRole.ID); err != nil {
			return fmt.Errorf("set default signup role: %w", err)
		}

		rootUser, err = identitySvc.CreateRootUser(ctx, rootTenant.ID, email, password)
		if err != nil {
			return fmt.Errorf("create root superadmin: %w", err)
		}
		if _, err := authzSvc.AssignSystemRole(ctx, rootUser.ID, superAdminRole.ID); err != nil {
			return fmt.Errorf("assign Super Admin role: %w", err)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("bootstrap failed, rolled back: %v", err)
	}

	fmt.Println()
	fmt.Println("Bootstrap complete.")
	fmt.Printf("  Root tenant: %s (%s)\n", rootTenant.Name, rootTenant.ID)
	fmt.Printf("  System app:  %s (client_id %s)\n", sysApp.Name, sysApp.ClientID)
	fmt.Printf("  Superadmin:  %s (%s)\n", email, rootUser.ID)
	fmt.Println()
	fmt.Println("Next: log in via POST /api/v1/auth/login, then pass the returned access_token")
	fmt.Println("as \"Authorization: Bearer <token>\" (or an access_token cookie) to register a")
	fmt.Println("domain for this tenant via POST /api/v1/domains/register. Authorization itself")
	fmt.Println("is still a stub — pass X-Debug-Scope to exercise a scope other than \"tenant\",")
	fmt.Println("per internal/adapters/rest/middleware.go, until real permission checks land.")
}
