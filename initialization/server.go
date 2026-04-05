package initialization

import (
	"admission/api"
	"os"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

// SetupRouter creates a Gin engine with compression, payload limits, and optional CORS.
func SetupRouter(logger *zap.SugaredLogger) *gin.Engine {
	port := os.Getenv("HTTP_PORT")
	httpServer := os.Getenv("HTTP_SERVER")

	// 1. init oauthCallbackURL, oauthAppCallbackURL and httpOrigin vars
	httpOrigin := httpServer + ":" + port
	logger.Infow("SetupRouter", "httpOrigin", httpOrigin)

	// 2. init GIN
	router := gin.Default()
	// 3. apply compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 4. fix a max POST payload size
	logger.Info("SetupRouter - set max POST payload size")
	router.Use(limits.RequestSizeLimiter(1024 * 1024))

	// 5. Configure CORS
	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	if os.Getenv("HTTP_CORS") == "true" {
		logger.Warnw("SetupRouter - CORS enabled", "httpOrigin", httpOrigin)
		config := cors.DefaultConfig()
		config.AllowOrigins = []string{
			"http://" + os.Getenv("INTERNAL_CLUSTER_PATH"),
			"http://" + os.Getenv("INTERNAL_CLUSTER_PATH") + ":80",
			"https://" + os.Getenv("INTERNAL_CLUSTER_PATH"),
			"https://" + os.Getenv("INTERNAL_CLUSTER_PATH") + ":443",
			"http://localhost",
			"http://localhost:80",
			"https://localhost",
			"https://localhost:443",
			"http://localhost:8082",
			"http://localhost:3000",
			httpOrigin,
		}
		router.Use(cors.New(config))
	} else {
		logger.Info("SetupRouter - CORS disabled")
	}
	return router
}

// RegisterRoutes sets up the HTTP routes for the admission service.
func RegisterRoutes(router *gin.Engine, logger *zap.SugaredLogger, validate *validator.Validate, client *mongo.Client) {
	register := api.NewRegister(logger, client, validate)

	// public API called by sensors and devices to register themselves
	router.POST("/admission/register", register.PostRegister)
	router.GET("/admission/keepalive", api.GetKeepAlive)
}
