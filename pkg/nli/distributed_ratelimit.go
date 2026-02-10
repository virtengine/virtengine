package nli

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/virtengine/virtengine/pkg/ratelimit"
)

// NewDistributedRateLimiter creates a Redis-backed rate limiter and applies memory policy.
func NewDistributedRateLimiter(ctx context.Context, config DistributedRateLimiterConfig, logger zerolog.Logger) (ratelimit.RateLimiter, error) {
	if !config.Enabled {
		return nil, nil
	}

	if config.RedisURL == "" {
		return nil, fmt.Errorf("nli: distributed rate limiter requires redis_url")
	}

	if config.RedisPrefix == "" {
		return nil, fmt.Errorf("nli: distributed rate limiter requires redis_prefix")
	}

	if config.RedisMaxMemoryMB > 0 || config.RedisEvictionPolicy != "" {
		if err := configureRedisPolicy(ctx, config.RedisURL, config.RedisMaxMemoryMB, config.RedisEvictionPolicy); err != nil {
			return nil, err
		}
	}

	rateLimitConfig := ratelimit.RateLimitConfig{
		RedisURL:    config.RedisURL,
		RedisPrefix: config.RedisPrefix,
		Enabled:     true,
		UserLimits: ratelimit.LimitRules{
			RequestsPerSecond: config.RequestsPerSecond,
			RequestsPerMinute: config.RequestsPerMinute,
			BurstSize:         config.BurstSize,
		},
	}

	return ratelimit.NewRedisRateLimiter(ctx, rateLimitConfig, logger)
}

func configureRedisPolicy(ctx context.Context, redisURL string, maxMemoryMB int, policy string) error {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("nli: invalid redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("nli: failed to connect to redis: %w", err)
	}

	return applyRedisMemoryPolicy(ctx, client, maxMemoryMB, policy)
}
