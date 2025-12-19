-- name: CreateOauthInfo :exec
INSERT INTO "oauth_configs" (
  client_secret, redirect_uris, success_callback_url, error_callback_url, jwt_algo, jwt_secret_resolver, app_id,
  created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
);