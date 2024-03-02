-- name: CreatePipeline :exec
INSERT INTO public.pipeline (status,  clone_url, commit_sha, repo_id)
VALUES ($1, $2, $3, $4)
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
Update public.pipeline
SET status = $1, finished_at = $2
WHERE id = $3;
