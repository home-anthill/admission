package initialization

import (
	"admission/db"
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// Start initializes all dependencies and returns the logger, router, and DB client.
func Start() (*zap.SugaredLogger, *gin.Engine, *mongo.Client, error) {
	// 1. Init logger
	logger := InitLogger()

	// 2. Init env
	if err := InitEnv(logger); err != nil {
		return logger, nil, nil, err
	}

	// 3. Init db
	ctx := context.Background()
	client, err := db.InitDb(ctx, logger)
	if err != nil {
		return logger, nil, nil, err
	}

	// 4. Init server
	router := BuildServer(logger, client)

	return logger, router, client, nil
}

// BuildServer - Exposed only for testing purposes
func BuildServer(logger *zap.SugaredLogger, client *mongo.Client) *gin.Engine {
	// Create a singleton validator instance. Validate is designed to be used as a singleton instance.
	// It caches information about struct and validations.
	validate := validator.New()

	// Config Gin framework mode based on env
	setGinMode()

	// Instantiate GIN and apply some middlewares
	logger.Info("BuildServer - GIN - Initializing...")
	router := SetupRouter(logger)
	RegisterRoutes(router, logger, validate, client)
	return router
}

func setGinMode() {
	if os.Getenv("ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else if os.Getenv("ENV") == "testing" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}
