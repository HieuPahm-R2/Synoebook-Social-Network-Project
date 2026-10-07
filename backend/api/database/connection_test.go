package database

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestMongoConnectionAndData(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to initialize mongo client: %v", err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	coll := client.Database("social_test").Collection("ping_test")
	defer coll.Drop(ctx)

	res, err := coll.InsertOne(ctx, bson.M{"service": "synoebook", "status": "ok", "time": time.Now()})
	if err != nil {
		t.Fatalf("Failed to insert test document: %v", err)
	}

	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"_id": res.InsertedID}).Decode(&doc)
	if err != nil {
		t.Fatalf("Failed to read test document: %v", err)
	}

	if doc["status"] != "ok" {
		t.Fatalf("Expected status 'ok', got %v", doc["status"])
	}

	t.Logf("MongoDB connection and test data verified successfully: %v", doc)
}
