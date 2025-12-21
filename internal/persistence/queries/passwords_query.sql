-- name: CreatePasswordForUser :exec
INSERT INTO "passwords" (
  hashed_password, created_by
) VALUES (
  $1, $2
);

-- name: FindUserPassword :one
SELECT * FROM "passwords" WHERE created_by = $1 AND deleted_at IS NULL;