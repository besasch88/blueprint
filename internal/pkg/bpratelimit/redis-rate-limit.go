package bpratelimit

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

/*
RedisRateLimitConfiguration represents a rate limit configuration
taking into account the client to store requests and their limits.
*/
type redisRateLimitConfiguration struct {
	RedisClient *redis.Client
	Rate        time.Duration
	Limit       int64
}

/*
RedisRateLimit represents an actual implementation of rate limit based on Redis.
*/
type redisRateLimit struct {
	config redisRateLimitConfiguration
}

/*
NewRedisRateLimit creates a new Rate Limit based on Redis.
*/
func newRedisRateLimit(config redisRateLimitConfiguration) rateLimitInterface {
	return redisRateLimit{
		config: config,
	}
}

/*
CanProceed represents the actual implementation of the method that verify
if the request can proceed or need to be blocked due to many requests.
*/
func (r redisRateLimit) canProceed(ctx *gin.Context, key string) (bool, int64) {
	now := time.Now().Unix()
	redisClient := r.config.RedisClient
	numberOfRequests := redisClient.LLen(ctx, key).Val()
	if numberOfRequests >= r.config.Limit {
		wait := redisClient.TTL(ctx, key).Val().Seconds()
		return false, int64(wait)
	}
	if redisClient.Exists(ctx, key).Val() == 0 {
		redisClient.TxPipelined(ctx, func(p redis.Pipeliner) error {
			p.RPush(ctx, key, now)
			p.Expire(ctx, key, r.config.Rate)
			return nil
		})
	} else {
		redisClient.RPushX(ctx, key, now)
	}
	return true, 0
}
