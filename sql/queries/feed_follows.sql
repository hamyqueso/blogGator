-- name: CreateFeedFollow :many
WITH inserted_feed_follows AS (
  INSERT INTO feed_follows (user_id, feed_id) 
  VALUES($1, $2)
  RETURNING *
)
SELECT
  inserted_feed_follows.*,
  feeds.name AS feed_name,
  users.name AS user_name
FROM inserted_feed_follows
INNER JOIN feeds ON inserted_feed_follows.feed_id = feeds.id
INNER JOIN users ON inserted_feed_follows.user_id = users.id;

-- name: GetFollowingFeeds :many
SELECT feed_id FROM feed_follows
WHERE user_id = $1;

-- name: UnfollowFeed :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;
