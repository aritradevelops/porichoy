-- name: CreateApp :one
INSERT INTO "apps" (
  name, domain, landing_url, logo, client_id, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: FindAppByClientID :one
SELECT sqlc.embed(app), sqlc.embed(oauth_config) FROM "apps" AS app
LEFT JOIN "oauth_configs" AS oauth_config ON app.id = oauth_config.app_id
WHERE app.client_id = $1 AND app.deleted_by IS NULL;

-- TODO: find some other way of finding the root app
-- name: FindRootApp :one
SELECT sqlc.embed(app), sqlc.embed(oauth_config) FROM "apps" AS app
LEFT JOIN "oauth_configs" AS oauth_config ON app.id = oauth_config.app_id
WHERE app.created_by = '00000000-0000-0000-0000-000000000000' AND app.deleted_by IS NULL;
