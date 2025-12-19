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
