package initialization

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const projectDirName = "admission"

// InitEnv loads the .env file and logs the configuration. The caller is
// responsible for handling failures.
func InitEnv(logger *zap.SugaredLogger) error {
	envFile, err := readEnv()
	logger.Debugf("BuildConfig - envFile = %s", envFile)
	if err != nil {
		return fmt.Errorf("failed to load env file at ./%s: %w", envFile, err)
	}
	printEnv(logger)
	return nil
}

func readEnv() (string, error) {
	// solution taken from https://stackoverflow.com/a/68347834/3590376
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current working directory: %w", err)
	}
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	envFilePath := filepath.Join(string(rootPath), ".env")
	err = godotenv.Load(envFilePath)
	return envFilePath, err
}

func printEnv(logger *zap.SugaredLogger) {
	logger.Infow("configuration",
		"ENV", os.Getenv("ENV"),
		"LOG_FOLDER", os.Getenv("LOG_FOLDER"),
		"MONGODB_URL", "[REDACTED]",
		"HTTP_SERVER", os.Getenv("HTTP_SERVER"),
		"HTTP_PORT", os.Getenv("HTTP_PORT"),
		"HTTP_CORS", os.Getenv("HTTP_CORS"),
		"HTTP_SENSOR_SERVER", os.Getenv("HTTP_SENSOR_SERVER"),
		"HTTP_SENSOR_PORT", os.Getenv("HTTP_SENSOR_PORT"),
		"HTTP_SENSOR_GETVALUE_API", os.Getenv("HTTP_SENSOR_GETVALUE_API"),
		"HTTP_SENSOR_REGISTER_API", os.Getenv("HTTP_SENSOR_REGISTER_API"),
		"HTTP_SENSOR_KEEPALIVE_API", os.Getenv("HTTP_SENSOR_KEEPALIVE_API"),
		"GRPC_URL", os.Getenv("GRPC_URL"),
		"GRPC_TLS", os.Getenv("GRPC_TLS"),
		"CERT_FOLDER_PATH", os.Getenv("CERT_FOLDER_PATH"),
		"INTERNAL_CLUSTER_PATH", os.Getenv("INTERNAL_CLUSTER_PATH"),
	)
}
