package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// lưu con trỏ giúp tái sử dụng connection pool sẵn có mà không phải sao chép toàn bộ dữ liệu struct trong bộ nhớ.
var Client *mongo.Client
var DB *mongo.Database

func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoUri := os.Getenv("MONGO_URI")
	if mongoUri == "" {
		mongoUri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "social"
	}

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		fmt.Printf("error connect to db: %s\n", err.Error())
		return
	}

	err = Client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Printf("error ping to db: %s\n", err.Error())
		return
	}

	fmt.Println("Connected to MongoDB successfully!")
	DB = Client.Database(dbName)
}
