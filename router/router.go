package router

import (
	"github.com/Livingpool/api"
	"github.com/gin-gonic/gin"
)

func Initialize(handler *api.HandlerStruct, collName string) *gin.Engine {
	router := gin.Default()

	err := handler.CreateIndex(collName)
	if err != nil {
		panic(err)
	}

	router.POST("/api/v1/ad", func(c *gin.Context) {
		handler.CreateAd(c, collName)
	})
	router.GET("/api/v1/ad", func(c *gin.Context) {
		handler.SearchAd(c, collName)
	})
	router.GET("/api/v1/allads", func(c *gin.Context) {
		handler.ReturnAllAds(c, collName)
	})
	return router
}
