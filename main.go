package main

import (
	"context"
	"crawl-file/advancedCrawl"
	"crawl-file/getDetails"
	"crawl-file/model"
	"encoding/json"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const (
	urlOrigin    = "https://www.cubdomain.com/domains-registered-dates/"
	redisQueue   = "update-domain-cubdomain"
	redisAddress = "localhost:6379"
)

var (
	mongoClient *mongo.Client
	err         error
)

func ConnectRedis() {
	clientOptions := options.Client().ApplyURI(advancedCrawl.Config.MongoURI)
	mongoClient, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve data from MongoDB
	collection := mongoClient.Database(advancedCrawl.Config.Database).Collection(advancedCrawl.Config.Collection)
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})
	PushDataInQueue(cur, redisClient)
}
func PushDataInQueue(cur *mongo.Cursor, redisClient *redis.Client) {
	// Push data to Redis queue
	for cur.Next(context.Background()) {
		var result model.Domain
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		jsonData, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}

		err = redisClient.RPush(redisQueue, jsonData).Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Close MongoDB connection
	err = mongoClient.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	advancedCrawl.HandleListDomain(urlOrigin) //	get all domains
	ConnectRedis()                            //push in a queue
	getDetails.UploadDomains()                //update domain
}
