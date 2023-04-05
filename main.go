package main

import (
	"crawl-file/advancedCrawl"
	"crawl-file/connectMongoDb"
	"crawl-file/queue"
	"flag"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"log"
	"runtime"
)

const (
	urlOrigin = "https://www.cubdomain.com/domains-registered-dates/1"
)

var (
	queueMode   bool
	Logger      *zap.Logger
	cur         *mongo.Cursor
	err         error
	redisClient *redis.Client
)

func main() {
	go runQueue()

	err = advancedCrawl.HandleListDomain(urlOrigin) //	get all domains
	if err != nil {
		log.Fatal(err)
	}
}

func runQueue() {
	if !queueMode {
		return
	}
	queues := []func(){
		queue.GenerateDomain,
	}

	for _, worker := range queues {
		go worker()
	}
}

func init() {
	err := connectMongoDb.ConnectToMongoDb()
	if err != nil {
		log.Fatal(err)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.BoolVar(&queueMode, "queue", true, "Enable schedule mode")
}
