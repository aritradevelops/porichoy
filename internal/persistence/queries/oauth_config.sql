-- name: CreateOauthInfo :exec
INSERT INTO "oauth_configs" (
  client_secret, redirect_uris, success_callback_url, error_callback_url, 
  jwt_algo, jwt_secret_resolver, jwt_lifetime, refresh_token_lifetime, app_id,
  created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);