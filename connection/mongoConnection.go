package connection

import (
	"context"
	"crawl-file/dataConfig"
	"crawl-file/model"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var (
	client *mongo.Client
	err    error
)

const (
	fileEnvPath = "/home/minhtuan/crawl-domain-advanced/data.env"
)

func ReadEnv() (*dataConfig.Config, error) {
	err := godotenv.Load(fileEnvPath)
	if err != nil {
		return nil, err
	}

	return &dataConfig.Config{
		MongoURI:   os.Getenv("MONGO_URI"),
		PassMongo:  os.Getenv("MONGO_PASS"),
		Database:   os.Getenv("MONGO_DATABASE"),
		Collection: os.Getenv("MONGO_COLLECTION"),
		UserMongo:  os.Getenv("MONGO_USER"),
	}, nil
}
func SaveFileToMongoDb(config *dataConfig.Config, doc []mongo.WriteModel) error {

	clientOptions := options.Client().ApplyURI(config.MongoURI)
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {

		}
	}(client, context.Background())
	collection := client.Database(config.Database).Collection(config.Collection)
	bulkWrite, err := collection.BulkWrite(context.Background(), doc)
	if bulkWrite != nil {
		return err
	}
	return nil
}
func UpdateDataMongodb(config *dataConfig.Config, domain model.Domain) error {
	clientOptions := options.Client().ApplyURI(config.MongoURI)
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {

		}
	}(client, context.Background())
	collection := client.Database(config.Database).Collection(config.Collection)

	filter := bson.M{"domain": domain.DomainUrl}
	fmt.Println(filter)
	// Define the update to apply
	update := bson.M{"$set": bson.M{
		"title":       domain.GetTitle(),
		"description": domain.GetDescription(),
		"keywords":    domain.GetKeywords(),
		"owner":       domain.GetOwner(),
		"expires":     domain.GetExpires(),
		"created":     domain.GetCreated(),
	}}

	// Update the first document matching the filter
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("Matched %v documents and modified %v documents.\n", result.MatchedCount, result.ModifiedCount)

	return nil
}
