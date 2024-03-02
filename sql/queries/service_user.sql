-- name: GetServiceUserIDsByServiceUsername :one
SELECT id, user_id
FROM public.service_user
WHERE username = $1 AND service = $2;
