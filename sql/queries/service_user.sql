-- name: CreateServiceUser :one
INSERT INTO public.service_user (service, username, email, access_token, refresh_token, token_type, token_expire, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: GetServiceUserByUserID :one
SELECT id, username, username, email, access_token, refresh_token, token_type, token_expire
FROM public.service_user
WHERE user_id = $1 AND service = $2;

-- name: GetUserID :one
SELECT user_id
FROM public.service_user
WHERE service = $1 AND username = $2;
