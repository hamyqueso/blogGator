-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: ListFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;

-- name: GetFeedByID :one
SELECT * FROM feeds
WHERE id = $1;
