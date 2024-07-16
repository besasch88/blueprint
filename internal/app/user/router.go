package user

import (
	"time"

	"github.com/besasch88/blueprint/internal/pkg/bpauth"
	"github.com/besasch88/blueprint/internal/pkg/bpratelimit"
	"github.com/besasch88/blueprint/internal/pkg/bprouter"
	"github.com/besasch88/blueprint/internal/pkg/bptimeout"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type userRouterInterface interface {
	register(engine *gin.RouterGroup)
}

type userRouter struct {
	service userServiceInterface
}

func newUserRouter(service userServiceInterface) userRouter {
	return userRouter{
		service: service,
	}
}

// Implementation
func (r userRouter) register(router *gin.RouterGroup) {
	router.GET(
		"/users/:userID",
		bpauth.AuthMiddleware([]string{bpauth.UserGet}),
		bptimeout.TimeoutMiddleware(time.Duration(1)*time.Second),
		bpratelimit.RateLimitMiddleware(),
		func(ctx *gin.Context) {
			// Input validation
			var request getUserInputDto
			bprouter.BindParameters(ctx, &request)
			if err := request.validate(); err != nil {
				bprouter.ReturnValidationError(ctx, err)
				return
			}
			// Business Logic
			item, err := r.service.getUserByID(ctx, request)
			if err == errUserNotFound {
				bprouter.ReturnNotFoundError(ctx, err)
				return
			}
			// Errors and output handler
			if err != nil {
				zap.L().Error("Something went wrong", zap.String("service", "user-router"), zap.Error(err))
				bprouter.ReturnGenericError(ctx)
				return
			}
			bprouter.ReturnOk(ctx, &gin.H{"item": item})
		})
}
