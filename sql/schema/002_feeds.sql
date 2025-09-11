-- +goose UP
CREATE TABLE feeds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(), 
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  url TEXT UNIQUE NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_feed_follows_user ON feed_follows(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_feed ON posts(feed_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_sort ON posts (feed_id, COALESCE(published_at, created_at) DESC, id DESC);

-- +goose DOWN
DROP TABLE feeds;
