package constants

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	DATABASE_NAME   string = "dcard"
	COLLECTION_NAME string = "advertisements"
)

var (
	REDIS_IP       string
	MONGO_IP       string
	REDIS_PORT     string
	MONGO_PORT     string
	CacheKeyFormat string = "offset=%s&limit=%s&age=%s&gender=%s&country=%s&platform=%s"
	MongoUserName  string
	MongoPassword  string
)

func FetchEnvVariables() error {
	var exists bool
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}

	REDIS_IP, exists = os.LookupEnv("REDIS_IP")
	if !exists {
		REDIS_IP = "redis"
	}

	MONGO_IP, exists = os.LookupEnv("MONGO_IP")
	if !exists {
		MONGO_IP = "db"
	}

	REDIS_PORT, exists = os.LookupEnv("REDIS_PORT")
	if !exists {
		REDIS_PORT = "6379"
	}

	MONGO_PORT, exists = os.LookupEnv("MONGO_PORT")
	if !exists {
		MONGO_PORT = "27017"
	}

	MongoUserName, exists = os.LookupEnv("MONGO_USERNAME")
	if !exists {
		log.Fatal("MONGO_USERNAME not found in .env")
	}

	MongoPassword, exists = os.LookupEnv("MONGO_PASSWORD")
	if !exists {
		log.Fatal("MONGO_PASSWORD not found in .env")
	}

	return nil
}
