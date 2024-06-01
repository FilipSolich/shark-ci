-- name: GetUser :one
SELECT id, username, email
FROM "user"
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO "user" (username, email)
VALUES ($1, $2)
RETURNING id;
