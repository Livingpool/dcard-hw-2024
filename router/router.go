package router

import (
	"github.com/Livingpool/api"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func Initialize(mongoColl *mongo.Collection) *gin.Engine {
	router := gin.Default()

	apiGroup := &api.ApiGroup{
		Collection: mongoColl,
	}

	// Set up routes
	router.POST("/api/v1/ad", apiGroup.CreateAd)
	router.GET("/api/v1/ad", apiGroup.SearchAd)
	router.GET("/api/v1/allads", apiGroup.ReturnAllAds)

	return router
}
