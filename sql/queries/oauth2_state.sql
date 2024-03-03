-- name: CleanOAuth2State :exec
DELETE FROM public.oauth2_state
WHERE expire < NOW();

-- name: GetAndDeleteOAuth2State :one
DELETE FROM public.oauth2_state
WHERE state = $1
RETURNING expire;

-- name: CreateOAuth2State :exec
INSERT INTO public.oauth2_state (state, expire)
VALUES ($1, $2);
