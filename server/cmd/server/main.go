// Command server is Porichoy's REST API entrypoint: loads config, connects to Postgres,
// runs pending migrations, wires the tenant/identity bounded contexts' Services onto their
// Postgres repositories, and starts the REST adapter's Fiber app (CODING_STANDARDS.md §5).
package main

import (
	"log"

	"github.com/aritradevelops/porichoy/server/config"
	"github.com/aritradevelops/porichoy/server/internal/adapters/cache"
	"github.com/aritradevelops/porichoy/server/internal/adapters/crypto"
	"github.com/aritradevelops/porichoy/server/internal/adapters/postgres"
	"github.com/aritradevelops/porichoy/server/internal/adapters/rest"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/redis/go-redis/v9"
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

	// redis.NewClient connects lazily — this succeeds even if Redis isn't reachable yet;
	// the first cache write (Login) is where an outage would actually surface.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	tenantSvc := tenant.NewService(
		postgres.NewTenantRepository(db),
		postgres.NewDomainRepository(db),
		postgres.NewProviderCredentialRepository(db),
	)
	identitySvc := identity.NewService(
		postgres.NewUserRepository(db),
		postgres.NewPasswordRepository(db),
		postgres.NewAppRepository(db),
		postgres.NewSessionRepository(db),
		postgres.NewRoleAssignmentRepository(db),
		crypto.NewIssuer(),
		postgres.NewTxRunner(db),
	)
	authzSvc := authorization.NewService(
		postgres.NewRoleRepository(db),
		postgres.NewRoleAssignmentRepository(db),
		cache.NewRedisCache(redisClient),
	)

	app := rest.New(tenantSvc, identitySvc, authzSvc)
	log.Printf("porichoy: listening on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
