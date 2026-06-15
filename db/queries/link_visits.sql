-- name: CreateLinkVisit :one
INSERT INTO link_visits (
  link_id, ip, user_agent, referer, status
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: CountLinkVisits :one
SELECT count(*) FROM link_visits;

-- name: ListLinkVisits :many
SELECT * FROM link_visits
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;
