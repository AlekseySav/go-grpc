package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type LikeCache struct {
	rdb *redis.Client
}

func NewLikeCache(rdb *redis.Client) *LikeCache {
	return &LikeCache{rdb: rdb}
}

func key(postID string) string {
	return fmt.Sprintf("likes:%s", postID)
}

func (c *LikeCache) Toggle(ctx context.Context, postID, userID string) (liked bool, err error) {
	added, err := c.rdb.SAdd(ctx, key(postID), userID).Result()
	if err != nil {
		return false, err
	}
	if added == 0 {
		// already liked — remove
		err = c.rdb.SRem(ctx, key(postID), userID).Err()
		return false, err
	}
	return true, nil
}

type LikeStats struct {
	Count  int64
	Liked  bool
}

func (c *LikeCache) GetStats(ctx context.Context, userID string, postIDs []string) (map[string]LikeStats, error) {
	if len(postIDs) == 0 {
		return map[string]LikeStats{}, nil
	}

	pipe := c.rdb.Pipeline()
	cardCmds := make([]*redis.IntCmd, len(postIDs))
	memberCmds := make([]*redis.BoolCmd, len(postIDs))

	for i, id := range postIDs {
		cardCmds[i] = pipe.SCard(ctx, key(id))
		memberCmds[i] = pipe.SIsMember(ctx, key(id), userID)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]LikeStats, len(postIDs))
	for i, id := range postIDs {
		result[id] = LikeStats{
			Count: cardCmds[i].Val(),
			Liked: memberCmds[i].Val(),
		}
	}
	return result, nil
}

func (c *LikeCache) Delete(ctx context.Context, postID string) error {
	return c.rdb.Del(ctx, key(postID)).Err()
}