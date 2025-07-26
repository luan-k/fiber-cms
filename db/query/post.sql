-- name: CreatePosts :one
INSERT INTO posts (
    title,
    description,
    user_id,
    username,
    content,
    url,
    images
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: CreateUserPost :one
INSERT INTO user_posts (
    post_id,
    user_id,
    "order"
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPost :one
SELECT * FROM posts 
WHERE id = $1 LIMIT 1;

-- name: ListPosts :many
SELECT * FROM posts 
ORDER BY id DESC
LIMIT $1
OFFSET $2;

-- name: UpdatePost :one
UPDATE posts
SET title = COALESCE($1, title),
    description = COALESCE($2, description),
    user_id = COALESCE($3, user_id),
    username = COALESCE($4, username),
    content = COALESCE($5, content),
    images = COALESCE($6, images),
    url = COALESCE($7, url),
    changed_at = now()
WHERE id = $8
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1;

-- name: DeleteUserPost :exec
DELETE FROM user_posts
WHERE post_id = $1;