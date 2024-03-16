package main

import (
	"context"
	"fmt"
	"go-project/api"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	dbName   string = "dcard"
	collName string = "advertisements"
)

func main() {
	router := gin.Default()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	mongoClient, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI("mongodb://tim:secret@localhost:27017/"),
	)

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

	handler := api.Handler(mongoClient)
	router.POST("/api/v1/ad", func(c *gin.Context) {
		handler.CreateAd(c, dbName, collName)
	})
	router.GET("/api/v1/ad", func(c *gin.Context) {
		handler.SearchAd(c, dbName, collName)
	})
	router.GET("/api/v1/allads", func(c *gin.Context) {
		handler.ReturnAllAds(c, dbName, collName)
	})

	router.Run("localhost:8080")
}
