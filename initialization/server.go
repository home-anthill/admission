package initialization

import (
	"admission/api"
	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"os"
)

var register *api.Register
var keepAlive *api.KeepAlive

// SetupRouter function
func SetupRouter(logger *zap.SugaredLogger) *gin.Engine {
	port := os.Getenv("HTTP_PORT")
	httpServer := os.Getenv("HTTP_SERVER")

	// 1. init oauthCallbackURL, oauthAppCallbackURL and httpOrigin vars
	httpOrigin := httpServer + ":" + port
	logger.Info("SetupRouter - httpOrigin is = " + httpOrigin)

	// 2. init GIN
	router := gin.Default()
	// 3. apply compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 4. fix a max POST payload size
	logger.Info("SetupRouter - set mac POST payload size")
	router.Use(limits.RequestSizeLimiter(1024 * 1024))

	// 5. Configure CORS
	// - No origin allowed by default
	// - GET,POST, PUT, HEAD methods
	// - Credentials share disabled
	// - Preflight requests cached for 12 hours
	if os.Getenv("HTTP_CORS") == "true" {
		logger.Warn("SetupRouter - CORS enabled and httpOrigin is = " + httpOrigin)
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

// RegisterRoutes function
func RegisterRoutes(ctx context.Context, router *gin.Engine, logger *zap.SugaredLogger, validate *validator.Validate, client *mongo.Client) {
	keepAlive = api.NewKeepAlive(ctx, logger)
	register = api.NewRegister(ctx, logger, client, validate)

	// public API called by sensors and devices to register themselves
	router.POST("/register", register.PostRegister)
	router.GET("/keepalive", keepAlive.GetKeepAlive)
}
