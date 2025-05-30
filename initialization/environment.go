package initialization

import (
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
	"regexp"
)

const projectDirName = "admission"

// InitEnv function
func InitEnv(logger *zap.SugaredLogger) {
	// Load .env file and print variables
	envFile, err := readEnv()
	logger.Debugf("BuildConfig - envFile = %s", envFile)
	if err != nil {
		logger.Error("BuildConfig - failed to load the env file")
		panic("InitEnv - failed to load the env file at ./" + envFile)
	}
	printEnv(logger)
}

func readEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := string(rootPath) + `/.env`
	err := godotenv.Load(envFilePath)
	return envFilePath, err
}

func printEnv(logger *zap.SugaredLogger) {
	logger.Info("ENVIRONMENT = " + os.Getenv("ENV"))
	logger.Info("MONGODB_URL = " + os.Getenv("MONGODB_URL"))
	logger.Info("HTTP_SERVER = " + os.Getenv("HTTP_SERVER"))
	logger.Info("HTTP_PORT = " + os.Getenv("HTTP_PORT"))
	logger.Info("HTTP_CORS = " + os.Getenv("HTTP_CORS"))
	logger.Info("HTTP_SENSOR_SERVER = " + os.Getenv("HTTP_SENSOR_SERVER"))
	logger.Info("HTTP_SENSOR_PORT = " + os.Getenv("HTTP_SENSOR_PORT"))
	logger.Info("HTTP_SENSOR_GETVALUE_API = " + os.Getenv("HTTP_SENSOR_GETVALUE_API"))
	logger.Info("HTTP_SENSOR_REGISTER_API = " + os.Getenv("HTTP_SENSOR_REGISTER_API"))
	logger.Info("HTTP_SENSOR_KEEPALIVE_API = " + os.Getenv("HTTP_SENSOR_KEEPALIVE_API"))
	logger.Info("GRPC_URL = " + os.Getenv("GRPC_URL"))
	logger.Info("GRPC_TLS = " + os.Getenv("GRPC_TLS"))
	logger.Info("CERT_FOLDER_PATH = " + os.Getenv("CERT_FOLDER_PATH"))
	logger.Info("INTERNAL_CLUSTER_PATH = " + os.Getenv("INTERNAL_CLUSTER_PATH"))
}
