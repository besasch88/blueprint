package bpratelimit

import "github.com/gin-gonic/gin"

/*
RateLimitInterface represents a generic interface to be implementeed
that determins the rules to proceed in the request or stop it
due to too many requests.
*/
type rateLimitInterface interface {
	canProceed(ctx *gin.Context, key string) (bool, int64)
}
