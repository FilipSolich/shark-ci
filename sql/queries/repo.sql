-- name: GetUserRepos :many
SELECT r.id, r.service, r.owner, r.name, r.repo_service_id, r.webhook_id, r.service_user_id
FROM "repo" r JOIN "service_user" su ON r.service_user_id = su.id
WHERE su.user_id = $1;

-- name: CreateRepo :one
INSERT INTO "repo" (service, owner, name, repo_service_id, webhook_id, service_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: DeleteRepo :exec
DELETE FROM "repo"
WHERE id = $1;
