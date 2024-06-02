-- name: GetPipelineLogs :many
SELECT "order", "cmd", "output", "exit_code"
FROM "pipeline_log"
WHERE "pipeline_id" = $1
ORDER BY "order";

-- name: CreatePipelineLog :one
INSERT INTO "pipeline_log" ("order", "cmd", "output", "exit_code", "pipeline_id")
VALUES ($1, $2, $3, $4, $5)
RETURNING "id";
