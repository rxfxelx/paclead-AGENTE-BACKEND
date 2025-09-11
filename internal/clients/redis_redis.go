//go:build redis

package clients

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
}

func NewRedisFromEnv() *Redis {
	u := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if u == "" {
		return &Redis{rdb: nil}
	}
	opt, err := redis.ParseURL(u)
	if err != nil {
		return &Redis{rdb: nil}
	}
	return &Redis{rdb: redis.NewClient(opt)}
}

func (c *Redis) healthy() bool { return c != nil && c.rdb != nil }
func bufferKey(number string) string { return "NUMBER_buffer_helsenia:" + strings.TrimSpace(number) }

func (c *Redis) PushBuffer(ctx context.Context, number, message string, ttl time.Duration) error {
	if !c.healthy() {
		return nil
	}
	if strings.TrimSpace(number) == "" || strings.TrimSpace(message) == "" {
		return nil
	}
	key := bufferKey(number)
	pipe := c.rdb.TxPipeline()
	pipe.RPush(ctx, key, message)
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *Redis) GetAllBuffer(ctx context.Context, number string) ([]string, error) {
	if !c.healthy() {
		return nil, nil
	}
	key := bufferKey(number)
	return c.rdb.LRange(ctx, key, 0, -1).Result()
}

func (c *Redis) PopAllBuffer(ctx context.Context, number string) ([]string, error) {
	if !c.healthy() {
		return nil, nil
	}
	key := bufferKey(number)
	pipe := c.rdb.TxPipeline()
	get := pipe.LRange(ctx, key, 0, -1)
	pipe.Del(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	return get.Val(), nil
}

func (c *Redis) ClearBuffer(ctx context.Context, number string) error {
	if !c.healthy() {
		return nil
	}
	key := bufferKey(number)
	return c.rdb.Del(ctx, key).Err()
}

func CombineBufferMessage(msgs []string, sep string, maxLen int) (string, error) {
	if len(msgs) == 0 {
		return "", errors.New("empty buffer")
	}
	if sep == "" {
		sep = "\n"
	}
	joined := strings.Join(msgs, sep)
	if maxLen > 0 && len([]rune(joined)) > maxLen {
		rs := []rune(joined)
		joined = string(rs[:maxLen])
	}
	return joined, nil
}
