-- name: CreateOauthCall :exec
INSERT INTO "oauth_calls" (app_id, code, user_id, expires_at) VALUES ($1, $2, $3, $4);

-- name: FindOauthCallByCode :one
SELECT * FROM "oauth_calls" WHERE code = $1 AND expires_at > NOW();



