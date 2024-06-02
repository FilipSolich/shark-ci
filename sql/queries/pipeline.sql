-- name: GetPipelineCreationInfo :one
SELECT su.username, su.access_token, su.refresh_token, su.token_type, su.token_expire, r.name
FROM "service_user" su JOIN "repo" r ON su.id = r.service_user_id
WHERE r.id = $1;

-- name: GetPipelineStateChangeInfo :one
SELECT p.url, p.commit_sha, p.started_at, r.service, r.owner, r.name, su.access_token, su.refresh_token, su.token_type, su.token_expire
FROM public.pipeline p JOIN public.repo r ON p.repo_id = r.id JOIN public.service_user su ON r.service_user_id = su.id
WHERE p.id = $1;

-- name: CreatePipeline :one
INSERT INTO "pipeline" (status, clone_url, commit_sha, repo_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: SetPipelineUrl :exec
UPDATE "pipeline"
SET url = $1
WHERE id = $2;

-- name: PipelineStarted :exec
UPDATE "pipeline"
SET status = $1, started_at = $2
WHERE id = $3;

-- name: PipelineFinished :exec
UPDATE "pipeline"
SET status = $1, finished_at = $2
WHERE id = $3;
