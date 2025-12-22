-- name: RegisterUser :one
INSERT INTO "users" (
  id, email, name, created_by
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM "users" WHERE email = $1 AND deleted_at IS NULL;

-- name: FindUserByID :one
SELECT * FROM "users" WHERE id = $1 AND deleted_at IS NULL;
