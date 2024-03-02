-- name: CleanOAuth2State :exec
DELETE FROM public.oauth2_state
WHERE expire < NOW();

-- name: GetOAuth2StateExpiration :one
SELECT expire
FROM public.oauth2_state
WHERE state = $1;

-- name: CreateOAuth2State :exec
INSERT INTO public.oauth2_state (state, expire)
VALUES ($1, $2);

-- name: DeleteOAuth2State :exec
DELETE FROM public.oauth2_state
WHERE state = $1;
