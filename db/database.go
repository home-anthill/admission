package db

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"
)

// Collections holds references to MongoDB collections.
type Collections struct {
	Profiles *mongo.Collection
	Devices  *mongo.Collection
}

// InitDb connects to MongoDB and returns the client. The caller is responsible
// for handling connection failures.
func InitDb(ctx context.Context, logger *zap.SugaredLogger) (*mongo.Client, error) {
	mongoDBUrl := os.Getenv("MONGODB_URL")
	logger.Info("InitDb - connecting to MongoDB")

	client, err := mongo.Connect(options.Client().ApplyURI(mongoDBUrl))
	if err != nil {
		return nil, fmt.Errorf("cannot connect to MongoDB: %w", err)
	}
	if os.Getenv("ENV") != "prod" {
		if err = client.Ping(ctx, readpref.Primary()); err != nil {
			return nil, fmt.Errorf("cannot ping MongoDB: %w", err)
		}
	}
	logger.Info("Connected to MongoDB")

	return client, nil
}

// GetCollections returns handles to the profiles and devices MongoDB collections.
func GetCollections(client *mongo.Client) *Collections {
	db := client.Database(getDbName())
	return &Collections{
		Profiles: db.Collection("profiles"),
		Devices:  db.Collection("devices"),
	}
}

// getDbName returns the database name based on the current environment.
func getDbName() string {
	if os.Getenv("ENV") == "testing" {
		return "api-server-test"
	}
	return "api-server"
}
