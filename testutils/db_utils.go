package testutils

import (
	"context"

	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func EnsureCollections(ctx context.Context, client *mongo.Client, dbName string) {
	db := client.Database(dbName)
	names, err := db.ListCollectionNames(ctx, map[string]any{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	existing := make(map[string]bool)
	for _, n := range names {
		existing[n] = true
	}
	for _, coll := range []string{"profiles", "devices"} {
		if !existing[coll] {
			err = db.CreateCollection(ctx, coll)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		}
	}
}

func DropAllCollections(ctx context.Context, collProfiles, collDevices *mongo.Collection) {
	err := collProfiles.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = collDevices.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func InsertOne(ctx context.Context, collection *mongo.Collection, obj any) error {
	_, err := collection.InsertOne(ctx, obj)
	return err
}
