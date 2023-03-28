package connectMongoDb

import (
	"context"
	"crawl-file/dataConfig"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"os"
)

var (
	Client     *mongo.Client
	err        error
	Logger     *zap.Logger
	Config     *dataConfig.Config
	Collection *mongo.Collection
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
func ConnectToMongoDb() {
	Config, err = ReadEnv()

	if err != nil {
		Logger.Error(err.Error())
	}

	clientOptions := options.Client().ApplyURI(Config.MongoURI)
	Client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		Logger.Error(err.Error())
	}
	Collection = Client.Database(Config.Database).Collection(Config.Collection)

}
