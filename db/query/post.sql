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