package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

func main() {
	// Redis Connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Get length of Redis queue
	length, err := redisClient.LLen("update-domain-cubdomain").Result()
	if err != nil {
		log.Fatal(err)
	}

	// Check if queue is empty
	fmt.Println(length)
}
