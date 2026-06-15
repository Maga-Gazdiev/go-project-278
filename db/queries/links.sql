-- name: GetLink :one
SELECT * FROM links
WHERE id = $1 LIMIT 1;

-- name: GetLinks :many
SELECT * FROM links;

-- name: CountLinks :one
SELECT count(*) FROM links;

-- name: ListLinks :many
SELECT * FROM links
ORDER BY original_url
LIMIT $1 OFFSET $2;

-- name: CreateLink :one
INSERT INTO links (
  original_url, short_name, short_url
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateLink :one
UPDATE links
  set original_url = $2,
  short_name = $3,
  short_url = $4,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLink :execrows
DELETE FROM links
WHERE id = $1;
