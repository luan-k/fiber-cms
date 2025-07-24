-- name: CreatePosts :one
INSERT INTO posts (
    name,
    description,
    user_id,
    username,
    content,
    url,
    images
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPost :one
SELECT * FROM posts 
WHERE id = $1 LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts 
ORDER BY id DESC
LIMIT $1
OFFSET $2;