package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Livingpool/api"
	"github.com/Livingpool/router"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	dbName   string = "dcard"
	collName string = "advertisements"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// connect to mongodb
	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI("mongodb://tim:secret@localhost:27017/"),
	)

	// close connection to mongodb
	defer func() {
		cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Fatalf("mongodb disconnect error : %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("connection error :%v", err)
		return
	}

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("ping mongodb error :%v", err)
		return
	}
	fmt.Println("ping mongodb success")

	// set up routes
	handler := api.Handler(mongoClient.Database(dbName))
	r := router.Initialize(handler, collName)

	r.Run("localhost:8080")
}
