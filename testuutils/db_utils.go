package testuutils

import (
	"context"

	"github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo"
)

func DropAllCollections(ctx context.Context, collProfiles, collDevices *mongo.Collection) {
	var err error
	err = collProfiles.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = collDevices.Drop(ctx)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func InsertOne(ctx context.Context, collection *mongo.Collection, obj interface{}) error {
	_, err := collection.InsertOne(ctx, obj)
	return err
}
