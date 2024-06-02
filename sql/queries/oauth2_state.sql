-- name: CleanOAuth2State :exec
DELETE FROM "oauth2_state"
WHERE expire <= now();

-- name: GetAndDeleteOAuth2State :one
DELETE FROM "oauth2_state"
WHERE state = $1
RETURNING expire;

-- name: CreateOAuth2State :exec
INSERT INTO "oauth2_state" (state, expire)
VALUES ($1, $2);
