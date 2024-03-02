-- name: GetReposByUser :many
SELECT r.id, r.service, r.owner, r.name, r.repo_service_id, r.webhook_id, r.service_user_id
FROM public.repo r JOIN public.service_user su ON r.service_user_id = su.id
WHERE su.user_id = $1;

-- name: GetRegisterWebhookInfoByRepo :one
SELECT r.service, r.owner, r.name, su.access_token, su.refresh_token, su.token_type, su.token_expire
FROM public.repo r JOIN public.service_user su ON r.service_user_id = su.id
WHERE r.id = $1;

-- name: SetRepoWebhook :exec
UPDATE public.repo
SET webhook_id = $1
WHERE id = $2;
