-- name: CreateTaxonomy :one
INSERT INTO taxonomies (
    name,
    description
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetTaxonomy :one
SELECT * FROM taxonomies
WHERE id = $1 LIMIT 1;

-- name: GetTaxonomyByName :one
SELECT * FROM taxonomies
WHERE name = $1 LIMIT 1;

-- name: ListTaxonomies :many
SELECT * FROM taxonomies
ORDER BY name
LIMIT $1
OFFSET $2;

-- name: UpdateTaxonomy :one
UPDATE taxonomies 
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description)
WHERE id = $1
RETURNING *;

-- name: DeleteTaxonomy :exec
DELETE FROM taxonomies
WHERE id = $1;

-- name: CreatePostTaxonomy :one
INSERT INTO posts_taxonomies (
    post_id,
    taxonomy_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetPostTaxonomies :many
SELECT t.* FROM taxonomies t
JOIN posts_taxonomies pt ON t.id = pt.taxonomy_id
WHERE pt.post_id = $1
ORDER BY t.name;

-- name: GetTaxonomyPosts :many
SELECT p.* FROM posts p
JOIN posts_taxonomies pt ON p.id = pt.post_id
WHERE pt.taxonomy_id = $1
ORDER BY p.created_at DESC
LIMIT $2
OFFSET $3;

-- name: DeletePostTaxonomy :exec
DELETE FROM posts_taxonomies
WHERE post_id = $1 AND taxonomy_id = $2;

-- name: DeletePostTaxonomies :exec
DELETE FROM posts_taxonomies
WHERE post_id = $1;

-- name: DeleteTaxonomyPosts :exec
DELETE FROM posts_taxonomies
WHERE taxonomy_id = $1;

-- name: GetPostTaxonomyCount :one
SELECT COUNT(*) FROM posts_taxonomies
WHERE post_id = $1;

-- name: GetTaxonomyPostCount :one
SELECT COUNT(*) FROM posts_taxonomies
WHERE taxonomy_id = $1;

-- name: ListTaxonomiesWithPostCount :many
SELECT 
    t.*,
    COUNT(pt.post_id) as post_count
FROM taxonomies t
LEFT JOIN posts_taxonomies pt ON t.id = pt.taxonomy_id
GROUP BY t.id, t.name, t.description
ORDER BY t.name
LIMIT $1
OFFSET $2;

-- name: GetPopularTaxonomies :many
SELECT 
    t.*,
    COUNT(pt.post_id) as post_count
FROM taxonomies t
JOIN posts_taxonomies pt ON t.id = pt.taxonomy_id
GROUP BY t.id, t.name, t.description
HAVING COUNT(pt.post_id) > 0
ORDER BY COUNT(pt.post_id) DESC
LIMIT $1;

-- name: SearchTaxonomiesByName :many
SELECT * FROM taxonomies
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name
LIMIT $2
OFFSET $3;

-- name: CountTotalTaxonomies :one
SELECT COUNT(*) AS total FROM taxonomies;