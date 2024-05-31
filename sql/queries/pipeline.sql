-- name: GetPipelineCreationInfo :one
SELECT su.username, su.access_token, su.refresh_token, su.token_type, su.token_expire, r.name
FROM public.service_user su JOIN repo r ON su.id = r.service_user_id
WHERE r.id = $1;

-- name: CreatePipeline :one
INSERT INTO public.pipeline (status, context, clone_url, commit_sha, repo_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: SetPipelineUrl :exec
UPDATE public.pipeline
SET url = $1
WHERE id = $2;

-- name: PipelineStarted :exec
UPDATE public.pipeline
SET status = $1, started_at = $2
WHERE id = $3;

-- name: PipelineFinished :exec
UPDATE public.pipeline
SET status = $1, finished_at = $2
WHERE id = $3;
