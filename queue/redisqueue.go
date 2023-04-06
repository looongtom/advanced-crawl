package queue

import (
	"context"
	"crawl-file/connectMongoDb"
	"crawl-file/getDetails"
	"crawl-file/model"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"time"
)

const (
	redisQueue   = "update-domain-cubdomain"
	redisAddress = "localhost:6379"
	urlBase      = "https://website.informer.com/"
)

var (
	Logger      *zap.Logger
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})
)

var GenerateDomain = func() {
	go func() {
		for range time.Tick(time.Second * 2) {
			collection := connectMongoDb.Client.Database(connectMongoDb.Config.Database).Collection(connectMongoDb.Config.Collection)
			filter := bson.M{"status": model.StatusDisable}
			cursor, err := collection.Find(context.Background(), filter)

			if err != nil {
				Logger.Error(err.Error())
				return
			}

			for cursor.Next(context.Background()) {
				var item model.Domain
				err := cursor.Decode(&item)
				if err != nil {
					continue
				}

				item.Status = model.StatusEnable

				//_, err = collection.ReplaceOne(context.Background(), bson.M{"_id": item.ID}, item)
				jsonData, err := json.Marshal(item)
				if err != nil {
					panic(err)
				}
				err = redisClient.RPush(redisQueue, jsonData).Err()
				if err != nil {
					continue
				}

				filter := bson.M{"_id": item.ID}
				update := bson.M{"$set": bson.M{"status": model.StatusEnable}}
				_, err = collection.UpdateOne(context.Background(), filter, update)
				if err != nil {
					continue
				}

			}

			if err := cursor.Err(); err != nil {
				Logger.Error(err.Error())
				return
			}

			errCh := make(chan error)
			go func() {
				err := getDetails.UploadDomains()
				errCh <- err
			}()

			er := <-errCh
			if er != nil {
				fmt.Println("Error:", er)
			}

		}
	}()

}
