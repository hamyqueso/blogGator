-- name: CreatePost :one
INSERT INTO posts (title, url, description, published_at, feed_id)
VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT
  posts.title,
  feeds.name AS blog_name,
  posts.url,
  posts.description,
  COALESCE(posts.published_at, posts.created_at) AS display_time
FROM posts
JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
JOIN feeds ON posts.feed_id = feeds.id
WHERE feed_follows.user_id = $1
ORDER BY 
  COALESCE(posts.published_at, posts.created_at) DESC,
  posts.id DESC
LIMIT $2;
