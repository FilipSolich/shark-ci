-- name: GetUser :one
SELECT *
FROM public.user
WHERE id = $1;
