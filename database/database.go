package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DatabaseClient = ConstructDB()

type DB struct {
	MongoClient *mongo.Client
}

func ConstructDB() *DB {
	db := new(DB)
	db.Connect()
	return db
}

func (db *DB) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to mongodb
	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI("mongodb://admin:pass@db:27017/"),
	)
	if err != nil {
		log.Fatalf("mongodb connection error :%v", err)
	}
	// Ping the database
	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("mongodb ping error :%v", err)
	}
	db.MongoClient = mongoClient
}

func (db *DB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.MongoClient.Disconnect(ctx); err != nil {
		log.Fatalf("mongodb disconnect error : %v", err)
	}

}

func (db *DB) GetCollection(dbName, collName string) *mongo.Collection {
	return db.MongoClient.Database(dbName).Collection(collName)
}
