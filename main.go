package main

import (
	"context"
	"crawl-file/connectMongoDb"
	"crawl-file/getDetails"
	"crawl-file/model"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

const (
	urlOrigin    = "https://www.cubdomain.com/domains-registered-dates/1"
	redisQueue   = "update-domain-cubdomain"
	redisAddress = "localhost:6379"
)

func ConnectRedis() error {
	collection := connectMongoDb.Client.Database(connectMongoDb.Config.Database).Collection(connectMongoDb.Config.Collection)
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	err = PushDataInQueue(cur, redisClient)
	if err != nil {
		return err
	}
	return nil
}

func PushDataInQueue(cur *mongo.Cursor, redisClient *redis.Client) error {
	// Push data to Redis queue
	for cur.Next(context.Background()) {

		var result model.Domain
		err := cur.Decode(&result)
		if err != nil {
			return err
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			return err
		}

		err = redisClient.RPush(redisQueue, jsonData).Err()
		if err != nil {
			return err
		}
	}

	if err := cur.Err(); err != nil {
		return err
	}

	fmt.Println("saved all domains in redis")
	return nil
}

func main() {
	err := connectMongoDb.ConnectToMongoDb()
	if err != nil {
		log.Fatal(err)
	}

	//err = advancedCrawl.HandleListDomain(urlOrigin) //	get all domains
	//if err != nil {
	//	log.Fatal(err)
	//}

	err = ConnectRedis() //push in a queue
	if err != nil {
		log.Fatal(err)
	}

	err = getDetails.UploadDomains() //update domain
	if err != nil {
		log.Fatal(err)
	}
}
