-- name: CreateSession :exec
INSERT INTO "sessions" (
  "user_id", "app_id", "refresh_token", "user_ip", "user_agent", "expires_at", "created_by"
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
);
-- name: FindSessionByRefreshTokenAndAppID :one
SELECT * FROM "sessions" WHERE "refresh_token" = $1 AND "app_id" = $2 AND expires_at > CURRENT_TIMESTAMP AND "deleted_at" IS NULL;