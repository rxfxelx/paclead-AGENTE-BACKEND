//go:build !redis

package clients

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Redis struct{}

func NewRedisFromEnv() *Redis { return &Redis{} }
func (c *Redis) PushBuffer(ctx context.Context, number, message string, ttl time.Duration) error { return nil }
func (c *Redis) GetAllBuffer(ctx context.Context, number string) ([]string, error) { return nil, nil }
func (c *Redis) PopAllBuffer(ctx context.Context, number string) ([]string, error) { return nil, nil }
func (c *Redis) ClearBuffer(ctx context.Context, number string) error { return nil }

// Helper para juntar mensagens
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
