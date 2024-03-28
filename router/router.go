package router

import (
	"github.com/Livingpool/api"
	"github.com/Livingpool/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func Initialize(mongoColl *mongo.Collection, redisClient *redis.Client) *gin.Engine {
	router := gin.Default()

	apiGroup := &api.ApiGroup{
		Collection:  mongoColl,
		RedisClient: redisClient,
	}

	router.Use(gin.Recovery())

	// Set up routes
	router.POST("/api/v1/ad", apiGroup.CreateAd)
	router.GET("/api/v1/ad", middleware.CacheByRequestQuery(redisClient), apiGroup.SearchAd)
	router.GET("/api/v1/allads", apiGroup.ReturnAllAds)

	return router
}
