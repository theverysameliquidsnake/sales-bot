package configs

import (
	"context"
	"errors"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var client *mongo.Client

func ConnectToMongo() error {
	conn, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return errors.Join(errors.New("config: cound not connect to mongo db:"), err)
	}
	client = conn

	return nil
}

func DisconnectFromMongo() error {
	if err := client.Disconnect(context.Background()); err != nil {
		return errors.Join(errors.New("config: cound not disconnect from mongo db:"), err)
	}

	return nil
}

func GetMongoDatabase() *mongo.Database {
	return client.Database("sales_bot")
}
