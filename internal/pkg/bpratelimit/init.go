package bpratelimit

import (
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

/*
Init initialies the Rate limit. In this implementation it leverages the Redis store but it can be quickly updated to use
a different store implementation.
*/
func Init(rlConnectionURI string, rlAnonymousRate int, rlAnonymousMaxRequests int, rlUserRate int, rlUserMaxRequests int) {
	zap.L().Info("Initializing Rate Limit Service on Redis. Connecting...", zap.String("service", "rate-limit"))
	opt, err := redis.ParseURL(rlConnectionURI)
	if err != nil {
		zap.L().Error("Error during Rate Limit Service initalization", zap.String("service", "rate-limit"), zap.Error(err))
		panic(err)
	}
	client := redis.NewClient(opt)

	ipRateLimitStore := newRedisRateLimit(redisRateLimitConfiguration{
		RedisClient: client,
		Rate:        time.Second * time.Duration(rlAnonymousRate),
		Limit:       int64(rlAnonymousMaxRequests),
	})

	userRateLimitStore := newRedisRateLimit(redisRateLimitConfiguration{
		RedisClient: client,
		Rate:        time.Second * time.Duration(rlUserRate),
		Limit:       int64(rlUserMaxRequests),
	})

	iPBasedRateLimit = ipRateLimitStore
	userBasedRateLimit = userRateLimitStore
	zap.L().Info("Rate Limit Service initialized on Redis. Connected!", zap.String("service", "rate-limit"))
}
