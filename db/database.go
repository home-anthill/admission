package db

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
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
	if err = ensureIndexes(ctx, client); err != nil {
		return nil, fmt.Errorf("cannot ensure MongoDB indexes: %w", err)
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

func ensureIndexes(ctx context.Context, client *mongo.Client) error {
	collections := GetCollections(client)
	if _, err := collections.Profiles.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "apiTokenHash", Value: 1}},
			Options: options.Index().SetName("profiles_apiTokenHash_unique").SetUnique(true),
		},
	}); err != nil {
		return err
	}
	if _, err := collections.Devices.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "mac", Value: 1}},
		Options: options.Index().SetName("devices_mac_unique").SetUnique(true),
	}); err != nil {
		return err
	}
	return nil
}
