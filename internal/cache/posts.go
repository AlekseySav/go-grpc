package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"blog/internal/db"

	"github.com/redis/go-redis/v9"
)

const postsTTL = 60 * time.Second

type PostsCache struct {
	rdb *redis.Client
}

func NewPostsCache(rdb *redis.Client) *PostsCache {
	return &PostsCache{rdb: rdb}
}

func postsKey(limit, offset int32) string {
	return fmt.Sprintf("posts:%d:%d", limit, offset)
}

func (c *PostsCache) Get(ctx context.Context, limit, offset int32) ([]db.Post, bool, error) {
	data, err := c.rdb.Get(ctx, postsKey(limit, offset)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var posts []db.Post
	if err := json.Unmarshal(data, &posts); err != nil {
		return nil, false, err
	}
	return posts, true, nil
}

func (c *PostsCache) Set(ctx context.Context, limit, offset int32, posts []db.Post) error {
	data, err := json.Marshal(posts)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, postsKey(limit, offset), data, postsTTL).Err()
}

func (c *PostsCache) Invalidate(ctx context.Context) error {
	keys, err := c.rdb.Keys(ctx, "posts:*").Result()
	if err != nil || len(keys) == 0 {
		return err
	}
	return c.rdb.Del(ctx, keys...).Err()
}