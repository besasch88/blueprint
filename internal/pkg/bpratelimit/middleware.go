package bpratelimit

import (
	"fmt"

	"github.com/besasch88/blueprint/internal/pkg/bpauth"
	"github.com/besasch88/blueprint/internal/pkg/bprouter"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var iPBasedRateLimit rateLimitInterface
var userBasedRateLimit rateLimitInterface

/*
RateLimitMiddleware is a middleware for APIs based on authenticated or anonymous users

In case of an anonymous user, we leverage its Real-IP
In case of an authenticated user, we leverage its UUID

The idea is to have 2 different rate limits for better workload control under different scenarios.

In particular, for authenticated users, we can leverage custom configurations for specific companies.
This is true when we are going to allow adding more users for one company.

Example of usage of this middleware:
router.GET(

	"/users/:userID",
	tc_middleware.RateLimitMiddleware(),
	... //other middlewares
	func(ctx *gin.Context) {
		... // your logic
*/
func RateLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// First we check if the requester is authenticated, so we leverage the right
		// rate limit.
		authUser := bpauth.GetAuthUserFromSession(ctx)
		var canProceed bool
		var waitTimeSeconds int64
		var key string
		if authUser != nil {
			// We refer to the User ID as unique requester. In this way we can block
			key = fmt.Sprintf("user:%s", authUser.ID.String())
			canProceed, waitTimeSeconds = userBasedRateLimit.canProceed(ctx, key)
		} else {
			// Please check the documentation of ClientIP, from Nginx we can send the
			// Real-IP header that is automatically considered in this scenario
			key = fmt.Sprintf("ip:%s", ctx.ClientIP())
			canProceed, waitTimeSeconds = iPBasedRateLimit.canProceed(ctx, key)
		}
		if !canProceed {
			zap.L().Info(fmt.Sprintf("Rate Limit reached for %s. Abort...", key), zap.String("service", "rate-limit"))
			bprouter.ReturnTooManyRequests(ctx, waitTimeSeconds)
			ctx.Abort()
		} else {
			zap.L().Info("Rate Limit not reached. Proceed...", zap.String("service", "rate-limit"))
			ctx.Next()
		}
	}
}
