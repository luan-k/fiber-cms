-- name: CreateImage :one
INSERT INTO images (
    name,
    description,
    alt,
    image_path,
    user_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetImage :one
SELECT * FROM images
WHERE id = $1 LIMIT 1;

-- name: GetImagesByUser :many
SELECT * FROM images
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: ListImages :many
SELECT * FROM images
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateImage :one
UPDATE images 
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    alt = COALESCE($4, alt),
    image_path = COALESCE($5, image_path),
    changed_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1;

-- name: DeleteImagesByUserID :exec
DELETE FROM images
WHERE user_id = $1;

-- name: SearchImagesByName :many
SELECT * FROM images
WHERE name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: GetImagesByPost :many
SELECT i.* FROM images i
JOIN post_images pi ON i.id = pi.image_id
WHERE pi.post_id = $1
ORDER BY pi."order", i.created_at;

-- name: CreatePostImage :one
INSERT INTO post_images (
    post_id,
    image_id,
    "order"
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: DeletePostImage :exec
DELETE FROM post_images
WHERE post_id = $1 AND image_id = $2;

-- name: DeletePostImages :exec
DELETE FROM post_images
WHERE post_id = $1;

-- name: DeleteImagePosts :exec
DELETE FROM post_images
WHERE image_id = $1;

-- name: GetImagePostCount :one
SELECT COUNT(*) FROM post_images
WHERE image_id = $1;

-- name: GetPostImageCount :one
SELECT COUNT(*) FROM post_images
WHERE post_id = $1;

-- name: ListImagesWithPostCount :many
SELECT 
    i.*,
    COUNT(pi.post_id) as post_count
FROM images i
LEFT JOIN post_images pi ON i.id = pi.image_id
GROUP BY i.id, i.name, i.description, i.alt, i.image_path, i.user_id, i.created_at, i.changed_at
ORDER BY i.created_at DESC
LIMIT $1
OFFSET $2;

-- name: GetPopularImages :many
SELECT 
    i.*,
    COUNT(pi.post_id) as post_count
FROM images i
JOIN post_images pi ON i.id = pi.image_id
GROUP BY i.id, i.name, i.description, i.alt, i.image_path, i.user_id, i.created_at, i.changed_at
HAVING COUNT(pi.post_id) > 0
ORDER BY COUNT(pi.post_id) DESC
LIMIT $1;

-- name: GetUserImageCount :one
SELECT COUNT(*) FROM images
WHERE user_id = $1;

-- name: TransferImagesToUser :exec
UPDATE images 
SET user_id = $2
WHERE user_id = $1;


-- name: GetPostWithImages :one
SELECT 
    p.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', i.id,
                'name', i.name,
                'description', i.description,
                'alt', i.alt,
                'image_path', i.image_path,
                'user_id', i.user_id,
                'created_at', i.created_at,
                'changed_at', i.changed_at,
                'order', pi."order"
            ) ORDER BY pi."order", i.created_at
        ) FILTER (WHERE i.id IS NOT NULL),
        '[]'::json
    ) as images
FROM posts p
LEFT JOIN post_images pi ON p.id = pi.post_id
LEFT JOIN images i ON pi.image_id = i.id
WHERE p.id = $1
GROUP BY p.id, p.title, p.description, p.content, p.user_id, p.username, p.url, p.created_at, p.changed_at;

-- name: ListPostsWithImages :many
SELECT 
    p.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', i.id,
                'name', i.name,
                'description', i.description,
                'alt', i.alt,
                'image_path', i.image_path,
                'user_id', i.user_id,
                'created_at', i.created_at,
                'changed_at', i.changed_at,
                'order', pi."order"
            ) ORDER BY pi."order", i.created_at
        ) FILTER (WHERE i.id IS NOT NULL),
        '[]'::json
    ) as images
FROM posts p
LEFT JOIN post_images pi ON p.id = pi.post_id
LEFT JOIN images i ON pi.image_id = i.id
GROUP BY p.id, p.title, p.description, p.content, p.user_id, p.username, p.url, p.created_at, p.changed_at
ORDER BY p.created_at DESC
LIMIT $1
OFFSET $2;

-- name: GetPostsByUserWithImages :many
SELECT 
    p.*,
    COALESCE(
        json_agg(
            json_build_object(
                'id', i.id,
                'name', i.name,
                'description', i.description,
                'alt', i.alt,
                'image_path', i.image_path,
                'user_id', i.user_id,
                'created_at', i.created_at,
                'changed_at', i.changed_at,
                'order', pi."order"
            ) ORDER BY pi."order", i.created_at
        ) FILTER (WHERE i.id IS NOT NULL),
        '[]'::json
    ) as images
FROM posts p
LEFT JOIN post_images pi ON p.id = pi.post_id
LEFT JOIN images i ON pi.image_id = i.id
WHERE p.user_id = $1
GROUP BY p.id, p.title, p.description, p.content, p.user_id, p.username, p.url, p.created_at, p.changed_at
ORDER BY p.created_at DESC
LIMIT $2
OFFSET $3;