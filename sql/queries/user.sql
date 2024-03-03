-- name: GetUser :one
SELECT id, username, email
FROM public.user
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO public.user (username, email)
VALUES ($1, $2)
RETURNING id;
