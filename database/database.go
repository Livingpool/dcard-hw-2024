package database

import (
	"context"
	"log"
	"time"

	"github.com/Livingpool/constants"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DatabaseClient = ConstructDB()

type DB struct {
	MongoClient *mongo.Client
	RedisClient *redis.Client
}

func ConstructDB() *DB {
	db := new(DB)
	db.Connect()
	return db
}

// CreateIndex for endAt field (ascending order) to optimize the search
func CreateIndex(ctx context.Context, coll *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "endAt", Value: 1}},
	}

	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load environment variables
	if err := constants.FetchEnvVariables(); err != nil {
		panic(err)
	}

	mongoUsername := constants.MongoUserName
	mongoPassword := constants.MongoPassword

	// Connect to mongodb
	uri := "mongodb://" + mongoUsername + ":" + mongoPassword + "@" + constants.MONGO_IP + ":" + constants.MONGO_PORT + "/"
	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(uri),
	)
	if err != nil {
		log.Fatalf("mongodb connection error :%v", err)
	}
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("mongodb ping error :%v", err)
	}
	db.MongoClient = mongoClient

	// Create index for endAt field
	coll := db.GetCollection(constants.DATABASE_NAME, constants.COLLECTION_NAME)
	if err := CreateIndex(ctx, coll); err != nil {
		log.Fatalf("create index error :%v", err)
	}

	// Connect to redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     constants.REDIS_IP + ":" + constants.REDIS_PORT,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("redis ping error :%v", err)
	}
	db.RedisClient = redisClient
}

func (db *DB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.MongoClient.Disconnect(ctx); err != nil {
		log.Fatalf("mongodb disconnect error : %v", err)
	}

	// if err := db.RedisClient.FlushAll(ctx).Err(); err != nil {
	// 	log.Fatalf("redis flushall error : %v", err)
	// }

	if err := db.RedisClient.Close(); err != nil {
		log.Fatalf("redis disconnect error : %v", err)
	}
}

func (db *DB) GetCollection(dbName, collName string) *mongo.Collection {
	return db.MongoClient.Database(dbName).Collection(collName)
}
