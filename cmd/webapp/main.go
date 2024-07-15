package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/besasch88/blueprint/internal/app/user"
	"github.com/besasch88/blueprint/internal/pkg/bpdb"
	"github.com/besasch88/blueprint/internal/pkg/bpenv"
	"github.com/besasch88/blueprint/internal/pkg/bpmiddleware"
	"github.com/besasch88/blueprint/internal/pkg/bppubsub"
	"github.com/besasch88/blueprint/internal/pkg/bprouter"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

/*
This is the entrypoint for the Webapp application built on top of
the GIN framework. It exposes a set of APIs.

To start it you can run ´go run ./cmd/webapp/main.go´
*/
func main() {
	// Set default Timezone
	os.Setenv("TZ", "UTC")
	// ENV Variables
	envs := bpenv.ReadEnvs()
	// Set Logger
	logger := zap.Must(zap.NewProduction())
	if envs.AppMode != "release" {
		logger = zap.Must(zap.NewDevelopment())
	}
	zap.ReplaceGlobals(logger)
	// DB Connection
	dbConnection := bpdb.NewDatabaseConnection(
		envs.DbHost,
		envs.DbUsername,
		envs.DbPassword,
		envs.DbName,
		envs.DbPort,
		envs.DbSslMode,
		envs.DbLogSlowQueryThreshold,
		envs.AppMode,
	)
	// PUB-SUB agent
	pubSubAgent := bppubsub.NewPubSubAgent()

	// Start Server
	zap.L().Info("Starting HTTP Server...", zap.String("service", "webapp"))
	gin.SetMode(envs.AppMode)
	r := gin.Default()
	r.SetTrustedProxies(nil)
	// Cors Middleware
	allowOrigins := []string{envs.AppCorsOrigin}
	if envs.AppMode != "release" {
		allowOrigins = append(allowOrigins, bpmiddleware.LocalhostOrigin)
	}
	r.Use(bpmiddleware.CorsMiddleware(allowOrigins))

	r.NoRoute(func(ctx *gin.Context) {
		bprouter.ReturnNotFoundError(ctx, errors.New("endpoint-not-found"))
	})

	// Init moduels that will start exposing endpoints and consumers of internal events
	v1Api := r.Group("api/v1")
	user.Init(envs, dbConnection, pubSubAgent, v1Api)

	// Start the application
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", envs.AppPort),
		Handler: r,
	}

	go func() {
		// Start the HTTP Server and listen for errors
		zap.L().Info(fmt.Sprintf("HTTP Server started on port %d", envs.AppPort), zap.String("service", "webapp"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("Server Startup Error", zap.String("service", "webapp"), zap.Error(err))
			panic(err)
		}
	}()

	/*
		Wait for interrupt Signals to gracefully shutdown the server
		with a timeout of 3 seconds to ensure all the connection are closed
		and all the pubsub chain activities are performed without receiving
		any additional http request
	*/
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	zap.L().Info("Shutdown Server in 3 seconds...", zap.String("service", "webapp"))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	bpdb.CloseDatabaseConnection(dbConnection)
	pubSubAgent.Close()
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("Server Shutdown Error", zap.String("service", "webapp"), zap.Error(err))
	}

	<-ctx.Done()
	zap.L().Info("Server exited!", zap.String("service", "webapp"))
}
