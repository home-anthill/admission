package main

import (
	"admission/initialization"
	"context"
	"os"
	"time"
)

func main() {
	logger, router, client, err := initialization.Start()
	if err != nil {
		if logger != nil {
			logger.Fatalw("Failed to initialize application", "error", err)
		}
		panic(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.Errorw("Failed to disconnect MongoDB", "error", err)
		}
	}()

	// Start server
	port := os.Getenv("HTTP_PORT")
	logger.Info("GIN - up and running with port: " + port)
	if err := router.Run(":" + port); err != nil {
		logger.Fatalw("Cannot start HTTP server", "error", err)
	}
}
